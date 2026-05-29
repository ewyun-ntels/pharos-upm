import { useState, useMemo } from 'react';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@pharos/shared/components/ui';
import { RefreshCcw } from 'lucide-react';
import {
  ClusterSummary,
  NodeTopology,
  PodDetailPanel,
  ActiveAlerts,
  MultiSelectFilter,
  useKubernetesData,
  useActiveAlerts,
  useHomeUPMConfig,
  useClusterMeta,
  usePodVolumeDetail,
} from '@pharos/core/features/home-dashboard';
import type { PodInfo, KubernetesConfig, AlarmItem } from '@pharos/core/features/home-dashboard';

const DEFAULT_CLUSTER = '';

const REFRESH_OPTIONS = [
  { label: 'Off', value: 0 },
  { label: '10s', value: 10_000 },
  { label: '30s', value: 30_000 },
  { label: '1m',  value: 60_000 },
  { label: '5m',  value: 300_000 },
];

function findRelatedAlerts(pod: PodInfo, alerts: AlarmItem[]): AlarmItem[] {
  const podLower = pod.name.toLowerCase();
  return alerts.filter(a =>
    a.resource?.toLowerCase().includes(podLower) ||
    a.event?.toLowerCase().includes(podLower) ||
    (a.service ?? []).some(s => s.toLowerCase().includes(podLower))
  );
}

export default function HomeUPMPage() {
  const [selectedPod, setSelectedPod] = useState<PodInfo | null>(null);
  const [refreshInterval, setRefreshInterval] = useState(30_000);
  const [selectedNamespaces, setSelectedNamespaces] = useState<string[]>([]);
  const [selectedNodes, setSelectedNodes] = useState<string[]>([]);

  const { data: homeConfig } = useHomeUPMConfig();
  const datasourceName = homeConfig?.datasource ?? 'prometheus-metric';

  const { data: clusterMeta, isLoading: metaLoading } = useClusterMeta(datasourceName);

  const k8sConfig: KubernetesConfig = useMemo(
    () => ({
      datasourceName,
      namespaces: selectedNamespaces,
      nodes: selectedNodes,
      cluster: DEFAULT_CLUSTER,
    }),
    [datasourceName, selectedNamespaces, selectedNodes],
  );

  const { data: k8sData, isLoading: k8sLoading, isError: k8sError, error: k8sErrorDetail } = useKubernetesData(k8sConfig, refreshInterval);
  const { alerts, criticalCount, majorCount, minorCount, isLoading: alertsLoading } = useActiveAlerts(refreshInterval);
  const { data: volumeDetails, isLoading: isVolumeLoading } = usePodVolumeDetail(selectedPod, datasourceName);

  const relatedAlerts = selectedPod ? findRelatedAlerts(selectedPod, alerts) : [];

  const handleNamespacesChange = (next: string[]) => {
    setSelectedNamespaces(next);
    setSelectedPod(null);
  };

  const handleNodesChange = (next: string[]) => {
    setSelectedNodes(next);
    setSelectedPod(null);
  };

  const currentRefreshLabel = REFRESH_OPTIONS.find(o => o.value === refreshInterval)?.label ?? '30s';

  return (
    <main className="flex flex-col w-full h-full overflow-auto">
      <div className="flex flex-col gap-4 p-4 min-w-0">
        {/* Header */}
        <div className="flex items-center justify-between gap-3 flex-wrap">
          <div className="flex items-center gap-2">
            <h1 className="text-lg font-semibold">Kubernetes Cluster</h1>
            {k8sLoading && (
              <RefreshCcw className="w-4 h-4 text-muted-foreground animate-spin" />
            )}
          </div>

          <div className="flex items-center gap-3 flex-wrap">
            {/* Node multi-select */}
            <div className="flex items-center gap-1.5">
              <span className="text-sm text-muted-foreground">Node</span>
              <MultiSelectFilter
                label="Node"
                options={clusterMeta?.nodes ?? []}
                selected={selectedNodes}
                onChange={handleNodesChange}
                isLoading={metaLoading}
              />
            </div>

            {/* Namespace multi-select */}
            <div className="flex items-center gap-1.5">
              <span className="text-sm text-muted-foreground">Namespace</span>
              <MultiSelectFilter
                label="Namespace"
                options={clusterMeta?.namespaces ?? []}
                selected={selectedNamespaces}
                onChange={handleNamespacesChange}
                isLoading={metaLoading}
              />
            </div>

            {/* Refresh */}
            <div className="flex items-center gap-1.5">
              <span className="text-xs text-muted-foreground">Refresh</span>
              <Select value={String(refreshInterval)} onValueChange={(v) => setRefreshInterval(Number(v))}>
                <SelectTrigger className="h-8 w-20 text-sm">
                  <SelectValue>{currentRefreshLabel}</SelectValue>
                </SelectTrigger>
                <SelectContent>
                  {REFRESH_OPTIONS.map((o) => (
                    <SelectItem key={o.value} value={String(o.value)}>{o.label}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>

        {/* Cluster Summary */}
        <ClusterSummary
          data={k8sData?.clusterSummary}
          isLoading={k8sLoading}
        />

        {/* Node Topology */}
        <div className="space-y-2">
          <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wide px-0.5">
            Node Topology
          </h2>
          <NodeTopology
            nodes={k8sData?.nodes ?? []}
            isLoading={k8sLoading}
            isError={k8sError}
            errorMessage={k8sErrorDetail instanceof Error ? k8sErrorDetail.message : undefined}
            onPodClick={setSelectedPod}
          />
        </div>

        {/* Pod Detail Panel */}
        {selectedPod && (
          <PodDetailPanel
            pod={selectedPod}
            relatedAlerts={relatedAlerts}
            onClose={() => setSelectedPod(null)}
            volumeDetails={volumeDetails}
            isVolumeLoading={isVolumeLoading}
          />
        )}

        {/* Legend */}
        <div className="flex items-center gap-4 text-xs text-muted-foreground px-0.5">
          <span className="flex items-center gap-1.5">
            <span className="w-2 h-2 rounded-full bg-green-500 inline-block" /> 정상
          </span>
          <span className="flex items-center gap-1.5">
            <span className="w-2 h-2 rounded-full bg-yellow-400 inline-block" /> 경고 (CPU/MEM &gt;80% 또는 재시작 3회+)
          </span>
          <span className="flex items-center gap-1.5">
            <span className="w-2 h-2 rounded-full bg-red-500 inline-block" /> 오류 (CrashLoop / Failed / &gt;95%)
          </span>
        </div>

        {/* Active Alarms */}
        <ActiveAlerts
          alerts={alerts}
          criticalCount={criticalCount}
          majorCount={majorCount}
          minorCount={minorCount}
          isLoading={alertsLoading}
        />
      </div>
    </main>
  );
}
