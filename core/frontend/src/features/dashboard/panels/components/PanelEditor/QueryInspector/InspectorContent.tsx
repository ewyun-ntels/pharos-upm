import React from 'react';
import AceEditor from '@lib/ace-editor';
import 'ace-builds/src-noconflict/mode-sql';
import 'ace-builds/src-noconflict/theme-github';
import 'ace-builds/src-noconflict/theme-tomorrow_night';
import {AlertCircle, CheckCircle2, Copy} from 'lucide-react';
import {Button} from '@pharos/shared/components/ui';
import type {InspectInfo, QueryInspectResponseStatistics} from '@pharos/shared/types/dashboard';
import {useTheme} from '@providers/theme-provider';
import {useToast} from '@hooks/use-toast';
import {cn} from '@lib/utils';
import {formatDateTime} from '@lib/format-date';

interface InspectorContentProps {
  inspect: InspectInfo;
  statistics?: QueryInspectResponseStatistics;
  rows?: number;
}

export function InspectorContent({inspect, statistics, rows}: InspectorContentProps) {
  const {resolvedTheme} = useTheme();
  const {toast} = useToast();

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast({title: 'Copied', description: 'Copied to clipboard'});
  };

  const {rawQuery, executedQuery, variables, datasourceName, timestamp, error} = inspect;
  const hasError = !!error;
  const hasVariables = rawQuery !== executedQuery;

  return (
    <div className="space-y-6">
      {/* Query Status */}
      <div
        className={cn(
          'border rounded-lg p-4',
          hasError
            ? 'bg-red-50 border-red-200 dark:bg-red-950/20 dark:border-red-800'
            : hasVariables
              ? 'bg-blue-50 border-blue-200 dark:bg-blue-950/20 dark:border-blue-800'
              : 'bg-green-50 border-green-200 dark:bg-green-950/20 dark:border-green-800',
        )}
      >
        <div className="flex items-center gap-2">
          {hasError ? (
            <>
              <AlertCircle className="h-5 w-5 text-red-600 dark:text-red-400" />
              <span className="font-medium text-red-900 dark:text-red-100">Query Failed</span>
            </>
          ) : hasVariables ? (
            <>
              <AlertCircle className="h-5 w-5 text-blue-600 dark:text-blue-400" />
              <span className="font-medium text-blue-900 dark:text-blue-100">Variables Applied</span>
            </>
          ) : (
            <>
              <CheckCircle2 className="h-5 w-5 text-green-600 dark:text-green-400" />
              <span className="font-medium text-green-900 dark:text-green-100">No Variables</span>
            </>
          )}
        </div>
        <p
          className={cn(
            'text-sm mt-1',
            hasError
              ? 'text-red-700 dark:text-red-300'
              : hasVariables
                ? 'text-blue-700 dark:text-blue-300'
                : 'text-green-700 dark:text-green-300',
          )}
        >
          {hasError
            ? 'The query failed during execution. Check the error message below.'
            : hasVariables
              ? 'The executed query has variables replaced with actual values.'
              : 'The query was executed without any variable substitutions.'}
        </p>
      </div>

      {/* Error Banner */}
      {error && (
        <div className="border border-destructive rounded-lg p-4 bg-destructive/5">
          <div className="flex items-start gap-3">
            <AlertCircle className="h-5 w-5 text-destructive mt-0.5 shrink-0" />
            <div className="flex-1 min-w-0">
              <h3 className="font-semibold text-destructive mb-2">Query Execution Error</h3>
              <pre className="bg-muted p-3 rounded-md text-xs overflow-auto max-h-36 font-mono whitespace-pre-wrap">
                {error}
              </pre>
            </div>
          </div>
        </div>
      )}

      {/* Original Query */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <h3 className="font-medium text-sm">Original Query</h3>
          <Button variant="ghost" size="sm" onClick={() => copyToClipboard(rawQuery)}>
            <Copy className="size-4" />
          </Button>
        </div>
        <div className="border border-border rounded-lg overflow-hidden">
          <AceEditor
            mode="sql"
            theme={resolvedTheme === 'dark' ? 'tomorrow_night' : 'github'}
            value={rawQuery || 'N/A'}
            readOnly
            name="original-query-inspector"
            editorProps={{$blockScrolling: true}}
            width="100%"
            height="160px"
            fontSize={13}
            showPrintMargin={false}
            showGutter={true}
            highlightActiveLine={false}
            setOptions={{useWorker: false, showLineNumbers: true, tabSize: 2}}
          />
        </div>
      </div>

      {/* Executed Query - only shown when variables were substituted */}
      {hasVariables && (
        <div>
          <div className="flex items-center justify-between mb-2">
            <h3 className="font-medium text-sm">Executed Query</h3>
            <Button variant="ghost" size="sm" onClick={() => copyToClipboard(executedQuery)}>
              <Copy className="size-4" />
            </Button>
          </div>
          <div className="border border-border rounded-lg overflow-hidden">
            <AceEditor
              mode="sql"
              theme={resolvedTheme === 'dark' ? 'tomorrow_night' : 'github'}
              value={executedQuery || 'N/A'}
              readOnly
              name="executed-query-inspector"
              editorProps={{$blockScrolling: true}}
              width="100%"
              height="160px"
              fontSize={13}
              showPrintMargin={false}
              showGutter={true}
              highlightActiveLine={false}
              setOptions={{useWorker: false, showLineNumbers: true, tabSize: 2}}
            />
          </div>
        </div>
      )}

      {/* Variables */}
      {Object.keys(variables || {}).length > 0 && (
        <div>
          <h3 className="font-medium text-sm mb-2">Variables</h3>
          <pre className="bg-muted p-4 rounded-md text-sm overflow-auto max-h-36">
            {JSON.stringify(variables, null, 2)}
          </pre>
        </div>
      )}

      {/* Statistics */}
      <div className="grid grid-cols-2 gap-4">
        <div className="border rounded-md p-4">
          <div className="text-xs text-muted-foreground mb-1">Elapsed Time</div>
          <div className="text-2xl font-medium">
            {statistics?.elapsed ? `${statistics.elapsed.toFixed(3)}s` : 'N/A'}
          </div>
        </div>
        <div className="border rounded-md p-4">
          <div className="text-xs text-muted-foreground mb-1">Rows Returned</div>
          <div className="text-2xl font-medium">{rows ?? 0}</div>
        </div>
      </div>

      {/* Datasource */}
      <div className="border rounded-md p-4">
        <div className="text-xs text-muted-foreground mb-1">Datasource</div>
        <div className="font-medium">{datasourceName || 'N/A'}</div>
      </div>

      {/* Timestamp */}
      <div className="border rounded-md p-4">
        <div className="text-xs text-muted-foreground mb-1">Timestamp</div>
        <div className="font-medium">
          {formatDateTime(timestamp) || 'N/A'}
        </div>
      </div>
    </div>
  );
}
