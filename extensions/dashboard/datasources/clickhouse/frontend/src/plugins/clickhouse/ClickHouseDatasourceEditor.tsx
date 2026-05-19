import React from 'react';
import { DatasourceEditorComponentProps } from '@pharos/core/datasource-editor-registry';
import { Database } from 'lucide-react';
import _AceEditor from 'react-ace';
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const AceEditor = ((_AceEditor as any).default ?? _AceEditor) as typeof _AceEditor;
import 'ace-builds/src-noconflict/mode-sql';
import 'ace-builds/src-noconflict/theme-github';
import 'ace-builds/src-noconflict/ext-language_tools';
import 'ace-builds/src-noconflict/theme-tomorrow_night';
import { useTheme } from '@providers/theme-provider';

export const ClickHouseDatasourceEditor: React.FC<DatasourceEditorComponentProps> = ({
  panelId,
  datasourceName,
  query,
  onQueryChange,
  isLoading = false
}) => {
  const { resolvedTheme } = useTheme();

  return (
    <div className="h-full flex flex-col">
      {/* ClickHouse 헤더 */}
      <div className="flex items-center justify-between mb-3 p-3 bg-linear-to-r from-slate-50 to-gray-50 dark:from-slate-900 dark:to-gray-900 border border-slate-200 dark:border-slate-700 rounded-lg">
        <div className="flex items-center gap-2">
          <Database className="h-4 w-4 text-slate-600 dark:text-slate-400" />
          <span className="text-slate-700 dark:text-slate-300 font-medium text-sm">ClickHouse Query Editor</span>
          {datasourceName && (
            <span className="text-xs text-muted-foreground">· {datasourceName}</span>
          )}
        </div>
      </div>

      {/* ClickHouse SQL Editor */}
      <div className="flex-1 min-h-0 border border-border rounded-lg overflow-hidden shadow-sm">
        <AceEditor
          mode="sql"
          theme={resolvedTheme === 'dark' ? 'tomorrow_night' : 'github'}
          onChange={onQueryChange}
          value={query}
          name={`clickhouse-editor-${panelId}`}
          editorProps={{ $blockScrolling: true }}
          width="100%"
          height="100%"
          fontSize={14}
          showPrintMargin={false}
          showGutter={true}
          highlightActiveLine={true}
          readOnly={isLoading}
          setOptions={{
            enableBasicAutocompletion: true,
            enableLiveAutocompletion: true,
            enableSnippets: false,
            showLineNumbers: true,
            tabSize: 2,
            useWorker: false
          }}
        />
      </div>
    </div>
  );
};
