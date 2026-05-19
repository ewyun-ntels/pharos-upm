import React, { useMemo } from 'react';
import { SelectBox } from '@pharos/shared/components/ui-extension/select/box';
import { Button, Collapsible, CollapsibleContent, CollapsibleTrigger } from '@pharos/shared/components/ui';
import { ChevronRight, Database } from 'lucide-react';
import { datasourceEditorRegistry, DefaultSQLEditor } from '@features/dashboard/datasources';
import { VariableReferencePanel } from './VariableReferencePanel';

// Import core datasource editors
import '@features/dashboard/datasources';

interface DatasourceEditorProps {
  panelId: string;
  datasourceName: string;
  datasourceType: string;
  query: string;
  onDatasourceChange: (name: string) => void;
  onQueryChange: (query: string) => void;
  onRunQuery: () => void;
  datasourceOptions: Array<{ value: string; label: string; type: string }>;
  isLoading?: boolean;
}

export const DatasourceEditor: React.FC<DatasourceEditorProps> = ({
  panelId,
  datasourceName,
  datasourceType,
  query,
  onDatasourceChange,
  onQueryChange,
  onRunQuery,
  datasourceOptions,
  isLoading = false,
}) => {
  const selectedDatasource = useMemo(
    () => datasourceOptions.find((ds) => ds.value === datasourceName),
    [datasourceOptions, datasourceName],
  );

  const editorPlugin = useMemo(() => {
    return datasourceEditorRegistry.get(datasourceType) ?? datasourceEditorRegistry.get('default');
  }, [datasourceType]);

  const renderQueryEditor = () => {
    const props = { panelId, datasourceName, query, onQueryChange, onRunQuery, isLoading };

    if (!editorPlugin) {
      console.warn(`No editor plugin found for datasource type: ${datasourceType}`);
      return React.createElement(DefaultSQLEditor, props);
    }

    return React.createElement(editorPlugin.component, props);
  };

  return (
    <div className="flex flex-col h-full">
      {/* Datasource Selection */}
      <div className="shrink-0 p-4 border-b bg-slate-50/50 dark:bg-slate-900/50">
        <div className="flex items-center gap-4">
          <div className="flex-1">
            <label className="block text-sm font-semibold mb-2 text-slate-700 dark:text-slate-300">
              Data Source
            </label>
            <SelectBox
              value={datasourceName}
              options={datasourceOptions}
              onChange={onDatasourceChange}
              placeholder="Select datasource..."
              className="w-full"
            />
          </div>
          <div className="shrink-0 pt-6">
            <Button
              onClick={onRunQuery}
              disabled={!datasourceName || !query.trim() || isLoading}
              size="sm"
            >
              {isLoading ? 'Running...' : 'Run Query'}
            </Button>
          </div>
        </div>

        {/* Datasource Info + Variable Reference */}
        {selectedDatasource && (
          <Collapsible>
            <div className="mt-2 flex items-center gap-2 text-xs text-muted-foreground">
              <Database className="h-3 w-3" />
              <span>Type: <span className="font-medium">{selectedDatasource.type}</span></span>
              <span className="text-muted-foreground/50">|</span>
              <span>Name: <span className="font-medium">{selectedDatasource.label}</span></span>
              <span className="text-muted-foreground/50">|</span>
              <CollapsibleTrigger className="flex items-center gap-0.5 hover:text-foreground transition-colors [&[data-state=open]>svg]:rotate-90">
                <ChevronRight className="h-3 w-3 transition-transform duration-150" />
                <span>내장 변수 참조</span>
              </CollapsibleTrigger>
            </div>
            <CollapsibleContent>
              <VariableReferencePanel />
            </CollapsibleContent>
          </Collapsible>
        )}
      </div>

      {/* Query Editor */}
      <div className="flex-1 overflow-hidden p-4">
        {datasourceName ? (
          renderQueryEditor()
        ) : (
          <div className="h-full flex items-center justify-center text-muted-foreground">
            <div className="text-center space-y-2">
              <Database className="h-12 w-12 mx-auto text-muted-foreground/50" />
              <p className="text-base font-medium">Select a data source to start</p>
              <p className="text-sm text-muted-foreground/70">Choose from SQL databases, ClickHouse, or other data sources</p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
