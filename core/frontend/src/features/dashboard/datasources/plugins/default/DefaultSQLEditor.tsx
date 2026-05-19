import React from 'react';
import AceEditor from '@lib/ace-editor';
import 'ace-builds/src-noconflict/mode-sql';
import 'ace-builds/src-noconflict/theme-github';
import 'ace-builds/src-noconflict/ext-language_tools';
import 'ace-builds/src-noconflict/theme-tomorrow_night';
import { useTheme } from '@providers/theme-provider';
import { DatasourceEditorComponentProps } from '@features/dashboard/datasources';
import { Database } from 'lucide-react';

/**
 * Default SQL Editor Component
 * Used as fallback for datasource types without custom editors
 */
export const DefaultSQLEditor: React.FC<DatasourceEditorComponentProps> = ({
  panelId,
  query,
  onQueryChange,
  isLoading = false
}) => {
  const { resolvedTheme } = useTheme();

  return (
    <div className="h-full flex flex-col">
      {/* SQL Editor 헤더 */}
      <div className="flex items-center justify-between mb-3 p-3 border border-border rounded-lg">
        <div className="flex items-center gap-2">
          <Database className="h-4 w-4 text-slate-600 dark:text-slate-400" />
          <span className="text-slate-700 dark:text-slate-300 font-medium text-sm">SQL Query Editor</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs text-slate-600 dark:text-slate-400">
            💡 SELECT, JOIN, WHERE, GROUP BY
          </span>
        </div>
      </div>

      {/* SQL Editor */}
      <div className="flex-1 min-h-0 border border-border rounded-lg overflow-hidden shadow-sm">
        <AceEditor
          mode="sql"
          theme={resolvedTheme === 'dark' ? 'tomorrow_night' : 'github'}
          onChange={onQueryChange}
          value={query}
          name={`sql-editor-${panelId}`}
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
