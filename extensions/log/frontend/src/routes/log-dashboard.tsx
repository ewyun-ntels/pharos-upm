import { axiosInstance } from '@/lib/axios';
import { useSheetStore } from '@components/sheet/sheetStore';
import {
    SheetTemplate,
    SheetTemplateContent,
    SheetTemplateGroup,
    SheetTemplateGroupGrid,
    SheetTemplateGroupGridItem,
    SheetTemplateHeader,
} from '@pharos/shared/components/template/sheet';
import {
    Button,
    Input,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@pharos/shared/components/ui';
import { IconButton } from '@pharos/shared/components/ui-extension';
import { DataGrid } from '@pharos/shared/components/ui-extension/data-grid';
import { ColumnDef } from '@tanstack/react-table';
import { AlertTriangle, Database, Download, Filter, RefreshCw, Search, X } from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface LogHit {
  _index: string;
  _id: string;
  _source: Record<string, unknown>;
}

interface SearchResponse {
  hits: {
    total: {value: number; relation: string};
    hits: LogHit[];
  };
}

interface FilterDef {
  name: string;
  field: string;
  display: string;
  case_insensitive?: boolean;
}

interface IndexSchema {
  pattern: string;
  filters: FilterDef[];
}

interface AggregationResponse {
  [key: string]: string[];
}

interface LogConfig {
  indices: string[];
  address: string;
  index_schema: IndexSchema[];
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const CUSTOM_RANGE_KEY = '__custom__';

const TIME_RANGES = [
  {label: 'Last 15 min',  from: 'now-15m',       to: 'now'},
  {label: 'Last 1 hour',  from: 'now-1h',         to: 'now'},
  {label: 'Last 4 hours', from: 'now-4h',         to: 'now'},
  {label: 'Last 12 hours',from: 'now-12h',        to: 'now'},
  {label: 'Last 24 hours',from: 'now-24h',        to: 'now'},
  {label: 'Last 7 days',  from: 'now-7d',         to: 'now'},
  {label: 'Today',        from: 'now/d',          to: 'now/d'},
  {label: 'Custom range', from: CUSTOM_RANGE_KEY, to: CUSTOM_RANGE_KEY},
];

const REFRESH_OPTIONS = [
  {label: 'Off', value: 0},
  {label: '5s',  value: 5_000},
  {label: '10s', value: 10_000},
  {label: '15s', value: 15_000},
  {label: '30s', value: 30_000},
  {label: '1m',  value: 60_000},
  {label: '5m',  value: 300_000},
];

const SIZE_OPTIONS = [
  {label: '100',  value: 100},
  {label: '250',  value: 250},
  {label: '500',  value: 500},
  {label: '1000', value: 1000},
];

const LEVEL_STYLES: Record<string, string> = {
  error:   'bg-red-600 text-white',
  err:     'bg-red-600 text-white',
  fatal:   'bg-red-800 text-white',
  warn:    'bg-yellow-400 text-black',
  warning: 'bg-yellow-400 text-black',
  info:    'bg-blue-500 text-white',
  debug:   'bg-slate-400 text-white',
  trace:   'bg-slate-300 text-black',
};

const ALL = '';

function toDatetimeLocal(d: Date): string {
  const p = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}T${p(d.getHours())}:${p(d.getMinutes())}`;
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Returns the schema whose pattern prefix best (longest) matches the index. */
function findSchemaForIndex(schemas: IndexSchema[], index: string): IndexSchema | null {
  let best: IndexSchema | null = null;
  let bestLen = -1;
  for (const schema of schemas) {
    const prefix = schema.pattern.replace(/\*$/, '');
    if (index.startsWith(prefix) && prefix.length > bestLen) {
      best = schema;
      bestLen = prefix.length;
    }
  }
  return best;
}

function getField(source: Record<string, unknown>, ...paths: string[]): string {
  for (const path of paths) {
    if (path in source) {
      const v = source[path];
      if (v !== null && v !== undefined && v !== '') return String(v);
    }
    const parts = path.split('.');
    let obj: unknown = source;
    for (const p of parts) {
      if (obj === null || typeof obj !== 'object') { obj = undefined; break; }
      obj = (obj as Record<string, unknown>)[p];
    }
    if (obj !== null && obj !== undefined && obj !== '') return String(obj);
  }
  return '';
}

function formatTime(iso: string): string {
  if (!iso) return '';
  try { return new Date(iso).toLocaleString(); } catch { return iso; }
}

function LevelBadge({level}: {level: string}) {
  const cls = LEVEL_STYLES[level.toLowerCase()] ?? 'bg-gray-400 text-white';
  return (
    <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-xs font-semibold uppercase ${cls}`}>
      {level}
    </span>
  );
}

