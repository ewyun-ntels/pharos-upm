import {useState, useCallback, useEffect} from 'react';
import {ColumnDef} from '@tanstack/react-table';
import {axiosInstance} from '@/lib/axios';
import {DataGrid} from '@pharos/shared/components/ui-extension/data-grid';
import {Button, Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@pharos/shared/components/ui';
import {
  SheetTemplate,
  SheetTemplateHeader,
  SheetTemplateContent,
  SheetTemplateGroup,
  SheetTemplateGroupTable,
  SheetTemplateGroupGrid,
  SheetTemplateGroupGridItem,
} from '@pharos/shared/components/template/sheet';
import {useSheetStore} from '@components/sheet/sheetStore';
import {RefreshCw, ExternalLink, AlertTriangle, X, Download} from 'lucide-react';
import {IconButton} from '@pharos/shared/components/ui-extension';

const ALERTA_FALLBACK_URL = 'http://10.255.254.22:30880';

const REFRESH_OPTIONS = [
  {label: 'Off',  value: 0},
  {label: '5s',   value: 5_000},
  {label: '10s',  value: 10_000},
  {label: '15s',  value: 15_000},
  {label: '30s',  value: 30_000},
  {label: '1m',   value: 60_000},
  {label: '5m',   value: 300_000},
];

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

type Severity =
  | 'critical'
  | 'major'
  | 'minor'
  | 'warning'
  | 'informational'
  | 'normal'
  | 'ok'
  | 'debug'
  | 'trace'
  | 'unknown';

type AlertStatus =
  | 'open'
  | 'ack'
  | 'assign'
  | 'shelved'
  | 'blackout'
  | 'closed'
  | 'expired'
  | 'unknown';

interface AlertaAlert {
  id: string;
  resource: string;
  event: string;
  environment: string;
  severity: Severity;
  correlate: string[];
  status: AlertStatus;
  service: string[];
  group: string;
  value: string;
  text: string;
  tags: string[];
  attributes: Record<string, unknown>;
  origin: string;
  type: string;
  createTime: string;
  timeout: number;
  rawData: string;
  duplicateCount: number;
  repeat: boolean;
  previousSeverity: string;
  trendIndication: string;
  receiveTime: string;
  lastReceiveId: string;
  lastReceiveTime: string;
  count: number;
}

interface AlertaResponse {
  alerts: AlertaAlert[];
  total: number;
  page: number;
  pageSize: number;
  pages: number;
  more: boolean;
  status: string;
}

interface AlertHistory {
  id: string;
  event: string;
  severity: string;
  status?: string;
  text: string;
  value: string;
  type: string;
  updateTime: string;
  user?: string;
}

interface AlertDetailResponse {
  alert: AlertaAlert & {history: AlertHistory[]};
}

interface AlarmConfig {
  alertaURL: string;
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const SEVERITY_STYLES: Record<string, string> = {
  critical:      'bg-red-600 text-white',
  major:         'bg-orange-500 text-white',
  minor:         'bg-yellow-400 text-black',
  warning:       'bg-amber-300 text-black',
  informational: 'bg-blue-500 text-white',
  normal:        'bg-green-500 text-white',
  ok:            'bg-green-500 text-white',
  debug:         'bg-slate-400 text-white',
  trace:         'bg-slate-400 text-white',
  unknown:       'bg-gray-400 text-white',
};

const STATUS_STYLES: Record<string, string> = {
  open:     'bg-red-100 text-red-800 border-red-300',
  ack:      'bg-yellow-100 text-yellow-800 border-yellow-300',
  assign:   'bg-blue-100 text-blue-800 border-blue-300',
  shelved:  'bg-slate-100 text-slate-700 border-slate-300',
  blackout: 'bg-gray-900 text-white border-gray-700',
  closed:   'bg-green-100 text-green-800 border-green-300',
  expired:  'bg-orange-100 text-orange-800 border-orange-300',
  unknown:  'bg-gray-100 text-gray-700 border-gray-300',
};

const SEVERITY_ORDER: Severity[] = [
  'critical', 'major', 'minor', 'warning', 'informational', 'normal', 'ok', 'debug', 'trace', 'unknown',
];

// Status groups shown as filter tabs
const STATUS_FILTERS = [
  {label: 'Active',  value: 'active',  statuses: ['open', 'ack', 'assign'] as AlertStatus[]},
  {label: 'Open',    value: 'open',    statuses: ['open'] as AlertStatus[]},
  {label: 'Closed',  value: 'closed',  statuses: ['closed', 'expired'] as AlertStatus[]},
  {label: 'All',     value: 'all',     statuses: [] as AlertStatus[]},
];

// ---------------------------------------------------------------------------
// Helper components
// ---------------------------------------------------------------------------

function SeverityBadge({severity}: {severity: string}) {
  const cls = SEVERITY_STYLES[severity.toLowerCase()] ?? SEVERITY_STYLES['unknown'];
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold uppercase tracking-wide ${cls}`}>
      {severity}
    </span>
  );
}

function StatusBadge({status}: {status: string}) {
  const cls = STATUS_STYLES[status.toLowerCase()] ?? STATUS_STYLES['unknown'];
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${cls}`}>
      {status}
    </span>
  );
}

