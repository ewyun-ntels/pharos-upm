import React, {useCallback, useMemo} from 'react';
import {Copy, RotateCcw, X} from 'lucide-react';
import {Button} from '@pharos/shared/components/ui';
import {useSheetStore} from '@components/sheet/sheetStore';
import {useToast} from '@hooks/use-toast';
import {formatDateTime} from '@lib/format-date';
import {JsonEditor} from '@components/editor/JsonEditor';
import type {DashboardHistoryRecord} from './columns';

interface DashboardHistoryViewSheetProps {
  record: DashboardHistoryRecord;
  onRollback?: (record: DashboardHistoryRecord) => void;
}

export function DashboardHistoryViewSheet({record, onRollback}: DashboardHistoryViewSheetProps) {
  const {setClose} = useSheetStore();
  const {toast} = useToast();

  const formattedJson = record.configJson
    ? JSON.stringify(JSON.parse(record.configJson), null, 2)
    : '';

  const parsedConfig = useMemo(() => {
    try {
      return record.configJson ? JSON.parse(record.configJson) : null;
    } catch {
      return null;
    }
  }, [record.configJson]);

  const displayName: string | undefined = parsedConfig?.displayName;
  const description: string | undefined = parsedConfig?.description;

  const handleRollback = useCallback(() => {
    setClose();
    onRollback?.(record);
  }, [record, onRollback, setClose]);

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(formattedJson);
    toast({description: 'Copied to clipboard.'});
  }, [formattedJson, toast]);

  return (
    <div className="flex flex-col h-full w-full">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b shrink-0">
        <div className="flex flex-col min-w-0 gap-0.5">
          <span className="font-semibold text-base truncate">
            {displayName || record.dashboardTitle}
          </span>
          {displayName && (
            <span className="text-xs text-muted-foreground truncate">{record.dashboardTitle}</span>
          )}
          {description && (
            <span className="text-xs text-muted-foreground truncate">{description}</span>
          )}
          <span className="text-xs text-muted-foreground">
            {formatDateTime(record.changedAt)}
            {record.changedBy ? ` · ${record.changedBy}` : ''}
          </span>
        </div>
        <div className="flex items-center gap-1 ml-2 shrink-0">
          {onRollback && (
            <Button size="icon" variant="outline" className="h-7 w-7" title="Rollback" onClick={handleRollback}>
              <RotateCcw className="h-4 w-4" />
            </Button>
          )}
          <Button size="icon" variant="outline" className="h-7 w-7" onClick={handleCopy}>
            <Copy className="h-4 w-4" />
          </Button>
          <Button size="icon" variant="ghost" className="h-7 w-7" onClick={setClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Body */}
      <div className="flex-1 overflow-hidden">
        <JsonEditor setValue={formattedJson} readOnly height="100%" />
      </div>
    </div>
  );
}
