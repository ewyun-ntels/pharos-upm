import { Card, CardContent, CardHeader, CardTitle, Button, Badge } from '@pharos/shared/components/ui';
import { X, ExternalLink, Server, RotateCcw, Circle, AlertTriangle, XCircle } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import type { PodInfo, ResourceStatus } from '../types';
import type { AlarmItem } from '../hooks/use-active-alerts';

function statusIcon(status: ResourceStatus) {
  if (status === 'error') return <XCircle className="w-4 h-4 text-red-500" />;
  if (status === 'warning') return <AlertTriangle className="w-4 h-4 text-yellow-400" />;
  return <Circle className="w-4 h-4 text-green-500 fill-green-500" />;
}

function phaseColor(phase: string) {
  if (phase === 'Running') return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300';
  if (phase === 'Pending') return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-300';
  if (phase === 'Failed') return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300';
  return 'bg-muted text-muted-foreground';
}

function UsageRow({ label, percent }: { label: string; percent: number }) {
  const pct = Math.round(percent * 100);
  const barColor = pct >= 95 ? 'bg-red-500' : pct >= 80 ? 'bg-yellow-400' : 'bg-green-500';

  return (
    <div className="space-y-1">
      <div className="flex justify-between text-xs">
        <span className="text-muted-foreground">{label}</span>
        <span className="font-medium tabular-nums">{pct}%</span>
      </div>
      <div className="h-2 rounded-full bg-muted overflow-hidden">
        <div className={`h-full rounded-full ${barColor}`} style={{ width: `${Math.min(pct, 100)}%` }} />
      </div>
    </div>
  );
}

function severityBadge(severity: string) {
  const s = severity.toLowerCase();
  if (s === 'critical') return <Badge variant="destructive" className="text-xs">{severity}</Badge>;
  if (s === 'major') return <Badge className="text-xs bg-orange-500 hover:bg-orange-500">{severity}</Badge>;
  if (s === 'minor') return <Badge className="text-xs bg-yellow-500 hover:bg-yellow-500">{severity}</Badge>;
  return <Badge variant="secondary" className="text-xs">{severity}</Badge>;
}

function formatBps(bps: number): string {
  if (bps < 1024) return `${bps.toFixed(0)} B/s`;
  if (bps < 1024 * 1024) return `${(bps / 1024).toFixed(1)} KB/s`;
  if (bps < 1024 * 1024 * 1024) return `${(bps / 1024 / 1024).toFixed(1)} MB/s`;
  return `${(bps / 1024 / 1024 / 1024).toFixed(1)} GB/s`;
}

function NetworkRow({ rxBps, txBps }: { rxBps: number; txBps: number }) {
  return (
    <div className="space-y-1">
      <p className="text-xs text-muted-foreground">Network</p>
      <div className="flex gap-3 text-xs font-medium tabular-nums">
        <span className="text-green-600 dark:text-green-400">↓ {formatBps(rxBps)}</span>
        <span className="text-blue-500 dark:text-blue-400">↑ {formatBps(txBps)}</span>
      </div>
    </div>
  );
}

function formatRelativeTime(iso: string): string {
  const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
  if (diff < 60) return `${diff}초 전`;
  if (diff < 3600) return `${Math.floor(diff / 60)}분 전`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}시간 전`;
  return `${Math.floor(diff / 86400)}일 전`;
}

interface PodDetailPanelProps {
  pod: PodInfo;
  relatedAlerts: AlarmItem[];
  onClose: () => void;
}

export function PodDetailPanel({ pod, relatedAlerts, onClose }: PodDetailPanelProps) {
  const navigate = useNavigate();

  return (
    <Card className="border border-border shadow-md">
      <CardHeader className="py-3 px-4 flex flex-row items-center justify-between gap-2">
        <div className="flex items-center gap-2 min-w-0">
          {statusIcon(pod.status)}
          <CardTitle className="text-base font-semibold truncate">{pod.name}</CardTitle>
          <span className={`shrink-0 text-xs px-2 py-0.5 rounded-full font-medium ${phaseColor(pod.phase)}`}>
            {pod.phase}
          </span>
        </div>
        <div className="flex items-center gap-2 shrink-0">
          <Button
            variant="outline"
            size="sm"
            className="h-7 text-xs gap-1"
            onClick={() => navigate('/dashboards')}
          >
            <ExternalLink className="w-3 h-3" />
            대시보드
          </Button>
          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={onClose}>
            <X className="w-4 h-4" />
          </Button>
        </div>
      </CardHeader>

      <CardContent className="px-4 pb-4 pt-0 space-y-4">
        {/* Pod metadata */}
        <div className="flex items-center gap-4 text-xs text-muted-foreground">
          <span className="flex items-center gap-1">
            <Server className="w-3 h-3" />
            {pod.node}
          </span>
          <span className="flex items-center gap-1">
            <RotateCcw className="w-3 h-3" />
            재시작 {pod.restarts}회
          </span>
          <span className="text-muted-foreground/60">ns: {pod.namespace}</span>
        </div>

        {/* Resource usage */}
        <div className="grid grid-cols-2 gap-4">
          <UsageRow label="CPU (Limits 대비)" percent={pod.cpuPercent} />
          <UsageRow label="Memory (Limits 대비)" percent={pod.memPercent} />
          <UsageRow label="Storage (Volume)" percent={pod.storagePercent} />
          <NetworkRow rxBps={pod.networkRxBps} txBps={pod.networkTxBps} />
        </div>

        {/* Related alerts */}
        {relatedAlerts.length > 0 && (
          <div className="space-y-2">
            <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">연관 알람</p>
            <div className="space-y-1.5">
              {relatedAlerts.map((alert) => (
                <div
                  key={alert.id}
                  className="flex items-center gap-2 text-xs rounded-md px-2 py-1.5 bg-muted/50"
                >
                  {severityBadge(alert.severity)}
                  <span className="flex-1 truncate">{alert.text || alert.event || alert.resource}</span>
                  <span className="text-muted-foreground shrink-0">
                    {formatRelativeTime(alert.lastReceiveTime)}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
