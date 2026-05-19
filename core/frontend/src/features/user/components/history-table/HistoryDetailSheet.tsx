import React from 'react';
import {X} from 'lucide-react';
import {Badge, Button} from '@pharos/shared/components/ui';
import {formatDateTime, formatRelativeTime} from '@lib/format-date';
import {useSheetStore} from '@components/sheet/sheetStore';
import {UAParser} from 'ua-parser-js';
import type {LoginHistoryRecord} from './columns';

interface HistoryDetailSheetProps {
  record: LoginHistoryRecord;
}

function parseUA(ua: string) {
  if (!ua) return null;

  return new UAParser(ua).getResult();
}

export function HistoryDetailSheet({record}: HistoryDetailSheetProps) {
  const {setClose} = useSheetStore();
  const ua = parseUA(record.user_agent);

  const endReasonBadge = () => {
    const r = record.end_reason;
    if (!record.ended_at && !r) {
      return <Badge variant="outline" className="text-green-600 border-green-600 text-xs">Active</Badge>;
    }
    if (r === 'logout') return <Badge variant="outline" className="text-orange-500 border-orange-500 text-xs">Logout</Badge>;
    if (r === 'expired') return <Badge variant="outline" className="text-muted-foreground text-xs">Expired</Badge>;
    if (r === 'revoked') return <Badge variant="outline" className="text-blue-500 border-blue-500 text-xs">Revoked</Badge>;
    if (r.startsWith('failed:')) return <Badge variant="outline" className="text-red-500 border-red-500 text-xs">Failed</Badge>;
    return <Badge variant="outline" className="text-xs">{r}</Badge>;
  };

  const failedDescription = record.end_reason.startsWith('failed:')
    ? record.end_reason.slice('failed:'.length).trim()
    : null;

  return (
    <div className="flex flex-col h-full w-full">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b shrink-0">
        <div className="flex items-center gap-2 min-w-0">
          <span className="font-semibold text-base truncate">{record.user_id}</span>
          {endReasonBadge()}
        </div>
        <Button size="icon" variant="ghost" className="h-7 w-7 shrink-0 ml-2" onClick={() => setClose()}>
          <X className="h-4 w-4" />
        </Button>
      </div>

      {/* Body */}
      <div className="flex-1 overflow-y-auto px-4 py-4">
        <div className="grid grid-cols-2 gap-3 text-sm">
          <div className="space-y-1">
            <p className="text-xs text-muted-foreground font-medium">Username</p>
            <p className="font-medium">{record.user_id}</p>
          </div>
          <div className="space-y-1">
            <p className="text-xs text-muted-foreground font-medium">Event</p>
            {endReasonBadge()}
          </div>

          {failedDescription && (
            <div className="space-y-1 col-span-2">
              <p className="text-xs text-muted-foreground font-medium">Failure reason</p>
              <p className="text-sm text-red-500">{failedDescription}</p>
            </div>
          )}

          {record.end_reason === 'revoked' && (
            <div className="space-y-1 col-span-2">
              <p className="text-xs text-muted-foreground font-medium">Revoke reason</p>
              <p className="text-sm text-blue-500">Revoked by administrator</p>
            </div>
          )}

          <div className="space-y-1">
            <p className="text-xs text-muted-foreground font-medium">Started at</p>
            <p>{formatDateTime(record.started_at)}</p>
            <p className="text-xs text-muted-foreground">{formatRelativeTime(record.started_at)}</p>
          </div>
          <div className="space-y-1">
            <p className="text-xs text-muted-foreground font-medium">Ended at</p>
            {record.ended_at ? (
              <>
                <p>{formatDateTime(record.ended_at)}</p>
                <p className="text-xs text-muted-foreground">{formatRelativeTime(record.ended_at)}</p>
              </>
            ) : (
              <span className="text-muted-foreground">-</span>
            )}
          </div>

          {record.client_ip && (
            <div className="space-y-1 col-span-2">
              <p className="text-xs text-muted-foreground font-medium">Client IP</p>
              <p className="font-mono text-xs">{record.client_ip}</p>
            </div>
          )}

          {ua && (
            <>
              {ua.browser.name && (
                <div className="space-y-1">
                  <p className="text-xs text-muted-foreground font-medium">Browser</p>
                  <p className="text-sm">{ua.browser.name} {ua.browser.version}</p>
                </div>
              )}
              {ua.os.name && (
                <div className="space-y-1">
                  <p className="text-xs text-muted-foreground font-medium">OS</p>
                  <p className="text-sm">{ua.os.name} {ua.os.version}</p>
                </div>
              )}
              {ua.device.type && (
                <div className="space-y-1">
                  <p className="text-xs text-muted-foreground font-medium">Device</p>
                  <p className="text-sm capitalize">{ua.device.type}</p>
                </div>
              )}
              {record.user_agent && (
                <div className="space-y-1 col-span-2">
                  <p className="text-xs text-muted-foreground font-medium">User agent</p>
                  <p className="text-xs break-all text-muted-foreground">{record.user_agent}</p>
                </div>
              )}
            </>
          )}

          <div className="space-y-1 col-span-2">
            <p className="text-xs text-muted-foreground font-medium">Session ID</p>
            <p className="font-mono text-xs break-all text-muted-foreground">{record.session_id}</p>
          </div>
        </div>
      </div>
    </div>
  );
}
