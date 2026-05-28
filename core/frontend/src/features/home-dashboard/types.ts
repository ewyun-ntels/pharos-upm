export type PodPhase = 'Running' | 'Pending' | 'Failed' | 'Unknown' | 'Succeeded';
export type ResourceStatus = 'normal' | 'warning' | 'error';

export interface PodInfo {
  name: string;
  node: string;
  namespace: string;
  phase: PodPhase;
  cpuPercent: number;
  memPercent: number;
  storagePercent: number;
  networkRxBps: number;
  networkTxBps: number;
  restarts: number;
  status: ResourceStatus;
}

export interface NodeInfo {
  name: string;
  cpuPercent: number;
  memPercent: number;
  pods: PodInfo[];
  status: ResourceStatus;
}

export interface ClusterSummaryData {
  cpuRequestsPercent: number;
  cpuLimitsPercent: number;
  memRequestsPercent: number;
  memLimitsPercent: number;
  nodeCount: number;
  podCount: number;
}

export interface KubernetesConfig {
  datasourceName: string;
  namespaces: string[];  // empty = all
  nodes: string[];       // empty = all
  cluster: string;
}

export type IndexRow = Record<string, unknown>;
