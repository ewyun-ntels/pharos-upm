import React from 'react';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  Button,
} from '@pharos/shared/components/ui';
import {LoadingIndicator} from '@pharos/shared/components/ui-extension';
import {AlertCircle, RefreshCw} from 'lucide-react';
import {cn} from '@lib/utils';
import {useQueryInspector} from './use-query-inspector';
import {InspectorContent} from './InspectorContent';

interface QueryInspectorSheetProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  dashboardId: string;
  datasourceName?: string;
  query?: string;
  templateStyle?: string;
  variables?: Record<string, unknown>;
}

export function QueryInspectorSheet({
  open,
  onOpenChange,
  dashboardId,
  datasourceName,
  query,
  templateStyle,
  variables,
}: QueryInspectorSheetProps) {
  const {
    query: {data, isLoading, refetch, error},
  } = useQueryInspector({
    dashboardId,
    datasourceName,
    query,
    templateStyle,
    variables,
    enabled: open && !!datasourceName && !!query && !!dashboardId,
  });

  const inspectData = data?.data;
  // 쿼리 실행은 됐지만 SQL 에러가 발생한 경우 (백엔드가 inspect 정보를 같이 반환)
  const errorWithInspect = error?.response?.data?.inspect ? error.response.data : null;

  const renderContent = () => {
    if (isLoading) {
      return <LoadingIndicator className="h-40" />;
    }

    if (errorWithInspect) {
      return (
        <InspectorContent
          inspect={errorWithInspect.inspect}
          statistics={errorWithInspect.statistics}
          rows={errorWithInspect.rows}
        />
      );
    }

    if (error) {
      return (
        <div className="border border-destructive rounded-lg p-6 bg-destructive/5">
          <div className="flex items-start gap-3">
            <AlertCircle className="h-6 w-6 text-destructive mt-0.5 shrink-0" />
            <div className="flex-1 min-w-0">
              <h3 className="font-semibold text-destructive mb-2">Inspector Request Failed</h3>
              <p className="text-sm text-muted-foreground mb-3">
                Unable to fetch inspector data from the backend.
              </p>
              <pre className="bg-muted p-4 rounded-md text-xs overflow-auto max-h-48 whitespace-pre-wrap">
                {error.message || JSON.stringify(error, null, 2)}
              </pre>
            </div>
          </div>
        </div>
      );
    }

    if (inspectData?.inspect) {
      return (
        <InspectorContent
          inspect={inspectData.inspect}
          statistics={inspectData.statistics}
          rows={inspectData.rows}
        />
      );
    }

    return (
      <div className="flex items-center justify-center h-40 text-muted-foreground text-sm">
        No data yet. Set datasource and query, then click Refresh.
      </div>
    );
  };

  const canRefresh = !!datasourceName && !!query && !!dashboardId;

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-[600px]! sm:max-w-[600px]! flex flex-col p-0 gap-0">
        <SheetHeader className="px-6 pt-6 pb-4 border-b shrink-0">
          <SheetTitle>Query Inspector</SheetTitle>
          <SheetDescription>Query execution details and variable substitution</SheetDescription>
        </SheetHeader>

        <div className="flex-1 overflow-y-auto px-6 py-4 space-y-4">
          <div className="flex justify-end">
            <Button
              variant="outline"
              size="sm"
              onClick={() => refetch?.()}
              disabled={isLoading || !canRefresh}
            >
              <RefreshCw className={cn('h-4 w-4 mr-2', isLoading && 'animate-spin')} />
              Refresh
            </Button>
          </div>
          {renderContent()}
        </div>
      </SheetContent>
    </Sheet>
  );
}
