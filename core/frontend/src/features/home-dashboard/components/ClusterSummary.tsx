import { Card, CardContent } from '@pharos/shared/components/ui';
import { Skeleton } from '@pharos/shared/components/ui';
import type { ClusterSummaryData } from '../types';

interface StatCardProps {
  title: string;
  rows: SummaryRowProps[];
}

interface SummaryRowProps {
  label: string;
  percent: number;
  dangerDirection?: 'high' | 'low';
}

function getStatusColor(percent: number, dangerDirection: 'high' | 'low' = 'high') {
  const warning = dangerDirection === 'high'
    ? percent >= 80
    : percent <= 20;
  const error = dangerDirection === 'high'
    ? percent >= 95
    : percent <= 5;

  if (error) return { bar: 'bg-red-500', text: 'text-red-600' };
  if (warning) return { bar: 'bg-yellow-400', text: 'text-yellow-600' };
  return { bar: 'bg-green-500', text: 'text-green-600' };
}

function SummaryRow({ label, percent, dangerDirection = 'high' }: SummaryRowProps) {
  const pct = Math.round(percent * 100);
  const color = getStatusColor(pct, dangerDirection);

  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between gap-2 text-xs">
        <span className="text-muted-foreground truncate">{label}</span>
        <span className={`font-semibold tabular-nums ${color.text}`}>{pct}%</span>
      </div>
      <div className="h-1.5 w-full rounded-full bg-muted overflow-hidden">
        <div
          className={`h-full rounded-full transition-all ${color.bar}`}
          style={{ width: `${Math.min(Math.max(pct, 0), 100)}%` }}
        />
      </div>
    </div>
  );
}

function StatCard({ title, rows }: StatCardProps) {
  return (
    <Card className="flex-1 min-w-0">
      <CardContent className="p-4 space-y-3">
        <p className="text-sm font-medium text-foreground truncate">{title}</p>
        <div className="space-y-2.5">
          {rows.map((row) => (
            <SummaryRow key={row.label} {...row} />
          ))}
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
        {Array.from({ length: 3 }).map((_, i) => (
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
        <StatCard
          title="CPU"
          rows={[
            { label: 'Requests', percent: data?.cpuRequestsPercent ?? 0 },
            { label: 'Limits', percent: data?.cpuLimitsPercent ?? 0 },
          ]}
        />
        <StatCard
          title="Memory"
          rows={[
            { label: 'Requests', percent: data?.memRequestsPercent ?? 0 },
            { label: 'Limits', percent: data?.memLimitsPercent ?? 0 },
          ]}
        />
        <StatCard
          title="Storage"
          rows={[
            { label: 'Used', percent: data?.storageUsedPercent ?? 0 },
            { label: 'Free', percent: data?.storageFreePercent ?? 0, dangerDirection: 'low' },
          ]}
        />
      </div>
    </div>
  );
}
