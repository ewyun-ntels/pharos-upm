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

function escapePromLabelValue(value: string): string {
  return value
    .replace(/\\/g, '\\\\')
    .replace(/"/g, '\\"')
    .replace(/\n/g, '\\n');
}

function escapePromRegexValue(value: string): string {
  return escapePromLabelValue(value.replace(PROM_REGEX_SPECIAL_CHARS, '\\$&'));
}

async function runBatchQuery(
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

function getPodKey(namespace: string, pod: string): string {
  return `${namespace}/${pod}`;
}

function buildQueries(namespaces: string[], cluster: string, nodes: string[]) {
  const ns = buildNsFilter(namespaces);
  const cl = cluster ? `, cluster="${escapePromLabelValue(cluster)}"` : '';
  const nd = buildNodeFilter(nodes);
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

  return {
    podInfo:       podInfoSelector,
    podPhase:      filterBySelectedPods(`kube_pod_status_phase{${ns}${cl}}`),
    podRestarts:   `sum by (namespace, pod) (${filterBySelectedPods(`kube_pod_container_status_restarts_total{${ns}${cl}}`)})`,
    cpuUsage:      `sum by (namespace, pod) (${filterBySelectedPods(cpuUsageByContainer)}) / sum by (namespace, pod) (${filterBySelectedPods(cpuLimitsByContainer)})`,
    memUsage:      `sum by (namespace, pod) (${filterBySelectedPods(memUsageByContainer)}) / sum by (namespace, pod) (${filterBySelectedPods(memLimitsByContainer)})`,
    clusterCpuReq: `sum(${filterBySelectedPods(cpuUsageByContainer)}) / sum(${filterBySelectedPods(cpuRequests)})`,
    clusterCpuLim: `sum(${filterBySelectedPods(cpuUsageByContainer)}) / sum(${filterBySelectedPods(cpuLimits)})`,
    clusterMemReq: `sum(${filterBySelectedPods(memUsageByContainer)}) / sum(${filterBySelectedPods(memRequests)})`,
    clusterMemLim: `sum(${filterBySelectedPods(memUsageByContainer)}) / sum(${filterBySelectedPods(memLimits)})`,
    storageUsage:  `sum by (namespace, pod) (${filterBySelectedPods(`kubelet_volume_stats_used_bytes{job="kubelet", ${ns}${cl}}`)}) / sum by (namespace, pod) (${filterBySelectedPods(`kubelet_volume_stats_capacity_bytes{job="kubelet", ${ns}${cl}}`)})`,
    networkRx:     `sum by (namespace, pod) (${filterBySelectedPods(`rate(container_network_receive_bytes_total{${ns}${cl}}[5m])`)})`,
    networkTx:     `sum by (namespace, pod) (${filterBySelectedPods(`rate(container_network_transmit_bytes_total{${ns}${cl}}[5m])`)})`,
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

function computePodStatus(phase: string, restarts: number, cpu: number, mem: number): ResourceStatus {
  if (phase === 'Failed' || phase === 'Unknown') return 'error';
  if (restarts >= 5 || cpu >= 0.95 || mem >= 0.95) return 'error';
  if (phase === 'Pending' || restarts >= 3 || cpu >= 0.8 || mem >= 0.8) return 'warning';
  return 'normal';
}

function computeNodeStatus(pods: PodInfo[], cpu: number, mem: number): ResourceStatus {
  if (pods.some((p) => p.status === 'error') || cpu >= 0.9 || mem >= 0.9) return 'error';
  if (pods.some((p) => p.status === 'warning') || cpu >= 0.7 || mem >= 0.7) return 'warning';
  return 'normal';
}

function processResults(data: Record<string, IndexRow[]>): {
  nodes: NodeInfo[];
  clusterSummary: ClusterSummaryData;
} {
  const phaseMap = new Map<string, string>();
  (data.podPhase ?? []).forEach((row) => {
    if (getVal(row) > 0) phaseMap.set(getPodKey(getStr(row, 'namespace'), getStr(row, 'pod')), getStr(row, 'phase'));
  });

  const restartsMap = new Map<string, number>();
  (data.podRestarts ?? []).forEach((row) => restartsMap.set(getPodKey(getStr(row, 'namespace'), getStr(row, 'pod')), getVal(row)));

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

  const nodeMap = new Map<string, PodInfo[]>();
  (data.podInfo ?? []).forEach((row) => {
    const name = getStr(row, 'pod');
    const node = getStr(row, 'node') || 'unknown';
    const namespace = getStr(row, 'namespace');
    const key = getPodKey(namespace, name);
    const phase = (phaseMap.get(key) ?? 'Unknown') as PodPhase;
    const restarts = restartsMap.get(key) ?? 0;
    const cpuPercent = cpuMap.get(key) ?? 0;
    const memPercent = memMap.get(key) ?? 0;
    const storagePercent = storageMap.get(key) ?? 0;
    const networkRxBps = networkRxMap.get(key) ?? 0;
    const networkTxBps = networkTxMap.get(key) ?? 0;
    const status = computePodStatus(phase, restarts, cpuPercent, memPercent);

    const pod: PodInfo = { name, node, namespace, phase, cpuPercent, memPercent, storagePercent, networkRxBps, networkTxBps, restarts, status };
    const list = nodeMap.get(node) ?? [];
    list.push(pod);
    nodeMap.set(node, list);
  });

  const nodes: NodeInfo[] = [];
  nodeMap.forEach((pods, nodeName) => {
    const cpuPercent = pods.length ? pods.reduce((s, p) => s + p.cpuPercent, 0) / pods.length : 0;
    const memPercent = pods.length ? pods.reduce((s, p) => s + p.memPercent, 0) / pods.length : 0;
    nodes.push({ name: nodeName, cpuPercent, memPercent, pods, status: computeNodeStatus(pods, cpuPercent, memPercent) });
  });
  nodes.sort((a, b) => a.name.localeCompare(b.name));

  const clusterSummary: ClusterSummaryData = {
    cpuRequestsPercent: getVal((data.clusterCpuReq ?? [])[0] ?? {}),
    cpuLimitsPercent:   getVal((data.clusterCpuLim ?? [])[0] ?? {}),
    memRequestsPercent: getVal((data.clusterMemReq ?? [])[0] ?? {}),
    memLimitsPercent:   getVal((data.clusterMemLim ?? [])[0] ?? {}),
    nodeCount: nodes.length,
    podCount: (data.podInfo ?? []).length,
  };

  return { nodes, clusterSummary };
}

export function useClusterMeta(datasourceName: string) {
  return useQuery({
    queryKey: ['home-v2-cluster-meta', datasourceName],
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
    queryKey: ['home-v2-kubernetes', config],
    queryFn: async () => {
      const { datasourceName, namespaces, nodes, cluster } = config;
      const queries = buildQueries(namespaces, cluster, nodes);
      const data = await runBatchQuery(datasourceName, queries);
      return processResults(data);
    },
    refetchInterval: refetchInterval === 0 ? false : refetchInterval,
    staleTime: 5_000,
  });
}
