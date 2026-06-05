export type PodPhase = 'Running' | 'Pending' | 'Failed' | 'Unknown' | 'Succeeded';
export type ResourceStatus = 'normal' | 'warning' | 'error';

export interface PodInfo {
  name: string;
  uid: string;
  node: string;
  namespace: string;
  phase: PodPhase;
  cpuPercent: number;
  cpuUsageCores: number;
  cpuLimitCores: number;
  memPercent: number;
  memUsageBytes: number;
  memLimitBytes: number;
  storagePercent: number;
  networkRxBps: number;
  networkTxBps: number;
  restarts: number;
  recentRestarts: number;
  status: ResourceStatus;
}

export interface NodeInfo {
  name: string;
  cpuPercent: number;
  memPercent: number;
  pods: PodInfo[];
  status: ResourceStatus;
  ready: boolean;
  roles: string[];
}

export interface ClusterSummaryData {
  cpuRequestsPercent: number;
  cpuLimitsPercent: number;
  cpuUsageCores: number;
  cpuRequestCores: number;
  cpuLimitCores: number;
  memRequestsPercent: number;
  memLimitsPercent: number;
  memUsageBytes: number;
  memRequestBytes: number;
  memLimitBytes: number;
  storageUsedPercent: number;
  storageFreePercent: number;
  storageUsedBytes: number;
  storageCapacityBytes: number;
  nodeCount: number;
  podCount: number;
  podNormalCount: number;
  podWarningCount: number;
  podErrorCount: number;
}

export interface KubernetesConfig {
  datasourceName: string;
  namespaces: string[];  // empty = all
  nodes: string[];       // empty = all
  cluster: string;
  restartWindow: string;
  warningRestarts: number;
  errorRestarts: number;
}

export type IndexRow = Record<string, unknown>;
