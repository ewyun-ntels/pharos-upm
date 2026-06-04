import { useQuery } from '@tanstack/react-query';
import { axiosInstance } from '@lib/axios';
import type {
  KubernetesConfig,
  IndexRow,
  PodInfo,
  PodPhase,
  NodeInfo,
  ClusterSummaryData,
  ResourceStatus,
} from '../types';

interface DsQueryResult {
  Frame: {
    data: Record<string, any>[];
    rows: number;
  };
  Error: string;
}

interface DsQueryResponse {
  Results: Record<string, DsQueryResult>;
}

const PROM_REGEX_SPECIAL_CHARS = /[\\^$.*+?()[\]{}|]/g;
const PROM_DURATION_PATTERN = /^[1-9][0-9]*(ms|s|m|h|d|w|y)$/;
const DEFAULT_RESTART_WINDOW = '1h';
const DEFAULT_WARNING_RESTARTS = 1;
const DEFAULT_ERROR_RESTARTS = 3;

export function escapePromLabelValue(value: string): string {
  return value
    .replace(/\\/g, '\\\\')
    .replace(/"/g, '\\"')
    .replace(/\n/g, '\\n');
}

function escapePromRegexValue(value: string): string {
  return escapePromLabelValue(value.replace(PROM_REGEX_SPECIAL_CHARS, '\\$&'));
}

export async function runBatchQuery(
  datasourceName: string,
  queries: Record<string, string>,
): Promise<Record<string, IndexRow[]>> {
  const body = {
    queries: Object.entries(queries).map(([id, sql]) => ({
      id,
      datasourceName,
      sql,
      timeout: 30,
    })),
  };

  const response = await axiosInstance.post<DsQueryResponse>('/plugins/ds/query', body);
  const results: Record<string, DsQueryResult> = response.data?.Results ?? {};

  const parsed: Record<string, IndexRow[]> = {};
  const errors: string[] = [];

  for (const id of Object.keys(queries)) {
    const result = results[id];
    if (!result) {
      parsed[id] = [];
      errors.push(`${id}: missing result`);
      continue;
    }

    if (result.Error) {
      parsed[id] = [];
      errors.push(`${id}: ${result.Error}`);
      continue;
    }

    parsed[id] = (result.Frame?.data ?? []) as IndexRow[];
  }

  if (errors.length > 0) {
    throw new Error(`Datasource query failed (${datasourceName}): ${errors.join('; ')}`);
  }

  return parsed;
}

function buildNsFilter(namespaces: string[]): string {
  if (namespaces.length === 0) return 'namespace!=""';
  if (namespaces.length === 1) return `namespace="${escapePromLabelValue(namespaces[0])}"`;
  return `namespace=~"${namespaces.map(escapePromRegexValue).join('|')}"`;
}

function buildNodeFilter(nodes: string[]): string {
  if (nodes.length === 0) return '';
  if (nodes.length === 1) return `, node="${escapePromLabelValue(nodes[0])}"`;
  return `, node=~"${nodes.map(escapePromRegexValue).join('|')}"`;
}

function normalizePromDuration(value: string): string {
  return PROM_DURATION_PATTERN.test(value) ? value : DEFAULT_RESTART_WINDOW;
}

function normalizePositiveThreshold(value: number, fallback: number): number {
  return Number.isFinite(value) && value > 0 ? value : fallback;
}

function getPodKey(namespace: string, pod: string): string {
  return `${namespace}/${pod}`;
}

function buildQueries(namespaces: string[], cluster: string, nodes: string[], restartWindow: string) {
  const ns = buildNsFilter(namespaces);
  const cl = cluster ? `, cluster="${escapePromLabelValue(cluster)}"` : '';
  const nd = buildNodeFilter(nodes);
  const restartRange = normalizePromDuration(restartWindow);
  const podInfoSelector = `kube_pod_info{${ns}${cl}${nd}}`;
  const filterBySelectedPods = (expr: string) =>
    nodes.length === 0 ? expr : `(${expr}) and on (namespace, pod) ${podInfoSelector}`;

  const cpuUsageByContainer = `max by (cluster, namespace, pod, container) (node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate5m{${ns}${cl}})`;
  const cpuLimitsByContainer = `max by (cluster, namespace, pod, container) (cluster:namespace:pod_cpu:active:kube_pod_container_resource_limits{${ns}${cl}})`;
  const memUsageByContainer = `max by (cluster, namespace, pod, container) (container_memory_working_set_bytes{job="kubelet", metrics_path="/metrics/cadvisor", ${ns}${cl}, container!="", image!=""})`;
  const memLimitsByContainer = `max by (cluster, namespace, pod, container) (cluster:namespace:pod_memory:active:kube_pod_container_resource_limits{${ns}${cl}})`;
  const cpuRequests = `kube_pod_container_resource_requests{job="kube-state-metrics", ${ns}${cl}, resource="cpu"}`;
  const cpuLimits = `kube_pod_container_resource_limits{job="kube-state-metrics", ${ns}${cl}, resource="cpu"}`;
  const memRequests = `kube_pod_container_resource_requests{job="kube-state-metrics", ${ns}${cl}, resource="memory"}`;
  const memLimits = `kube_pod_container_resource_limits{job="kube-state-metrics", ${ns}${cl}, resource="memory"}`;
  const storageSelector = `job="kubelet", ${ns}${cl}`;
  const selectedPvcFilter = nodes.length === 0
    ? ''
    : ` * on(namespace, persistentvolumeclaim) group_left() max by (namespace, persistentvolumeclaim) (` +
      `kube_pod_spec_volumes_persistentvolumeclaims_info{${ns}${cl}} * on(namespace, pod) group_left() ${podInfoSelector}` +
      `)`;
  const clusterStorageUsed = `sum(max by (namespace, persistentvolumeclaim) ` +
    `(kubelet_volume_stats_used_bytes{${storageSelector}})${selectedPvcFilter}) / ` +
    `sum(max by (namespace, persistentvolumeclaim) ` +
    `(kubelet_volume_stats_capacity_bytes{${storageSelector}})${selectedPvcFilter})`;

  return {
    podInfo:       podInfoSelector,
    podPhase:      filterBySelectedPods(`kube_pod_status_phase{${ns}${cl}}`),
    podRestarts:   `sum by (namespace, pod) (${filterBySelectedPods(`kube_pod_container_status_restarts_total{${ns}${cl}}`)})`,
    podRecentRestarts: `sum by (namespace, pod) (${filterBySelectedPods(`increase(kube_pod_container_status_restarts_total{${ns}${cl}}[${restartRange}])`)})`,
    cpuUsage:      `sum by (namespace, pod) (${filterBySelectedPods(cpuUsageByContainer)}) / sum by (namespace, pod) (${filterBySelectedPods(cpuLimitsByContainer)})`,
    memUsage:      `sum by (namespace, pod) (${filterBySelectedPods(memUsageByContainer)}) / sum by (namespace, pod) (${filterBySelectedPods(memLimitsByContainer)})`,
    clusterCpuReq: `sum(${filterBySelectedPods(cpuUsageByContainer)}) / sum(${filterBySelectedPods(cpuRequests)})`,
    clusterCpuLim: `sum(${filterBySelectedPods(cpuUsageByContainer)}) / sum(${filterBySelectedPods(cpuLimits)})`,
    clusterMemReq: `sum(${filterBySelectedPods(memUsageByContainer)}) / sum(${filterBySelectedPods(memRequests)})`,
    clusterMemLim: `sum(${filterBySelectedPods(memUsageByContainer)}) / sum(${filterBySelectedPods(memLimits)})`,
    clusterStorageUsed,
    storageUsage:  `sum by (namespace, pod) (${filterBySelectedPods(`kubelet_volume_stats_used_bytes{job="kubelet", ${ns}${cl}}`)}) / sum by (namespace, pod) (${filterBySelectedPods(`kubelet_volume_stats_capacity_bytes{job="kubelet", ${ns}${cl}}`)})`,
    networkRx:     `sum by (namespace, pod) (${filterBySelectedPods(`rate(container_network_receive_bytes_total{${ns}${cl}}[5m])`)})`,
    networkTx:     `sum by (namespace, pod) (${filterBySelectedPods(`rate(container_network_transmit_bytes_total{${ns}${cl}}[5m])`)})`,
    nodeReady:     `kube_node_status_condition{condition="Ready",status="true"}`,
    nodeRole:      `kube_node_role{}`,
  };
}

function getVal(row: IndexRow): number {
  const raw = row['Value'] ?? row['value'];
  const parsed = parseFloat(String(raw ?? 0));
  return isFinite(parsed) ? parsed : 0;
}

function getStr(row: IndexRow, key: string): string {
  return String(row[key] ?? '');
}

function computePodStatus(
  phase: string,
  recentRestarts: number,
  cpu: number,
  mem: number,
  warningRestarts: number,
  errorRestarts: number,
): ResourceStatus {
  if (phase === 'Failed' || phase === 'Unknown') return 'error';
  if (recentRestarts >= errorRestarts || cpu >= 0.95 || mem >= 0.95) return 'error';
  if (phase === 'Pending' || recentRestarts >= warningRestarts || cpu >= 0.8 || mem >= 0.8) return 'warning';
  return 'normal';
}

function computeNodeStatus(pods: PodInfo[], cpu: number, mem: number, ready: boolean): ResourceStatus {
  if (!ready) return 'error';
  if (pods.some((p) => p.status === 'error') || cpu >= 0.9 || mem >= 0.9) return 'error';
  if (pods.some((p) => p.status === 'warning') || cpu >= 0.7 || mem >= 0.7) return 'warning';
  return 'normal';
}

function processResults(
  data: Record<string, IndexRow[]>,
  warningRestarts: number,
  errorRestarts: number,
): {
  nodes: NodeInfo[];
  clusterSummary: ClusterSummaryData;
} {
  const phaseMap = new Map<string, string>();
  (data.podPhase ?? []).forEach((row) => {
    if (getVal(row) > 0) phaseMap.set(getPodKey(getStr(row, 'namespace'), getStr(row, 'pod')), getStr(row, 'phase'));
  });

  const restartsMap = new Map<string, number>();
  (data.podRestarts ?? []).forEach((row) => restartsMap.set(getPodKey(getStr(row, 'namespace'), getStr(row, 'pod')), Math.round(getVal(row))));

  const recentRestartsMap = new Map<string, number>();
  (data.podRecentRestarts ?? []).forEach((row) => recentRestartsMap.set(getPodKey(getStr(row, 'namespace'), getStr(row, 'pod')), Math.round(getVal(row))));

  const cpuMap = new Map<string, number>();
  (data.cpuUsage ?? []).forEach((row) => cpuMap.set(getPodKey(getStr(row, 'namespace'), getStr(row, 'pod')), getVal(row)));

  const memMap = new Map<string, number>();
  (data.memUsage ?? []).forEach((row) => memMap.set(getPodKey(getStr(row, 'namespace'), getStr(row, 'pod')), getVal(row)));

  const storageMap = new Map<string, number>();
  (data.storageUsage ?? []).forEach((row) => storageMap.set(getPodKey(getStr(row, 'namespace'), getStr(row, 'pod')), getVal(row)));

  const networkRxMap = new Map<string, number>();
  (data.networkRx ?? []).forEach((row) => networkRxMap.set(getPodKey(getStr(row, 'namespace'), getStr(row, 'pod')), getVal(row)));

  const networkTxMap = new Map<string, number>();
  (data.networkTx ?? []).forEach((row) => networkTxMap.set(getPodKey(getStr(row, 'namespace'), getStr(row, 'pod')), getVal(row)));

  const nodeReadyMap = new Map<string, boolean>();
  (data.nodeReady ?? []).forEach((row) => nodeReadyMap.set(getStr(row, 'node'), getVal(row) === 1));

  const nodeRolesMap = new Map<string, string[]>();
  (data.nodeRole ?? []).forEach((row) => {
    const nodeName = getStr(row, 'node');
    const role = getStr(row, 'role');
    if (role) {
      const list = nodeRolesMap.get(nodeName) ?? [];
      list.push(role);
      nodeRolesMap.set(nodeName, list);
    }
  });

  const nodeMap = new Map<string, PodInfo[]>();
  (data.podInfo ?? []).forEach((row) => {
    const name = getStr(row, 'pod');
    const node = getStr(row, 'node') || 'unknown';
    const namespace = getStr(row, 'namespace');
    const key = getPodKey(namespace, name);
    const phase = (phaseMap.get(key) ?? 'Unknown') as PodPhase;
    const restarts = restartsMap.get(key) ?? 0;
    const recentRestarts = recentRestartsMap.get(key) ?? 0;
    const cpuPercent = cpuMap.get(key) ?? 0;
    const memPercent = memMap.get(key) ?? 0;
    const storagePercent = storageMap.get(key) ?? 0;
    const networkRxBps = networkRxMap.get(key) ?? 0;
    const networkTxBps = networkTxMap.get(key) ?? 0;
    const status = computePodStatus(phase, recentRestarts, cpuPercent, memPercent, warningRestarts, errorRestarts);

    const pod: PodInfo = { name, node, namespace, phase, cpuPercent, memPercent, storagePercent, networkRxBps, networkTxBps, restarts, recentRestarts, status };
    const list = nodeMap.get(node) ?? [];
    list.push(pod);
    nodeMap.set(node, list);
  });

  const nodes: NodeInfo[] = [];
  nodeMap.forEach((pods, nodeName) => {
    const cpuPercent = pods.length ? pods.reduce((s, p) => s + p.cpuPercent, 0) / pods.length : 0;
    const memPercent = pods.length ? pods.reduce((s, p) => s + p.memPercent, 0) / pods.length : 0;
    const ready = nodeReadyMap.get(nodeName) ?? true;
    const roles = nodeRolesMap.get(nodeName) ?? [];
    nodes.push({ name: nodeName, cpuPercent, memPercent, pods, status: computeNodeStatus(pods, cpuPercent, memPercent, ready), ready, roles });
  });
  nodes.sort((a, b) => a.name.localeCompare(b.name));

  const storageUsedPercent = getVal((data.clusterStorageUsed ?? [])[0] ?? {});

  const clusterSummary: ClusterSummaryData = {
    cpuRequestsPercent: getVal((data.clusterCpuReq ?? [])[0] ?? {}),
    cpuLimitsPercent:   getVal((data.clusterCpuLim ?? [])[0] ?? {}),
    memRequestsPercent: getVal((data.clusterMemReq ?? [])[0] ?? {}),
    memLimitsPercent:   getVal((data.clusterMemLim ?? [])[0] ?? {}),
    storageUsedPercent,
    storageFreePercent: Math.max(0, 1 - storageUsedPercent),
    nodeCount: nodes.length,
    podCount: (data.podInfo ?? []).length,
  };

  return { nodes, clusterSummary };
}

export function useClusterMeta(datasourceName: string) {
  return useQuery({
    queryKey: ['home-upm-cluster-meta', datasourceName],
    queryFn: async () => {
      const data = await runBatchQuery(datasourceName, {
        namespaces: 'group by (namespace) (kube_namespace_status_phase{phase="Active"})',
        nodes: 'group by (node) (kube_node_info)',
      });
      const namespaces = (data.namespaces ?? [])
        .map(row => String(row['namespace'] ?? ''))
        .filter(Boolean)
        .sort();
      const nodes = (data.nodes ?? [])
        .map(row => String(row['node'] ?? ''))
        .filter(Boolean)
        .sort();
      return { namespaces, nodes };
    },
    staleTime: 5 * 60 * 1000,
  });
}

export function useKubernetesData(config: KubernetesConfig, refetchInterval = 30_000) {
  return useQuery({
    queryKey: ['home-upm-kubernetes', config],
    queryFn: async () => {
      const { datasourceName, namespaces, nodes, cluster, restartWindow } = config;
      const errorRestarts = normalizePositiveThreshold(config.errorRestarts, DEFAULT_ERROR_RESTARTS);
      const warningRestarts = Math.min(
        normalizePositiveThreshold(config.warningRestarts, DEFAULT_WARNING_RESTARTS),
        errorRestarts,
      );
      const queries = buildQueries(namespaces, cluster, nodes, restartWindow);
      const data = await runBatchQuery(datasourceName, queries);
      return processResults(data, warningRestarts, errorRestarts);
    },
    refetchInterval: refetchInterval === 0 ? false : refetchInterval,
    staleTime: 5_000,
  });
}
