import { Card, CardContent, CardHeader, CardTitle, Button, Badge } from '@pharos/shared/components/ui';
import { X, Server, RotateCcw, Circle, AlertTriangle, XCircle, Trash2 } from 'lucide-react';
import type { PodInfo, ResourceStatus } from '../types';
import type { AlarmItem } from '../hooks/use-active-alerts';
import type { PVCVolumeDetail } from '../hooks/use-pod-volume-detail';

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

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes.toFixed(0)} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
  if (bytes < 1024 ** 4) return `${(bytes / 1024 / 1024 / 1024).toFixed(1)} GB`;
  return `${(bytes / 1024 ** 4).toFixed(1)} TB`;
}

function NetworkRow({ rxBps, txBps }: { rxBps: number; txBps: number }) {
  const total = rxBps + txBps;
  const scale = Math.max(rxBps, txBps) || 1;
  const hasData = rxBps > 0 || txBps > 0;

  return (
    <div className="space-y-1.5">
      <div className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">Network</p>
        <span className="text-xs text-muted-foreground tabular-nums">합계 {formatBps(total)}</span>
      </div>
      {hasData ? (
        <div className="space-y-1">
          <div className="flex items-center gap-2 text-xs">
            <span className="text-cyan-400 w-10 shrink-0">↓ 수신</span>
            <div className="flex-1 h-2 rounded-full bg-muted overflow-hidden">
              <div className="h-full rounded-full bg-cyan-500" style={{ width: `${(rxBps / scale) * 100}%` }} />
            </div>
            <span className="text-cyan-400 font-medium tabular-nums w-20 text-right">{formatBps(rxBps)}</span>
          </div>
          <div className="flex items-center gap-2 text-xs">
            <span className="text-green-400 w-10 shrink-0">↑ 송신</span>
            <div className="flex-1 h-2 rounded-full bg-muted overflow-hidden">
              <div className="h-full rounded-full bg-green-500" style={{ width: `${(txBps / scale) * 100}%` }} />
            </div>
            <span className="text-green-400 font-medium tabular-nums w-20 text-right">{formatBps(txBps)}</span>
          </div>
        </div>
      ) : (
        <p className="text-xs text-muted-foreground">0 B/s</p>
      )}
    </div>
  );
}

function VolumeSection({
  volumeDetails,
  isLoading,
}: {
  volumeDetails?: PVCVolumeDetail[];
  isLoading?: boolean;
}) {
  return (
    <div className="space-y-1.5">
      <p className="text-xs text-muted-foreground">Storage (Volume)</p>
      {isLoading ? (
        <p className="text-xs text-muted-foreground">조회 중...</p>
      ) : !volumeDetails || volumeDetails.length === 0 ? (
        <p className="text-xs text-muted-foreground italic">No Volumes</p>
      ) : (
        <div className="space-y-1.5">
          {volumeDetails.map((v) => {
            const pct = Math.round(v.usedPercent);
            const barColor = pct >= 90 ? 'bg-red-500' : pct >= 80 ? 'bg-yellow-400' : 'bg-green-500';
            return (
              <div key={v.pvcName} className="rounded-md border border-border/50 px-2 py-1.5 space-y-1">
                <p className="text-xs font-medium truncate">{v.pvcName}</p>
                <div className="flex items-center gap-2 text-xs">
                  <div className="flex-1 h-2 rounded-full bg-muted overflow-hidden">
                    <div className={`h-full rounded-full ${barColor}`} style={{ width: `${Math.min(pct, 100)}%` }} />
                  </div>
                  <span className="tabular-nums font-medium w-8 text-right">{pct}%</span>
                  <span className="text-muted-foreground tabular-nums">
                    {formatBytes(v.usedBytes)} / {formatBytes(v.usedBytes + v.freeBytes)}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      )}
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
  onDelete?: (pod: PodInfo) => void;
  canDeletePod?: boolean;
  isDeletingPod?: boolean;
  deleteRequested?: boolean;
  volumeDetails?: PVCVolumeDetail[];
  isVolumeLoading?: boolean;
}

export function PodDetailPanel({
  pod,
  relatedAlerts,
  onClose,
  onDelete,
  canDeletePod = false,
  isDeletingPod = false,
  deleteRequested = false,
  volumeDetails,
  isVolumeLoading,
}: PodDetailPanelProps) {
  return (
    <Card className="border border-border shadow-md">
      <CardHeader className="py-3 px-4 flex flex-row items-center justify-between gap-2">
        <div className="flex items-center gap-2 min-w-0">
          {statusIcon(pod.status)}
          <CardTitle className="text-base font-semibold truncate">{pod.name}</CardTitle>
          <span className={`shrink-0 text-xs px-2 py-0.5 rounded-full font-medium ${phaseColor(pod.phase)}`}>
            {pod.phase}
          </span>
          {deleteRequested && (
            <span className="shrink-0 text-xs px-2 py-0.5 rounded-full font-medium bg-amber-500/15 text-amber-300 border border-amber-500/30">
              Delete requested
            </span>
          )}
        </div>
        <div className="flex items-center gap-2 shrink-0">
          {canDeletePod && onDelete && (
            <Button
              variant="destructive"
              size="sm"
              className="h-7 text-xs gap-1"
              aria-label={isDeletingPod || deleteRequested ? 'Delete requested' : 'Delete pod'}
              disabled={isDeletingPod || deleteRequested}
              onClick={() => onDelete(pod)}
            >
              <Trash2 className="w-3 h-3" />
              {isDeletingPod ? 'Deleting...' : deleteRequested ? 'Delete requested' : 'Delete'}
            </Button>
          )}
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
          <VolumeSection volumeDetails={volumeDetails} isLoading={isVolumeLoading} />
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
