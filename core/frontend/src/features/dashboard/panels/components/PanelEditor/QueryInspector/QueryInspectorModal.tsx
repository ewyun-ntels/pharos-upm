import React, { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  Button,
  Tabs,
  TabsList,
  TabsTrigger,
  TabsContent,
} from '@pharos/shared/components/ui';
import type { QueryInspectResponse } from '@pharos/shared/types/dashboard';
import { Copy } from 'lucide-react';
import { LoadingIndicator } from '@pharos/shared/components/ui-extension';
import { useToast } from '@hooks/use-toast';
import {formatDateTime} from '@lib/format-date';

interface QueryInspectorModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  data: QueryInspectResponse | null;
  isLoading?: boolean;
}

/**
 * Query Inspector Modal - displays query execution details
 * 
 * Features:
 * - Query tab: Shows original and executed queries
 * - Data tab: Shows query results in table format
 * - JSON tab: Shows full response in JSON format
 * - Stats tab: Shows execution statistics
 */
export function QueryInspectorModal({
  open,
  onOpenChange,
  data,
  isLoading = false,
}: QueryInspectorModalProps) {
  const [activeTab, setActiveTab] = useState('query');
  const { toast } = useToast();

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast({
      title: 'Copied',
      description: 'Copied to clipboard',
    });
  };

  if (!data && !isLoading) {
    return null;
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-5xl max-h-[80vh] overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle>Query Inspector</DialogTitle>
        </DialogHeader>

        <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col overflow-hidden">
          <TabsList>
            <TabsTrigger value="query">Query</TabsTrigger>
            <TabsTrigger value="data">Data</TabsTrigger>
            <TabsTrigger value="json">JSON</TabsTrigger>
            <TabsTrigger value="stats">Stats</TabsTrigger>
          </TabsList>

          {/* Query Tab */}
          <TabsContent value="query" className="flex-1 overflow-auto space-y-4">
            {isLoading ? (
              <LoadingIndicator className="h-20" />
            ) : (
              <>
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="font-medium text-sm">Original Query</h3>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => copyToClipboard(data?.inspect?.rawQuery || '')}
                    >
                      <Copy className="size-4" />
                    </Button>
                  </div>
                  <pre className="bg-muted p-4 rounded-md text-sm overflow-auto">
                    {data?.inspect?.rawQuery || 'N/A'}
                  </pre>
                </div>

                <div>
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="font-medium text-sm">Executed Query (with variables)</h3>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => copyToClipboard(data?.inspect?.executedQuery || '')}
                    >
                      <Copy className="size-4" />
                    </Button>
                  </div>
                  <pre className="bg-muted p-4 rounded-md text-sm overflow-auto">
                    {data?.inspect?.executedQuery || 'N/A'}
                  </pre>
                </div>

                <div>
                  <h3 className="font-medium text-sm mb-2">Variables</h3>
                  <pre className="bg-muted p-4 rounded-md text-sm overflow-auto">
                    {JSON.stringify(data?.inspect?.variables || {}, null, 2)}
                  </pre>
                </div>
              </>
            )}
          </TabsContent>

          {/* Data Tab */}
          <TabsContent value="data" className="flex-1 overflow-auto">
            {isLoading ? (
              <LoadingIndicator className="h-20" />
            ) : (
              <div className="overflow-auto">
                {data?.data && data.data.length > 0 ? (
                  <table className="w-full border-collapse text-sm">
                    <thead>
                      <tr className="border-b">
                        {data.meta?.map((col) => (
                          <th key={col.name} className="text-left p-2 font-medium">
                            {col.name}
                          </th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {data.data.map((row, idx) => (
                        <tr key={idx} className="border-b">
                          {data.meta?.map((col) => (
                            <td key={col.name} className="p-2">
                              {String(row[col.name] ?? '')}
                            </td>
                          ))}
                        </tr>
                      ))}
                    </tbody>
                  </table>
                ) : (
                  <div className="text-muted-foreground">No data</div>
                )}
              </div>
            )}
          </TabsContent>

          {/* JSON Tab */}
          <TabsContent value="json" className="flex-1 overflow-auto">
            {isLoading ? (
              <LoadingIndicator className="h-20" />
            ) : (
              <div>
                <div className="flex items-center justify-end mb-2">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => copyToClipboard(JSON.stringify(data, null, 2))}
                  >
                    <Copy className="size-4" />
                  </Button>
                </div>
                <pre className="bg-muted p-4 rounded-md text-sm overflow-auto">
                  {JSON.stringify(data, null, 2)}
                </pre>
              </div>
            )}
          </TabsContent>

          {/* Stats Tab */}
          <TabsContent value="stats" className="flex-1 overflow-auto space-y-2">
            {isLoading ? (
              <LoadingIndicator className="h-20" />
            ) : (
              <>
                <div className="grid grid-cols-2 gap-4">
                  <div className="border rounded-md p-4">
                    <div className="text-sm text-muted-foreground">Elapsed Time</div>
                    <div className="text-2xl font-medium">
                      {data?.statistics?.elapsed ? `${data.statistics.elapsed.toFixed(3)}s` : 'N/A'}
                    </div>
                  </div>

                  <div className="border rounded-md p-4">
                    <div className="text-sm text-muted-foreground">Rows Returned</div>
                    <div className="text-2xl font-medium">{data?.rows ?? 0}</div>
                  </div>
                </div>

                <div className="border rounded-md p-4">
                  <div className="text-sm text-muted-foreground mb-2">Datasource</div>
                  <div className="font-medium">{data?.inspect?.datasourceName || 'N/A'}</div>
                </div>

                <div className="border rounded-md p-4">
                  <div className="text-sm text-muted-foreground mb-2">Timestamp</div>
                  <div className="font-medium">
                    {data?.inspect?.timestamp
                      ? formatDateTime(data.inspect.timestamp)
                      : 'N/A'}
                  </div>
                </div>

                {data?.inspect?.error && (
                  <div className="border border-destructive rounded-md p-4">
                    <div className="text-sm text-destructive mb-2">Error</div>
                    <pre className="text-sm">{data.inspect.error}</pre>
                  </div>
                )}
              </>
            )}
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}
