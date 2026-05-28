import { Skeleton } from '@pharos/shared/components/ui';
import { NodeCard } from './NodeCard';
import type { NodeInfo, PodInfo } from '../types';

interface NodeTopologyProps {
  nodes: NodeInfo[];
  isLoading: boolean;
  isError: boolean;
  errorMessage?: string;
  onPodClick: (pod: PodInfo) => void;
}

export function NodeTopology({ nodes, isLoading, isError, errorMessage, onPodClick }: NodeTopologyProps) {
  if (isLoading) {
    return (
      <div className="grid grid-cols-[repeat(auto-fill,minmax(220px,1fr))] gap-3">
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} className="h-64 rounded-lg" />
        ))}
      </div>
    );
  }

  if (isError) {
    return (
      <div className="flex flex-col items-center justify-center h-48 rounded-lg border border-dashed text-muted-foreground text-sm gap-2">
        <span>Prometheus 데이터를 불러올 수 없습니다.</span>
        <span className="text-xs">{errorMessage || 'datasource 이름 또는 PromQL 메트릭 존재 여부를 확인해주세요.'}</span>
      </div>
    );
  }

  if (nodes.length === 0) {
    return (
      <div className="flex items-center justify-center h-48 rounded-lg border border-dashed text-muted-foreground text-sm">
        해당 namespace의 Pod 정보가 없습니다.
      </div>
    );
  }

  return (
    <div className="grid grid-cols-[repeat(auto-fill,minmax(220px,1fr))] gap-3">
      {nodes.map((node) => (
        <NodeCard key={node.name} node={node} onPodClick={onPodClick} />
      ))}
    </div>
  );
}
