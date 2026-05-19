import React, {useCallback, useState} from 'react';
import {useDatasourceList} from '@hooks/use-datasource-list';
import {useDashboardStore} from '@features/dashboard/hooks/use-dashboard-store';
import {useVariableState} from '@features/dashboard/hooks/use-variable-state';
import {DatasourceEditor} from './DatasourceEditor';
import {ChartQuery} from '@pharos/shared/types/dashboard';
import {Button} from '@pharos/shared/components/ui';
import {Plus, X, ChevronDown, ChevronRight} from 'lucide-react';
import {InlineEdit} from './InlineEdit';
import {QueryInspectorSheet} from './QueryInspector';
import {usePanelArgs} from '@features/dashboard/hooks/use-panel-args';

interface LeftBottomPanelProps {
  panelId: string;
  queries: ChartQuery[];
  onQueriesChange: (queries: ChartQuery[]) => void;
  onRunQueries?: () => void;
}

const createEmptyQuery = (): ChartQuery => ({
  datasourceName: '',
  query: '',
  label: '',
});

const LeftBottomPanel = ({panelId, queries, onQueriesChange, onRunQueries}: LeftBottomPanelProps) => {
  const dashboardId = useDashboardStore((state) => state.id);
  const panel = useDashboardStore((state) => state.panelMap[panelId]);
  const {dataSourceList} = useDatasourceList();

  const [expandedIndices, setExpandedIndices] = useState<Set<number>>(new Set([0]));
  const [isLoading, setIsLoading] = useState(false);
  const [inspectorOpen, setInspectorOpen] = useState(false);

  const { filterValues, filterIds, datetime, step, refreshCount } = useVariableState();
  const firstQuery = queries[0];
  const isPrometheus = firstQuery?.datasourceName?.toLowerCase().includes('prometheus') ?? false;
  const args = usePanelArgs(filterValues, filterIds, datetime, step, refreshCount, isPrometheus ? 'regex' : 'sql');
  const inspectorVariables = Object.fromEntries(args) as Record<string, unknown>;

  const toggleExpanded = useCallback((index: number) => {
    setExpandedIndices((prev) => {
      const next = new Set(prev);
      if (next.has(index)) {
        next.delete(index);
      } else {
        next.add(index);
      }
      return next;
    });
  }, []);

  const handleAddQuery = useCallback(() => {
    const newIndex = queries.length;
    onQueriesChange([...queries, createEmptyQuery()]);
    setExpandedIndices((prev) => new Set([...prev, newIndex]));
  }, [queries, onQueriesChange]);

  const handleRemoveQuery = useCallback(
    (index: number) => {
      if (queries.length <= 1) return;
      onQueriesChange(queries.filter((_, i) => i !== index));
      setExpandedIndices((prev) => {
        const next = new Set<number>();
        prev.forEach((i) => {
          if (i < index) next.add(i);
          else if (i > index) next.add(i - 1);
        });
        return next;
      });
    },
    [queries, onQueriesChange],
  );

  const updateQueryField = useCallback((index: number, field: keyof ChartQuery, value: string) => {
    onQueriesChange(queries.map((q, i) => (i === index ? {...q, [field]: value} : q)));
  }, [queries, onQueriesChange]);

  const handleDatasourceChange = useCallback((index: number, value: string) => {
    onQueriesChange(
      queries.map((q, i) => (i === index ? {...q, datasourceName: value, query: ''} : q)),
    );
  }, [queries, onQueriesChange]);

  const handleRunQuery = useCallback(async () => {
    const validQueries = queries.filter((q) => q.datasourceName && q.query.trim());
    if (validQueries.length === 0) return;

    setIsLoading(true);
    try {
      // ✅ store 직접 수정 없이 부모 콜백으로 위임 (effectivePanel을 통한 미리보기만 업데이트)
      onRunQueries?.();
    } finally {
      setIsLoading(false);
    }
  }, [queries, onRunQueries]);

  const getDatasourceType = useCallback(
    (datasourceName: string) => {
      const ds = dataSourceList.find((d) => d.value === datasourceName);
      return ds?.type || 'default';
    },
    [dataSourceList],
  );

  return (
    <div className="h-full border-l border-border bg-background flex flex-col">
      {/* Header */}
      <div className="shrink-0 p-3 border-b border-border flex items-center justify-between">
        <span className="text-sm font-medium text-muted-foreground">
          Queries ({queries.length})
        </span>
        <div className="flex gap-2">
          <Button
            onClick={() => setInspectorOpen(true)}
            size="sm"
            variant="outline"
            disabled={!panel?.dataProvider}
          >
            Inspect
          </Button>
          <Button
            onClick={handleRunQuery}
            disabled={isLoading || queries.every((q) => !q.datasourceName || !q.query.trim())}
            size="sm"
          >
            {isLoading ? 'Running...' : 'Run All Queries'}
          </Button>
        </div>
      </div>

      {/* Query list */}
      <div className="flex-1 overflow-y-auto">
        {queries.map((query, index) => {
          const isExpanded = expandedIndices.has(index);

          return (
            <div key={index} className="border-b border-border">
              <div
                className="flex items-center gap-2 px-3 py-2 bg-muted/30 cursor-pointer hover:bg-muted/50 transition-colors"
                onClick={() => toggleExpanded(index)}
              >
                {isExpanded ? (
                  <ChevronDown className="h-4 w-4 text-muted-foreground shrink-0" />
                ) : (
                  <ChevronRight className="h-4 w-4 text-muted-foreground shrink-0" />
                )}
                <span className="flex items-center justify-center w-6 h-6 rounded bg-primary/10 text-primary text-xs font-bold shrink-0">
                  {String.fromCharCode(65 + index)}
                </span>
                <div className="flex-1 min-w-0 flex items-center gap-2">
                  <InlineEdit
                    value={query.label}
                    placeholder={`Query ${String.fromCharCode(65 + index)}`}
                    onSave={(value) => updateQueryField(index, 'label', value)}
                  />
                  {query.datasourceName && (
                    <span className="text-xs text-muted-foreground shrink-0">
                      ({query.datasourceName})
                    </span>
                  )}
                </div>
                {queries.length > 1 && (
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleRemoveQuery(index);
                    }}
                    className="h-6 w-6 p-0 hover:bg-destructive/20 shrink-0"
                    title="Remove query"
                  >
                    <X className="h-3.5 w-3.5 text-muted-foreground hover:text-destructive" />
                  </Button>
                )}
              </div>

              {isExpanded && (
                <div className="h-150">
                  <DatasourceEditor
                    panelId={panelId}
                    datasourceName={query.datasourceName}
                    datasourceType={getDatasourceType(query.datasourceName)}
                    query={query.query}
                    onDatasourceChange={(value) => handleDatasourceChange(index, value)}
                    onQueryChange={(value) => updateQueryField(index, 'query', value)}
                    onRunQuery={handleRunQuery}
                    datasourceOptions={dataSourceList.map((ds) => ({
                      value: ds.value,
                      label: ds.label,
                      type: ds.type || 'default',
                    }))}
                    isLoading={isLoading}
                  />
                </div>
              )}
            </div>
          );
        })}

        <div className="p-3">
          <Button variant="outline" size="sm" onClick={handleAddQuery} className="w-full">
            <Plus className="h-4 w-4 mr-2" />
            Add Query
          </Button>
        </div>
      </div>

      <QueryInspectorSheet
        open={inspectorOpen}
        onOpenChange={setInspectorOpen}
        dashboardId={dashboardId || ''}
        datasourceName={firstQuery?.datasourceName}
        query={firstQuery?.query}
        variables={inspectorVariables}
      />
    </div>
  );
};

export default LeftBottomPanel;