function SeverityCountCard({severity, count}: {severity: string; count: number}) {
  if (count === 0) return null;
  const cls = SEVERITY_STYLES[severity.toLowerCase()] ?? SEVERITY_STYLES['unknown'];
  return (
    <div className={`flex flex-col items-center justify-center rounded-md px-4 py-2 min-w-[72px] ${cls}`}>
      <span className="text-xl font-bold leading-none">{count}</span>
      <span className="text-xs mt-1 capitalize opacity-90">{severity}</span>
    </div>
  );
}

function formatTime(iso: string): string {
  if (!iso) return '';
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

// ---------------------------------------------------------------------------
// Export
// ---------------------------------------------------------------------------

function exportAlertsToCSV(data: AlertaAlert[]) {
  const headers = ['Alert', 'Item', 'Code', 'Severity', 'Status', 'Resource', 'Event', 'Message', 'Last Received'];
  const rows = data.map(a => [
    a.environment ?? '',
    (a.service ?? []).join(', '),
    String(a.attributes?.errCode ?? a.attributes?.errorCode ?? a.attributes?.err_code ?? ''),
    a.severity ?? '',
    a.status ?? '',
    a.resource ?? '',
    a.event ?? '',
    (a.text ?? '').replace(/[\r\n]+/g, ' '),
    formatTime(a.lastReceiveTime),
  ]);

  const csv = [headers, ...rows]
    .map(row => row.map(cell => `"${String(cell).replace(/"/g, '""')}"`).join(','))
    .join('\r\n');

  const bom = '﻿';
  const blob = new Blob([bom + csv], {type: 'text/csv;charset=utf-8;'});
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = `alarms_${new Date().toISOString().slice(0, 10)}.csv`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

// ---------------------------------------------------------------------------
// History Sheet
// ---------------------------------------------------------------------------

const historyColumns: ColumnDef<AlertHistory>[] = [
  {
    accessorKey: 'updateTime',
    header: 'Time',
    size: 160,
    cell: ({getValue}) => (
      <span className="text-xs whitespace-nowrap">{formatTime(getValue<string>())}</span>
    ),
  },
  {
    accessorKey: 'severity',
    header: 'Severity',
    size: 100,
    cell: ({getValue}) => {
      const v = getValue<string>();
      return v ? <SeverityBadge severity={v} /> : null;
    },
  },
  {
    accessorKey: 'status',
    header: 'Status',
    size: 90,
    cell: ({getValue}) => {
      const v = getValue<string>();
      return v ? <StatusBadge status={v} /> : null;
    },
  },
  {
    accessorKey: 'text',
    header: 'Message',
    cell: ({getValue}) => (
      <span className="text-xs text-muted-foreground">{getValue<string>()}</span>
    ),
  },
];

function AlertHistorySheet({alert}: {alert: AlertaAlert}) {
  const {setClose} = useSheetStore();
  const [history, setHistory] = useState<AlertHistory[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    axiosInstance
      .get<AlertDetailResponse>(`/alarm/alerts/${alert.id}`)
      .then(resp => {
        const raw = [...(resp.data.alert.history ?? [])];
        raw.sort((a, b) => new Date(b.updateTime).getTime() - new Date(a.updateTime).getTime());
        setHistory(raw);
      })
      .catch(() => setError('Failed to load history'))
      .finally(() => setIsLoading(false));
  }, [alert.id]);

  return (
    <SheetTemplate>
      <SheetTemplateHeader
        title={`${alert.resource} · ${alert.event}`}
        leftInfoGroup={<SeverityBadge severity={alert.severity} />}
        rightButtonGroup={
          <IconButton icon={<X className="h-4 w-4" />} variant="ghost" className="h-7 w-7 text-muted-foreground" onClick={setClose}>
            Close
          </IconButton>
        }
      />
      <SheetTemplateContent>
        <SheetTemplateGroup title="기본 정보">
          <SheetTemplateGroupGrid cols={2}>
            <SheetTemplateGroupGridItem cols={2} label="Resource" value={alert.resource} />
            <SheetTemplateGroupGridItem cols={2} label="Event" value={alert.event} />
            <SheetTemplateGroupGridItem cols={2} label="Alert" value={alert.environment} />
            <SheetTemplateGroupGridItem cols={2} label="Item" value={(alert.service ?? []).join(', ')} />
            <SheetTemplateGroupGridItem cols={2} label="Status" value={alert.status} />
            <SheetTemplateGroupGridItem cols={2} label="Last Received" value={formatTime(alert.lastReceiveTime)} />
          </SheetTemplateGroupGrid>
        </SheetTemplateGroup>

        <SheetTemplateGroup title="History">
          <SheetTemplateGroupTable>
            {error ? (
              <p className="text-xs text-destructive p-2">{error}</p>
            ) : isLoading ? (
              <p className="text-xs text-muted-foreground p-2">Loading...</p>
            ) : (
              <DataGrid
                columns={historyColumns}
                data={history}
                enablePagination={false}
                tableHeight="auto"
              />
            )}
          </SheetTemplateGroupTable>
        </SheetTemplateGroup>
      </SheetTemplateContent>
    </SheetTemplate>
  );
}

// ---------------------------------------------------------------------------
// Column definitions
// ---------------------------------------------------------------------------

const columns: ColumnDef<AlertaAlert>[] = [
  {
    accessorKey: 'environment',
    header: 'Alert',
    size: 100,
  },
  {
    accessorKey: 'service',
    header: 'Item',
    cell: ({getValue}) => (getValue<string[]>() ?? []).join(', '),
    size: 140,
  },
  {
    id: 'errCode',
    header: 'Code',
    size: 100,
    cell: ({row}) => {
      const attrs = row.original.attributes ?? {};
      const code = attrs['errCode'] ?? attrs['errorCode'] ?? attrs['err_code'] ?? '';
      return <span className="font-mono text-xs">{String(code)}</span>;
    },
  },
  {
    accessorKey: 'severity',
    header: 'Severity',
    cell: ({getValue}) => <SeverityBadge severity={getValue<string>()} />,
    size: 110,
  },
  {
    accessorKey: 'status',
    header: 'Status',
    cell: ({getValue}) => <StatusBadge status={getValue<string>()} />,
    size: 100,
  },
  {
    accessorKey: 'resource',
    header: 'Resource',
    size: 160,
  },
  {
    accessorKey: 'event',
    header: 'Event',
    size: 160,
  },
  {
    accessorKey: 'text',
    header: 'Message',
    size: 280,
    cell: ({getValue}) => (
      <span className="line-clamp-2 text-xs text-muted-foreground">{getValue<string>()}</span>
    ),
  },
  {
    accessorKey: 'lastReceiveTime',
    header: 'Last Received',
    cell: ({getValue}) => (
      <span className="text-xs text-muted-foreground whitespace-nowrap">{formatTime(getValue<string>())}</span>
    ),
    size: 160,
  },
];

// ---------------------------------------------------------------------------
// Main page component
// ---------------------------------------------------------------------------

export function AlertaDashboard() {
  const [alerts, setAlerts] = useState<AlertaAlert[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [alertaURL, setAlertaURL] = useState<string>('');
  const [activeFilter, setActiveFilter] = useState('active');
  const [lastRefreshed, setLastRefreshed] = useState<Date | null>(null);
  const [refreshInterval, setRefreshInterval] = useState(30_000);
  const {setSheet} = useSheetStore();

  const fetchConfig = useCallback(async () => {
    try {
      const resp = await axiosInstance.get<AlarmConfig>('/alarm/config');
      setAlertaURL(resp.data.alertaURL ?? ALERTA_FALLBACK_URL);
    } catch {
      setAlertaURL(ALERTA_FALLBACK_URL);
    }
  }, []);

  const fetchAlerts = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const resp = await axiosInstance.get<AlertaResponse>('/alarm/alerts');
      setAlerts(resp.data.alerts ?? []);
      setLastRefreshed(new Date());
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to fetch alerts';
      setError(message);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchConfig();
    fetchAlerts();
  }, [fetchConfig, fetchAlerts]);

  // 자동 새로고침 (Off 이면 비활성)
  useEffect(() => {
    if (refreshInterval === 0) return;
    const timer = setInterval(fetchAlerts, refreshInterval);
    return () => clearInterval(timer);
  }, [fetchAlerts, refreshInterval]);

  // Compute severity counts (across ALL alerts, not just filtered)
  const severityCounts = SEVERITY_ORDER.reduce<Record<string, number>>((acc, sev) => {
    acc[sev] = alerts.filter(a => a.severity.toLowerCase() === sev).length;
    return acc;
  }, {});

  // Apply status filter
  const currentFilter = STATUS_FILTERS.find(f => f.value === activeFilter) ?? STATUS_FILTERS[0];
  const filteredAlerts =
    currentFilter.statuses.length === 0
      ? alerts
      : alerts.filter(a => currentFilter.statuses.includes(a.status as AlertStatus));

  // Sort: critical first, then by lastReceiveTime desc
  const sortedAlerts = [...filteredAlerts].sort((a, b) => {
    const sevA = SEVERITY_ORDER.indexOf(a.severity.toLowerCase() as Severity);
    const sevB = SEVERITY_ORDER.indexOf(b.severity.toLowerCase() as Severity);
    if (sevA !== sevB) return sevA - sevB;
    return new Date(b.lastReceiveTime).getTime() - new Date(a.lastReceiveTime).getTime();
  });

  return (
    <main className="flex flex-col w-full h-full gap-0">
      {/* ---- Header ---- */}
      <div className="flex items-center justify-between px-6 py-3 border-b bg-background shrink-0">
        <div>
          <h1 className="text-lg font-semibold">Alarm Dashboard</h1>
          {lastRefreshed && (
            <p className="text-xs text-muted-foreground mt-0.5">
              Last refreshed: {lastRefreshed.toLocaleTimeString()}
              {refreshInterval > 0 && ` · auto-refresh ${REFRESH_OPTIONS.find(o => o.value === refreshInterval)?.label}`}
            </p>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Select
            value={String(refreshInterval)}
            onValueChange={(v: string) => setRefreshInterval(Number(v))}
          >
            <SelectTrigger className="h-8 w-24 text-xs">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {REFRESH_OPTIONS.map(o => (
                <SelectItem key={o.value} value={String(o.value)}>
                  {o.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button
            variant="outline"
            size="sm"
            onClick={() => exportAlertsToCSV(sortedAlerts)}
            disabled={sortedAlerts.length === 0}
            className="h-8"
          >
            <Download className="h-3.5 w-3.5 mr-1.5" />
            Export
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={fetchAlerts}
            disabled={isLoading}
            className="h-8"
          >
            <RefreshCw className={`h-3.5 w-3.5 mr-1.5 ${isLoading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="h-8"
            onClick={() =>
              window.open(
                `${alertaURL || ALERTA_FALLBACK_URL}/alert/`,
                '_blank',
                'noopener,noreferrer',
              )
            }
          >
            <ExternalLink className="h-3.5 w-3.5 mr-1.5" />
            Alerta UI
          </Button>
        </div>
      </div>

      {/* ---- Severity summary bar ---- */}
      <div className="flex items-center gap-2 px-6 py-3 border-b bg-muted/30 shrink-0 overflow-x-auto">
        {SEVERITY_ORDER.map(sev => (
          <SeverityCountCard key={sev} severity={sev} count={severityCounts[sev] ?? 0} />
        ))}
        <span className="ml-auto text-sm text-muted-foreground whitespace-nowrap">
          Total: <span className="font-semibold text-foreground">{alerts.length}</span>
        </span>
      </div>

      {/* ---- Status filter tabs ---- */}
      <div className="flex items-center gap-1 px-6 py-2 border-b bg-background shrink-0">
        {STATUS_FILTERS.map(f => (
          <button
            key={f.value}
            onClick={() => setActiveFilter(f.value)}
            className={`px-3 py-1 rounded-md text-sm font-medium transition-colors ${
              activeFilter === f.value
                ? 'bg-primary text-primary-foreground'
                : 'text-muted-foreground hover:text-foreground hover:bg-muted'
            }`}
          >
            {f.label}
            {f.statuses.length > 0 && (
              <span className="ml-1.5 text-xs opacity-70">
                (
                {alerts.filter(a => f.statuses.includes(a.status as AlertStatus)).length}
                )
              </span>
            )}
          </button>
        ))}
      </div>

      {/* ---- Error state ---- */}
      {error && (
        <div className="flex items-center gap-2 px-6 py-3 bg-destructive/10 border-b border-destructive/20 shrink-0">
          <AlertTriangle className="h-4 w-4 text-destructive shrink-0" />
          <span className="text-sm text-destructive">{error}</span>
          <Button
            variant="ghost"
            size="sm"
            className="ml-auto h-7 text-destructive"
            onClick={fetchAlerts}
          >
            Retry
          </Button>
        </div>
      )}

      {/* ---- Alerts table ---- */}
      <div className="flex-1 overflow-hidden p-4">
        <DataGrid
          data={sortedAlerts}
          columns={columns}
          tableKey="alarm-alerta-alerts"
          tableHeight="100%"
          searchableColumns={['resource', 'event', 'text', 'service']}
          useTableSorting={true}
          rowCursor={true}
          onRowClick={(row: AlertaAlert) => setSheet(<AlertHistorySheet alert={row} />)}
          emptyCustomMessage={
            isLoading ? 'Loading alerts...' : 'No alerts match the current filter'
          }
        />
      </div>
    </main>
  );
}
