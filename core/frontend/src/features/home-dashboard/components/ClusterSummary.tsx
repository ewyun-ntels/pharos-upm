import { Card, CardContent } from '@pharos/shared/components/ui';
import { Skeleton } from '@pharos/shared/components/ui';
import type { ClusterSummaryData } from '../types';

type PodStatusTone = 'normal' | 'warning' | 'error';

interface StatCardProps {
  title: string;
  rows: SummaryRowProps[];
}

interface SummaryRowProps {
  label: string;
  percent: number;
  detail?: string;
  hasValue?: boolean;
  dangerDirection?: 'high' | 'low';
}

interface PodStatusRowProps {
  label: string;
  count: number;
  tone: PodStatusTone;
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

function SummaryRow({ label, percent, detail, hasValue = true, dangerDirection = 'high' }: SummaryRowProps) {
  const pct = hasValue ? Math.round(percent * 100) : 0;
  const color = hasValue ? getStatusColor(pct, dangerDirection) : { bar: 'bg-muted-foreground/30', text: 'text-muted-foreground' };

  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between gap-2 text-xs">
        <span className="text-muted-foreground truncate">{label}</span>
        <span className={`font-semibold tabular-nums ${color.text}`}>{hasValue ? `${pct}%` : '-'}</span>
      </div>
      <div className="h-1.5 w-full rounded-full bg-muted overflow-hidden">
        <div
          className={`h-full rounded-full transition-all ${color.bar}`}
          style={{ width: `${Math.min(Math.max(pct, 0), 100)}%` }}
        />
      </div>
      {detail && (
        <p className="text-xs text-muted-foreground tabular-nums text-right">{detail}</p>
      )}
    </div>
  );
}

function formatCpuCores(cores: number): string {
  const value = Number.isFinite(cores) ? cores : 0;
  if (value <= 0) return '0m';
  if (value < 1) return `${Math.round(value * 1000)}m`;
  const formatted = value.toFixed(value >= 10 ? 0 : 2).replace(/\.?0+$/, '');
  return `${formatted} ${formatted === '1' ? 'core' : 'cores'}`;
}

function formatBytes(bytes: number): string {
  const value = Number.isFinite(bytes) ? bytes : 0;
  if (value < 1024) return `${value.toFixed(0)} B`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
  if (value < 1024 * 1024 * 1024) return `${(value / 1024 / 1024).toFixed(1)} MB`;
  if (value < 1024 ** 4) return `${(value / 1024 / 1024 / 1024).toFixed(1)} GB`;
  return `${(value / 1024 ** 4).toFixed(1)} TB`;
}

function formatUsageTotal(usage: string, total: string, hasTotal: boolean, fallback: string): string {
  return hasTotal ? `${usage} / ${total}` : `${usage} / ${fallback}`;
}

function StatCard({ title, rows }: StatCardProps) {
  return (
    <Card className="min-w-0">
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

const podStatusColor: Record<PodStatusTone, string> = {
  normal: 'bg-green-500',
  warning: 'bg-yellow-400',
  error: 'bg-red-500',
};

function PodStatusRow({ label, count, tone }: PodStatusRowProps) {
  return (
    <div className="flex items-center justify-between gap-3 text-xs">
      <span className="flex items-center gap-2 text-muted-foreground min-w-0">
        <span className={`h-2.5 w-2.5 rounded-full shrink-0 ${podStatusColor[tone]}`} />
        <span className="truncate">{label}</span>
      </span>
      <span className="text-foreground font-semibold tabular-nums">{count}</span>
    </div>
  );
}

function PodStatusCard({ data }: { data: ClusterSummaryData | undefined }) {
  return (
    <Card className="min-w-0">
      <CardContent className="p-4 space-y-3">
        <p className="text-sm font-medium text-foreground truncate">Pod Status</p>
        <div className="space-y-2.5">
          <PodStatusRow label="정상" count={data?.podNormalCount ?? 0} tone="normal" />
          <PodStatusRow label="경고" count={data?.podWarningCount ?? 0} tone="warning" />
          <PodStatusRow label="오류" count={data?.podErrorCount ?? 0} tone="error" />
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
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-3">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-24 rounded-lg" />
        ))}
      </div>
    );
  }

  const cpuUsage = data?.cpuUsageCores ?? 0;
  const cpuRequest = data?.cpuRequestCores ?? 0;
  const cpuLimit = data?.cpuLimitCores ?? 0;
  const memUsage = data?.memUsageBytes ?? 0;
  const memRequest = data?.memRequestBytes ?? 0;
  const memLimit = data?.memLimitBytes ?? 0;
  const storageUsed = data?.storageUsedBytes ?? 0;
  const storageCapacity = data?.storageCapacityBytes ?? 0;
  const storageFree = Math.max(0, storageCapacity - storageUsed);

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
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-3">
        <StatCard
          title="CPU"
          rows={[
            {
              label: 'Usage / Requests',
              percent: data?.cpuRequestsPercent ?? 0,
              detail: formatUsageTotal(formatCpuCores(cpuUsage), formatCpuCores(cpuRequest), cpuRequest > 0, 'No request'),
            },
            {
              label: 'Usage / Limits',
              percent: data?.cpuLimitsPercent ?? 0,
              detail: formatUsageTotal(formatCpuCores(cpuUsage), formatCpuCores(cpuLimit), cpuLimit > 0, 'No limit'),
            },
          ]}
        />
        <StatCard
          title="Memory"
          rows={[
            {
              label: 'Usage / Requests',
              percent: data?.memRequestsPercent ?? 0,
              detail: formatUsageTotal(formatBytes(memUsage), formatBytes(memRequest), memRequest > 0, 'No request'),
            },
            {
              label: 'Usage / Limits',
              percent: data?.memLimitsPercent ?? 0,
              detail: formatUsageTotal(formatBytes(memUsage), formatBytes(memLimit), memLimit > 0, 'No limit'),
            },
          ]}
        />
        <StatCard
          title="Storage"
          rows={[
            {
              label: 'Used',
              percent: data?.storageUsedPercent ?? 0,
              detail: formatUsageTotal(formatBytes(storageUsed), formatBytes(storageCapacity), storageCapacity > 0, 'No capacity'),
              hasValue: storageCapacity > 0,
            },
            {
              label: 'Free',
              percent: data?.storageFreePercent ?? 0,
              detail: formatUsageTotal(formatBytes(storageFree), formatBytes(storageCapacity), storageCapacity > 0, 'No capacity'),
              hasValue: storageCapacity > 0,
              dangerDirection: 'low',
            },
          ]}
        />
        <PodStatusCard data={data} />
      </div>
    </div>
  );
}