function hasActiveFilters(filters: Record<string, string>): boolean {
  return Object.values(filters).some(v => v !== ALL);
}

// ---------------------------------------------------------------------------
// CSV export
// ---------------------------------------------------------------------------

function exportToCSV(hits: LogHit[]) {
  const headers = ['Time', 'Index', 'Namespace', 'Pod', 'Level', 'Message', '_source'];
  const rows = hits.map(h => {
    const s = h._source;
    return [
      getField(s, '@timestamp'),
      h._index,
      getField(s, 'kubernetes.namespace_name', 'namespace'),
      getField(s, 'kubernetes.pod_name', 'pod'),
      getField(s, 'level', 'log.level', 'severity'),
      getField(s, 'msg', 'message', 'log', 'MESSAGE'),
      JSON.stringify(s).replace(/[\r\n]+/g, ' '),
    ];
  });
  const csv = [headers, ...rows]
    .map(row => row.map(cell => `"${String(cell).replace(/"/g, '""')}"`).join(','))
    .join('\r\n');
  const blob = new Blob(['﻿' + csv], {type: 'text/csv;charset=utf-8;'});
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = `logs_${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}.csv`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

// ---------------------------------------------------------------------------
// Detail sheet — shows all _source fields
// ---------------------------------------------------------------------------

function flattenObject(
  obj: unknown,
  prefix = '',
  result: [string, string][] = [],
): [string, string][] {
  if (obj === null || obj === undefined) return result;
  if (typeof obj !== 'object') { result.push([prefix, String(obj)]); return result; }
  if (Array.isArray(obj)) { result.push([prefix, JSON.stringify(obj)]); return result; }
  const rec = obj as Record<string, unknown>;
  for (const key of Object.keys(rec)) {
    const fullKey = prefix ? `${prefix}.${key}` : key;
    const val = rec[key];
    if (val !== null && val !== undefined && typeof val === 'object' && !Array.isArray(val)) {
      flattenObject(val, fullKey, result);
    } else {
      result.push([fullKey, val === null || val === undefined ? '' : String(val)]);
    }
  }
  return result;
}

function LogDetailSheet({hit}: {hit: LogHit}) {
  const {setClose} = useSheetStore();
  const ts = getField(hit._source, '@timestamp');
  const namespace = getField(hit._source, 'kubernetes.namespace_name', 'namespace');
  const pod = getField(hit._source, 'kubernetes.pod_name', 'pod');
  const flatEntries = flattenObject(hit._source);

  return (
    <SheetTemplate>
      <SheetTemplateHeader
        title={`${namespace || hit._index} · ${pod || hit._id}`}
        leftInfoGroup={<span className="text-xs text-muted-foreground font-mono">{formatTime(ts)}</span>}
        rightButtonGroup={
          <IconButton icon={<X className="h-4 w-4" />} variant="ghost" className="h-7 w-7 text-muted-foreground" onClick={setClose}>
            Close
          </IconButton>
        }
      />
      <SheetTemplateContent>
        <SheetTemplateGroup title="Log Fields">
          <SheetTemplateGroupGrid cols={2}>
            {flatEntries.map(([k, v]) => (
              <SheetTemplateGroupGridItem key={k} cols={v.length > 60 ? 2 : 1} label={k} value={v} />
            ))}
          </SheetTemplateGroupGrid>
        </SheetTemplateGroup>
      </SheetTemplateContent>
    </SheetTemplate>
  );
}

// ---------------------------------------------------------------------------
// Table column helpers (static parts reused in useMemo below)
// ---------------------------------------------------------------------------

const TIME_COL: ColumnDef<LogHit> = {
  id: 'timestamp',
  header: 'Time',
  size: 170,
  cell: ({row}) => (
    <span className="text-xs text-muted-foreground whitespace-nowrap font-mono">
      {formatTime(getField(row.original._source, '@timestamp'))}
    </span>
  ),
};

const MESSAGE_COL: ColumnDef<LogHit> = {
  id: 'message',
  header: 'Message',
  cell: ({row}) => {
    const msg = getField(row.original._source, 'msg', 'message', 'log', 'MESSAGE');
    const err = getField(row.original._source, 'error', 'err');
    const text = err ? `${msg} [error: ${err}]` : msg;
    return (
      <span className="text-xs text-muted-foreground line-clamp-2 break-all">{text}</span>
    );
  },
};

// ---------------------------------------------------------------------------
// FilterSelect — compact select with "All" sentinel
// ---------------------------------------------------------------------------

function FilterSelect({
  placeholder,
  value,
  options,
  onChange,
  disabled,
}: {
  placeholder: string;
  value: string;
  options: string[];
  onChange: (v: string) => void;
  disabled?: boolean;
}) {
  return (
    <Select
      value={value === ALL ? '__all__' : value}
      onValueChange={v => onChange(v === '__all__' ? ALL : v)}
      disabled={disabled}
    >
      <SelectTrigger className={`h-8 text-xs min-w-[140px] max-w-[200px] ${value !== ALL ? 'border-primary text-primary' : ''}`}>
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="__all__">{placeholder} (All)</SelectItem>
        {options.map(opt => (
          <SelectItem key={opt} value={opt}>{opt}</SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

// ---------------------------------------------------------------------------
// Main component
// ---------------------------------------------------------------------------

export function LogDashboard() {
  // --- config / index / schema
  const [indices, setIndices]             = useState<string[]>(['app-*', 'node-*']);
  const [selectedIndex, setSelectedIndex] = useState('app-*');
  const [indexSchemas, setIndexSchemas]   = useState<IndexSchema[]>([]);
  const [currentSchema, setCurrentSchema] = useState<IndexSchema | null>(null);

  // --- time / refresh
  const [timeRange, setTimeRange]           = useState(TIME_RANGES[0]);
  const [customFrom, setCustomFrom]         = useState(() => { const d = new Date(); d.setHours(d.getHours() - 1); return toDatetimeLocal(d); });
  const [customTo, setCustomTo]             = useState(() => toDatetimeLocal(new Date()));
  const [refreshInterval, setRefreshInterval] = useState(0);
  const [size, setSize]                     = useState(500);

  // --- search
  const [inputQuery, setInputQuery] = useState('');
  const [query, setQuery]           = useState('');

  // --- dynamic filters (keyed by filterDef.name)
  const [filterOptions, setFilterOptions] = useState<Record<string, string[]>>({});
  const [filterValues, setFilterValues]   = useState<Record<string, string>>({});

  // --- results
  const [hits, setHits]                   = useState<LogHit[]>([]);
  const [total, setTotal]                 = useState(0);
  const [isLoading, setIsLoading]         = useState(false);
  const [error, setError]                 = useState<string | null>(null);
  const [lastRefreshed, setLastRefreshed] = useState<Date | null>(null);

  const {setSheet} = useSheetStore();
  const timerRef   = useRef<ReturnType<typeof setInterval> | null>(null);
  const isFirst    = useRef(true);

  // ---------------------------------------------------------------------------
  // Derived time values
  // ---------------------------------------------------------------------------

  const isCustom      = timeRange.from === CUSTOM_RANGE_KEY;
  const effectiveFrom = isCustom ? new Date(customFrom).toISOString() : timeRange.from;
  const effectiveTo   = isCustom ? new Date(customTo).toISOString()   : timeRange.to;

  // ---------------------------------------------------------------------------
  // Data fetching
  // ---------------------------------------------------------------------------

  /**
   * Fetches aggregation buckets for all filter dropdowns.
   * Pass appliedFilters to enable cascading (e.g. namespace narrows pod list).
   * Pass schema explicitly so the callback doesn't need it as a closure dep.
   */
  const fetchAggregations = useCallback(async (
    appliedFilters: Record<string, string> = {},
    schema: IndexSchema | null = null,
  ) => {
    try {
      const params = new URLSearchParams({
        index:     selectedIndex,
        from_time: effectiveFrom,
        to_time:   effectiveTo,
      });
      Object.entries(appliedFilters).forEach(([k, v]) => { if (v) params.set(k, v); });

      const resp = await axiosInstance.get<AggregationResponse>(`/log/aggregations?${params.toString()}`);

      // Normalize case-insensitive fields (e.g. "INFO"/"info" → "info")
      const normalized: Record<string, string[]> = {};
      for (const [key, values] of Object.entries(resp.data)) {
        const strValues = values as string[];
        const def = schema?.filters.find(f => f.name === key);
        normalized[key] = def?.case_insensitive
          ? [...new Set(strValues.map((v: string) => v.toLowerCase()))]
          : strValues;
      }
      setFilterOptions(normalized);
    } catch {
      // Silently ignore — filter dropdowns are best-effort
    }
  }, [selectedIndex, effectiveFrom, effectiveTo]);

  /**
   * Fetches log hits from OpenSearch via /log/search.
   * Filters are sent as a map so the backend can match them against the index schema.
   */
  const fetchLogs = useCallback(async (overrides?: {
    query?: string;
    filters?: Record<string, string>;
  }) => {
    setIsLoading(true);
    setError(null);
    try {
      const resp = await axiosInstance.post<SearchResponse>('/log/search', {
        index:     selectedIndex,
        query:     overrides?.query   ?? query,
        from_time: effectiveFrom,
        to_time:   effectiveTo,
        from:      0,
        size,
        filters:   overrides?.filters ?? filterValues,
      });
      setHits(resp.data.hits?.hits ?? []);
      setTotal(resp.data.hits?.total?.value ?? 0);
      setLastRefreshed(new Date());
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to fetch logs');
    } finally {
      setIsLoading(false);
    }
  }, [selectedIndex, query, effectiveFrom, effectiveTo, size, filterValues]);

  // ---------------------------------------------------------------------------
  // Effects
  // ---------------------------------------------------------------------------

  // Initial load: fetch config (indices + schemas), then prime aggregations and logs.
  useEffect(() => {
    axiosInstance.get<LogConfig>('/log/config').then(r => {
      const schemas  = r.data.index_schema ?? [];
      const firstIdx = r.data.indices?.[0] ?? selectedIndex;
      if (r.data.indices?.length) setIndices(r.data.indices);
      if (schemas.length)         setIndexSchemas(schemas);
      const schema = findSchemaForIndex(schemas, firstIdx);
      setCurrentSchema(schema);
      fetchAggregations({}, schema);
    }).catch(() => {
      fetchAggregations();
    });
    fetchLogs();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Re-fetch when index, time range, or page size changes (skip first mount).
  useEffect(() => {
    if (isFirst.current) { isFirst.current = false; return; }
    const schema = findSchemaForIndex(indexSchemas, selectedIndex);
    setCurrentSchema(schema);
    setFilterValues({});
    setFilterOptions({});
    fetchAggregations({}, schema);
    fetchLogs({filters: {}});
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedIndex, timeRange, size]);

  // Auto-refresh: reset interval when fetchLogs reference or interval setting changes.
  useEffect(() => {
    if (timerRef.current) clearInterval(timerRef.current);
    if (refreshInterval > 0) {
      timerRef.current = setInterval(() => fetchLogs(), refreshInterval);
    }
    return () => { if (timerRef.current) clearInterval(timerRef.current); };
  }, [fetchLogs, refreshInterval]);

  // ---------------------------------------------------------------------------
  // Event handlers
  // ---------------------------------------------------------------------------

  function handleSearch() {
    setQuery(inputQuery);
    fetchLogs({query: inputQuery});
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter') handleSearch();
  }

  function handleClearSearch() {
    setInputQuery('');
    setQuery('');
    fetchLogs({query: ''});
  }

  function handleFilterChange(name: string, value: string) {
    const next = {...filterValues, [name]: value};
    setFilterValues(next);
    // Re-fetch aggregations with the new filter applied (enables cascading dropdowns).
    fetchAggregations(next, currentSchema);
    fetchLogs({filters: next});
  }

  function clearAllFilters() {
    setFilterValues({});
    fetchAggregations({}, currentSchema);
    fetchLogs({filters: {}});
  }

  const activeFilters = hasActiveFilters(filterValues);
  const refreshLabel  = REFRESH_OPTIONS.find(o => o.value === refreshInterval)?.label ?? 'Off';

  // ---------------------------------------------------------------------------
  // Dynamic table columns — Time + schema filters + Message
  // ---------------------------------------------------------------------------

  const tableColumns = useMemo<ColumnDef<LogHit>[]>(() => {
    if (!currentSchema || currentSchema.filters.length === 0) {
      return [TIME_COL, MESSAGE_COL];
    }

    const filterCols: ColumnDef<LogHit>[] = currentSchema.filters.map((f: FilterDef) => {
      // Strip .keyword suffix to read values from _source
      const srcField = f.field.replace(/\.keyword$/, '');
      return {
        id:     f.name,
        header: f.display,
        size:   f.name === 'level' || f.name === 'priority' ? 80 : 160,
        cell: ({row}: {row: {original: LogHit}}) => {
          const value = getField(row.original._source, srcField);
          if (f.name === 'level') {
            return value ? <LevelBadge level={value} /> : null;
          }
          return <span className="text-xs truncate">{value}</span>;
        },
      };
    });

    return [TIME_COL, ...filterCols, MESSAGE_COL];
  }, [currentSchema]);

  // ---------------------------------------------------------------------------
  // Render
  // ---------------------------------------------------------------------------

  return (
    <main className="flex flex-col w-full h-full gap-0">

      {/* ---- Header ---- */}
      <div className="flex items-center justify-between px-6 py-3 border-b bg-background shrink-0">
        <div>
          <h1 className="text-lg font-semibold">Logs</h1>
          {lastRefreshed && (
            <p className="text-xs text-muted-foreground mt-0.5">
              Last refreshed: {lastRefreshed.toLocaleTimeString()}
              {refreshInterval > 0 && ` · auto-refresh ${refreshLabel}`}
            </p>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Select value={String(refreshInterval)} onValueChange={v => setRefreshInterval(Number(v))}>
            <SelectTrigger className="h-8 w-24 text-xs"><SelectValue /></SelectTrigger>
            <SelectContent>
              {REFRESH_OPTIONS.map(o => (
                <SelectItem key={o.value} value={String(o.value)}>{o.label}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button variant="outline" size="sm" onClick={() => exportToCSV(hits)} disabled={hits.length === 0} className="h-8">
            <Download className="h-3.5 w-3.5 mr-1.5" />
            Download CSV
          </Button>
          <Button variant="outline" size="sm" onClick={() => fetchLogs()} disabled={isLoading} className="h-8">
            <RefreshCw className={`h-3.5 w-3.5 mr-1.5 ${isLoading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      {/* ---- Primary toolbar: index / time / search ---- */}
      <div className="flex items-center gap-2 px-4 py-2 border-b bg-muted/20 shrink-0 flex-wrap">
        {/* Index selector */}
        <div className="flex items-center gap-1.5 shrink-0">
          <Database className="h-3.5 w-3.5 text-muted-foreground" />
          <Select value={selectedIndex} onValueChange={setSelectedIndex}>
            <SelectTrigger className="h-8 w-36 text-xs"><SelectValue /></SelectTrigger>
            <SelectContent>
              {indices.map(idx => <SelectItem key={idx} value={idx}>{idx}</SelectItem>)}
            </SelectContent>
          </Select>
        </div>

        {/* Time range */}
        <Select
          value={timeRange.from}
          onValueChange={v => setTimeRange(TIME_RANGES.find(t => t.from === v) ?? TIME_RANGES[0])}
        >
          <SelectTrigger className={`h-8 w-36 text-xs shrink-0 ${isCustom ? 'border-primary text-primary' : ''}`}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {TIME_RANGES.map(t => <SelectItem key={t.from} value={t.from}>{t.label}</SelectItem>)}
          </SelectContent>
        </Select>

        {/* Custom date/time pickers */}
        {isCustom && (
          <div className="flex items-center gap-1 shrink-0 flex-wrap">
            <input
              type="datetime-local"
              value={customFrom}
              onChange={e => setCustomFrom(e.target.value)}
              className="h-8 text-xs border rounded px-2 bg-background text-foreground [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:invert [&::-webkit-calendar-picker-indicator]:opacity-70 hover:[&::-webkit-calendar-picker-indicator]:opacity-100"
            />
            <span className="text-xs text-muted-foreground">~</span>
            <input
              type="datetime-local"
              value={customTo}
              onChange={e => setCustomTo(e.target.value)}
              className="h-8 text-xs border rounded px-2 bg-background text-foreground [&::-webkit-calendar-picker-indicator]:cursor-pointer [&::-webkit-calendar-picker-indicator]:invert [&::-webkit-calendar-picker-indicator]:opacity-70 hover:[&::-webkit-calendar-picker-indicator]:opacity-100"
            />
            <Button size="sm" className="h-8 px-3 text-xs" onClick={() => fetchLogs()}>
              Apply
            </Button>
          </div>
        )}

        {/* Search input */}
        <div className="flex-1 flex items-center gap-1 min-w-[200px]">
          <div className="relative flex-1">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
            <Input
              className="h-8 pl-8 pr-8 text-xs font-mono"
              placeholder='Lucene query  (e.g.  level:error  OR  msg:"timeout")'
              value={inputQuery}
              onChange={e => setInputQuery(e.target.value)}
              onKeyDown={handleKeyDown}
            />
            {inputQuery && (
              <button
                className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                onClick={handleClearSearch}
              >
                <X className="h-3.5 w-3.5" />
              </button>
            )}
          </div>
          <Button size="sm" className="h-8 px-3 text-xs" onClick={handleSearch}>Search</Button>
        </div>

        {/* Result size */}
        <Select value={String(size)} onValueChange={v => setSize(Number(v))}>
          <SelectTrigger className="h-8 w-24 text-xs shrink-0"><SelectValue /></SelectTrigger>
          <SelectContent>
            {SIZE_OPTIONS.map(o => <SelectItem key={o.value} value={String(o.value)}>{o.label} rows</SelectItem>)}
          </SelectContent>
        </Select>
      </div>

      {/* ---- Filter bar: rendered dynamically from the current index schema ---- */}
      <div className="flex items-center gap-2 px-4 py-2 border-b bg-background shrink-0 flex-wrap">
        <Filter className="h-3.5 w-3.5 text-muted-foreground shrink-0" />

        {currentSchema?.filters.map(f => (
          <FilterSelect
            key={f.name}
            placeholder={f.display}
            value={filterValues[f.name] ?? ALL}
            options={filterOptions[f.name] ?? []}
            onChange={v => handleFilterChange(f.name, v)}
            disabled={(filterOptions[f.name] ?? []).length === 0}
          />
        ))}

        {activeFilters && (
          <Button
            variant="ghost"
            size="sm"
            className="h-8 text-xs text-muted-foreground hover:text-foreground"
            onClick={clearAllFilters}
          >
            <X className="h-3 w-3 mr-1" />
            Clear filters
          </Button>
        )}

        {!currentSchema && (
          <span className="text-xs text-muted-foreground">
            — no filter schema configured for this index
          </span>
        )}

        {currentSchema && !activeFilters && Object.keys(filterOptions).length === 0 && (
          <span className="text-xs text-muted-foreground">
            — filter values load after first search
          </span>
        )}
      </div>

      {/* ---- Stats bar ---- */}
      <div className="flex items-center gap-3 px-4 py-1.5 border-b bg-background shrink-0 text-xs text-muted-foreground">
        <span>
          Showing{' '}
          <span className="font-semibold text-foreground">{hits.length.toLocaleString()}</span>
          {' / '}
          <span className="font-semibold text-foreground">{total.toLocaleString()}</span>
          {' '}results
          {query && <> for <span className="font-mono text-foreground">"{query}"</span></>}
        </span>
        <span className="text-muted-foreground/50">·</span>
        <span>{selectedIndex}</span>
        <span className="text-muted-foreground/50">·</span>
        <span>{isCustom ? `${customFrom.replace('T', ' ')} ~ ${customTo.replace('T', ' ')}` : timeRange.label}</span>
        {activeFilters && (
          <>
            <span className="text-muted-foreground/50">·</span>
            <span className="text-primary font-medium">
              {Object.entries(filterValues)
                .filter(([, v]) => v)
                .map(([k, v]) => `${k}:${v}`)
                .join('  ')}
            </span>
          </>
        )}
      </div>

      {/* ---- Error state ---- */}
      {error && (
        <div className="flex items-center gap-2 px-4 py-2 bg-destructive/10 border-b border-destructive/20 shrink-0">
          <AlertTriangle className="h-4 w-4 text-destructive shrink-0" />
          <span className="text-sm text-destructive">{error}</span>
          <Button variant="ghost" size="sm" className="ml-auto h-7 text-destructive" onClick={() => fetchLogs()}>Retry</Button>
        </div>
      )}

      {/* ---- Data table ---- */}
      <div className="flex-1 overflow-hidden p-4">
        <DataGrid
          data={hits}
          columns={tableColumns}
          tableKey="log-logs"
          tableHeight="100%"
          searchableColumns={[]}
          useTableSorting={false}
          rowCursor={true}
          onRowClick={(row: LogHit) => setSheet(<LogDetailSheet hit={row} />)}
          emptyCustomMessage={isLoading ? 'Loading logs...' : 'No logs found for the current filter'}
        />
      </div>
    </main>
  );
}
