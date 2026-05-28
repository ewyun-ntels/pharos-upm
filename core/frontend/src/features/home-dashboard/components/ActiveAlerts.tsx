import { Badge, Button, Skeleton } from '@pharos/shared/components/ui';
import { Bell, ExternalLink, ChevronDown, ChevronUp } from 'lucide-react';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import type { AlarmItem } from '../hooks/use-active-alerts';

const DEFAULT_VISIBLE = 6;

function formatRelativeTime(iso: string): string {
  const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
  if (diff < 60) return `${diff}초 전`;
  if (diff < 3600) return `${Math.floor(diff / 60)}분 전`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}시간 전`;
  return `${Math.floor(diff / 86400)}일 전`;
}

function SeverityDot({ severity }: { severity: string }) {
  const s = severity.toLowerCase();
  const cls =
    s === 'critical' ? 'bg-red-500' :
    s === 'major' ? 'bg-orange-500' :
    s === 'minor' ? 'bg-yellow-400' :
    s === 'warning' ? 'bg-amber-300' :
    'bg-green-500';
  return <span className={`inline-block w-2 h-2 rounded-full shrink-0 ${cls}`} />;
}

function severityBadge(severity: string) {
  const s = severity.toLowerCase();
  if (s === 'critical') return <Badge variant="destructive" className="text-xs font-medium">CRIT</Badge>;
  if (s === 'major') return <Badge className="text-xs font-medium bg-orange-500 hover:bg-orange-500">MAJOR</Badge>;
  if (s === 'minor') return <Badge className="text-xs font-medium bg-yellow-500 hover:bg-yellow-500 text-black">MINOR</Badge>;
  if (s === 'warning') return <Badge className="text-xs font-medium bg-amber-400 hover:bg-amber-400 text-black">WARN</Badge>;
  return <Badge variant="secondary" className="text-xs font-medium">{severity.toUpperCase().slice(0, 4)}</Badge>;
}

function AlertRow({ alert }: { alert: AlarmItem }) {
  const name = alert.event || alert.resource || '-';
  const message = alert.text || '-';
  return (
    <div className="grid grid-cols-[80px_1fr_1fr_100px] items-center gap-3 py-2 px-3 rounded-md hover:bg-muted/50 transition-colors text-sm border-b border-border/40 last:border-0">
      <div>{severityBadge(alert.severity)}</div>
      <span className="truncate font-medium text-xs">{name}</span>
      <span className="truncate text-xs text-muted-foreground">{message}</span>
      <span className="text-xs text-muted-foreground text-right">{formatRelativeTime(alert.lastReceiveTime)}</span>
    </div>
  );
}

interface SummaryBadgeProps {
  count: number;
  label: string;
  colorClass: string;
  severity: string;
}

function SummaryBadge({ count, label, colorClass, severity }: SummaryBadgeProps) {
  if (count === 0) return null;
  return (
    <span className={`flex items-center gap-1 text-sm font-medium ${colorClass}`}>
      <SeverityDot severity={severity} />
      {count} {label}
    </span>
  );
}

interface ActiveAlertsProps {
  alerts: AlarmItem[];
  criticalCount: number;
  majorCount: number;
  minorCount: number;
  isLoading: boolean;
}

export function ActiveAlerts({ alerts, criticalCount, majorCount, minorCount, isLoading }: ActiveAlertsProps) {
  const [expanded, setExpanded] = useState(false);
  const navigate = useNavigate();

  const visibleAlerts = expanded ? alerts : alerts.slice(0, DEFAULT_VISIBLE);
  const hasMore = alerts.length > DEFAULT_VISIBLE;

  return (
    <div className="rounded-lg border border-border bg-card">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-border">
        <div className="flex items-center gap-3">
          <Bell className="w-4 h-4 text-muted-foreground" />
          <span className="font-semibold text-sm">Active Alarms</span>
          {!isLoading && (
            <div className="flex items-center gap-3">
              <SummaryBadge count={criticalCount} label="Critical" severity="critical" colorClass="text-red-500" />
              <SummaryBadge count={majorCount} label="Major" severity="major" colorClass="text-orange-500" />
              <SummaryBadge count={minorCount} label="Minor" severity="minor" colorClass="text-yellow-500" />
              {criticalCount === 0 && majorCount === 0 && minorCount === 0 && (
                <span className="text-sm text-green-600 font-medium">All Clear</span>
              )}
            </div>
          )}
        </div>
        <Button
          variant="ghost"
          size="sm"
          className="h-7 text-xs gap-1"
          onClick={() => navigate('/extensions/alarm/alerta')}
        >
          <ExternalLink className="w-3 h-3" />
          전체 알람
        </Button>
      </div>

      {/* Body */}
      {isLoading ? (
        <div className="p-3 space-y-2">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-8 w-full rounded" />
          ))}
        </div>
      ) : alerts.length === 0 ? (
        <div className="flex items-center justify-center h-16 text-sm text-muted-foreground">
          활성 알람이 없습니다.
        </div>
      ) : (
        <div>
          {/* Table header */}
          <div className="grid grid-cols-[80px_1fr_1fr_100px] gap-3 py-1.5 px-3 text-xs text-muted-foreground font-medium uppercase tracking-wide border-b border-border/40">
            <span>심각도</span>
            <span>알람명</span>
            <span>메시지</span>
            <span className="text-right">발생시각</span>
          </div>

          {/* Rows */}
          <div>
            {visibleAlerts.map((alert) => (
              <AlertRow key={alert.id} alert={alert} />
            ))}
          </div>

          {/* Show more / less */}
          {hasMore && (
            <div className="flex justify-center py-2 border-t border-border/40">
              <Button
                variant="ghost"
                size="sm"
                className="h-7 text-xs gap-1 text-muted-foreground"
                onClick={() => setExpanded((v) => !v)}
              >
                {expanded ? (
                  <>
                    <ChevronUp className="w-3 h-3" />
                    접기
                  </>
                ) : (
                  <>
                    <ChevronDown className="w-3 h-3" />
                    {alerts.length - DEFAULT_VISIBLE}개 더보기
                  </>
                )}
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
