export type PodPhase = 'Running' | 'Pending' | 'Failed' | 'Unknown' | 'Succeeded';
export type ResourceStatus = 'normal' | 'warning' | 'error';

export interface PodInfo {
  name: string;
  uid: string;
  node: string;
  namespace: string;
  phase: PodPhase;
  cpuPercent: number;
  memPercent: number;
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
  memRequestsPercent: number;
  memLimitsPercent: number;
  storageUsedPercent: number;
  storageFreePercent: number;
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
