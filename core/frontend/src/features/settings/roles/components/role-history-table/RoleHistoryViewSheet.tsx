import React, {useCallback} from 'react';
import {Copy, X} from 'lucide-react';
import {Button} from '@pharos/shared/components/ui';
import {useSheetStore} from '@components/sheet/sheetStore';
import {useToast} from '@hooks/use-toast';
import {formatDateTime} from '@lib/format-date';
import {JsonEditor} from '@components/editor/JsonEditor';
import type {RoleMetadataConfigEntity} from '@pharos/shared/types/role';

interface RoleHistoryViewSheetProps {
  record: RoleMetadataConfigEntity;
}

export function RoleHistoryViewSheet({record}: RoleHistoryViewSheetProps) {
  const {setClose} = useSheetStore();
  const {toast} = useToast();

  const formattedJson = JSON.stringify(record.config, null, 2);

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(formattedJson);
    toast({description: 'Copied to clipboard.'});
  }, [formattedJson, toast]);

  return (
    <div className="flex flex-col h-full w-full">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b shrink-0">
        <div className="flex flex-col min-w-0">
          <span className="font-semibold text-base">Role Configuration</span>
          <span className="text-xs text-muted-foreground">
            {formatDateTime(record.created_at)}
            {record.updated_by ? ` · ${record.updated_by}` : ''}
          </span>
        </div>
        <div className="flex items-center gap-1 ml-2 shrink-0">
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
