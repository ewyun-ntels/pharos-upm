import { Card, CardContent, CardHeader, CardTitle } from '@pharos/shared/components/ui';
import { Server, Circle, AlertTriangle, XCircle } from 'lucide-react';
import type { NodeInfo, PodInfo, ResourceStatus } from '../types';

function statusIcon(status: ResourceStatus) {
  if (status === 'error') return <XCircle className="w-3.5 h-3.5 text-red-500 shrink-0" />;
  if (status === 'warning') return <AlertTriangle className="w-3.5 h-3.5 text-yellow-400 shrink-0" />;
  return <Circle className="w-3.5 h-3.5 text-green-500 shrink-0 fill-green-500" />;
}

function nodeStatusBorder(status: ResourceStatus) {
  if (status === 'error') return 'border-red-400';
  if (status === 'warning') return 'border-yellow-400';
  return 'border-border';
}

function UsageBar({ percent, status }: { percent: number; status: ResourceStatus }) {
  const pct = Math.round(percent * 100);
  const barColor =
    status === 'error' ? 'bg-red-500' :
    status === 'warning' ? 'bg-yellow-400' :
    'bg-green-500';

  return (
    <div className="flex items-center gap-1.5 min-w-0">
      <div className="flex-1 h-1.5 rounded-full bg-muted overflow-hidden">
        <div className={`h-full rounded-full ${barColor}`} style={{ width: `${Math.min(pct, 100)}%` }} />
      </div>
      <span className="text-xs tabular-nums w-8 text-right">{pct}%</span>
    </div>
  );
}

interface PodItemProps {
  pod: PodInfo;
  onClick: (pod: PodInfo) => void;
}

function PodItem({ pod, onClick }: PodItemProps) {
  const cpuPct = Math.round(pod.cpuPercent * 100);
  const memPct = Math.round(pod.memPercent * 100);

  return (
    <button
      className="w-full flex items-start gap-1.5 rounded px-1.5 py-1 text-left hover:bg-muted/60 transition-colors group"
      onClick={() => onClick(pod)}
    >
      <span className="mt-0.5">{statusIcon(pod.status)}</span>
      <div className="flex-1 min-w-0">
        <p className="text-xs font-medium truncate group-hover:text-primary">{pod.name}</p>
        <p className="text-xs text-muted-foreground">
          cpu {cpuPct}% · mem {memPct}%
          {pod.recentRestarts > 0 && <span className="ml-1 text-yellow-500">↺{pod.recentRestarts}</span>}
        </p>
      </div>
    </button>
  );
}

interface NodeCardProps {
  node: NodeInfo;
  onPodClick: (pod: PodInfo) => void;
}

export function NodeCard({ node, onPodClick }: NodeCardProps) {
  const cpuStatus = node.cpuPercent >= 0.9 ? 'error' : node.cpuPercent >= 0.7 ? 'warning' : 'normal';
  const memStatus = node.memPercent >= 0.9 ? 'error' : node.memPercent >= 0.7 ? 'warning' : 'normal';

  return (
    <Card className={`border ${nodeStatusBorder(node.status)}`}>
      <CardHeader className="py-2.5 px-3">
        <div className="flex items-center gap-2">
          <Server className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
          <CardTitle className="text-sm font-semibold truncate flex-1">{node.name}</CardTitle>
          {statusIcon(node.status)}
        </div>
        <div className="space-y-1 mt-1">
          <div className="flex items-center gap-1.5">
            <span className="text-xs text-muted-foreground w-8">CPU</span>
            <UsageBar percent={node.cpuPercent} status={cpuStatus} />
          </div>
          <div className="flex items-center gap-1.5">
            <span className="text-xs text-muted-foreground w-8">MEM</span>
            <UsageBar percent={node.memPercent} status={memStatus} />
          </div>
        </div>
        {(!node.ready || node.roles.length > 0) && (
          <div className="flex flex-wrap gap-1 mt-1.5">
            {!node.ready && (
              <span className="text-[10px] px-1.5 py-0.5 rounded-sm bg-red-500/15 text-red-400 font-medium border border-red-500/20">
                NotReady
              </span>
            )}
            {node.roles.map((role) => (
              <span
                key={role}
                className="text-[10px] px-1.5 py-0.5 rounded-sm bg-blue-500/15 text-blue-400 font-medium border border-blue-500/20"
              >
                {role}
              </span>
            ))}
          </div>
        )}
      </CardHeader>

      <CardContent className="px-2 pb-2 pt-0">
        <div className="border-t border-border/50 pt-1.5 space-y-0.5 max-h-52 overflow-y-auto">
          {node.pods.length === 0 ? (
            <p className="text-xs text-muted-foreground px-1.5 py-1">No pods</p>
          ) : (
            node.pods.map((pod) => (
              <PodItem key={`${pod.namespace}/${pod.name}`} pod={pod} onClick={onPodClick} />
            ))
          )}
        </div>
      </CardContent>
    </Card>
  );
}
