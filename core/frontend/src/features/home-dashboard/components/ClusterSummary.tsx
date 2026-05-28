import { Card, CardContent } from '@pharos/shared/components/ui';
import { Skeleton } from '@pharos/shared/components/ui';
import type { ClusterSummaryData } from '../types';

interface StatCardProps {
  label: string;
  percent: number;
}

function StatCard({ label, percent }: StatCardProps) {
  const pct = Math.round(percent * 100);
  const color =
    pct >= 95 ? 'bg-red-500' :
    pct >= 80 ? 'bg-yellow-400' :
    'bg-green-500';
  const textColor =
    pct >= 95 ? 'text-red-600' :
    pct >= 80 ? 'text-yellow-600' :
    'text-green-600';

  return (
    <Card className="flex-1 min-w-0">
      <CardContent className="p-4">
        <p className="text-xs text-muted-foreground truncate mb-2">{label}</p>
        <p className={`text-2xl font-bold ${textColor}`}>{pct}%</p>
        <div className="mt-2 h-1.5 w-full rounded-full bg-muted overflow-hidden">
          <div
            className={`h-full rounded-full transition-all ${color}`}
            style={{ width: `${Math.min(pct, 100)}%` }}
          />
        </div>
      </CardContent>
    </Card>
  );
}

interface ClusterSummaryProps {
  data: ClusterSummaryData | undefined;
  isLoading: boolean;
}

export function ClusterSummary({ data, isLoading }: ClusterSummaryProps) {
  if (isLoading) {
    return (
      <div className="flex gap-3">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="flex-1 h-24 rounded-lg" />
        ))}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center gap-4 text-sm text-muted-foreground px-1">
        {data && (
          <>
            <span>Nodes: <strong className="text-foreground">{data.nodeCount}</strong></span>
            <span>Pods: <strong className="text-foreground">{data.podCount}</strong></span>
          </>
        )}
      </div>
      <div className="flex gap-3">
        <StatCard label="CPU (Requests)" percent={data?.cpuRequestsPercent ?? 0} />
        <StatCard label="CPU (Limits)" percent={data?.cpuLimitsPercent ?? 0} />
        <StatCard label="Memory (Requests)" percent={data?.memRequestsPercent ?? 0} />
        <StatCard label="Memory (Limits)" percent={data?.memLimitsPercent ?? 0} />
      </div>
    </div>
  );
}
