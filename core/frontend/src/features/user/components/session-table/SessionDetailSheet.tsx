import React from 'react';
import {X, LogOut} from 'lucide-react';
import {Badge, Button} from '@pharos/shared/components/ui';
import {useHasPermission} from '@pharos/shared/features/auth';
import {PermissionKeysSchema} from '@pharos/shared/types/role';
import {formatDateTime, formatRelativeTime} from '@lib/format-date';
import {useSheetStore} from '@components/sheet/sheetStore';
import type {Session} from './columns';

interface SessionDetailSheetProps {
  session: Session;
  onRevoke: () => void;
}

export function SessionDetailSheet({session, onRevoke}: SessionDetailSheetProps) {
  const {setClose} = useSheetStore();
  const isExpired = new Date(session.expires_at) < new Date();
  const hasUserUpdate = useHasPermission(PermissionKeysSchema.enum['role:user_update']);

  return (
    <div className="flex flex-col h-full w-full">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b shrink-0">
        <div className="flex items-center gap-2 min-w-0">
          <span className="font-semibold text-base truncate">{session.username}</span>
          {session.blocked ? (
            <Badge variant="destructive" className="text-xs shrink-0">Blocked</Badge>
          ) : (
            <Badge variant="outline" className="text-green-600 border-green-600 text-xs shrink-0">Active</Badge>
          )}
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
            <p className="font-medium">{session.username}</p>
          </div>
          <div className="space-y-1">
            <p className="text-xs text-muted-foreground font-medium">Status</p>
            {session.blocked ? (
              <Badge variant="destructive" className="text-xs">Blocked</Badge>
            ) : (
              <Badge variant="outline" className="text-green-600 border-green-600 text-xs">Active</Badge>
            )}
          </div>
          <div className="space-y-1 col-span-2">
            <p className="text-xs text-muted-foreground font-medium">Client ID</p>
            <p className="font-mono text-xs break-all">{session.client_id}</p>
          </div>
          <div className="space-y-1">
            <p className="text-xs text-muted-foreground font-medium">Created at</p>
            <p>{formatDateTime(session.created_at)}</p>
            <p className="text-xs text-muted-foreground">{formatRelativeTime(session.created_at)}</p>
          </div>
          <div className="space-y-1">
            <p className="text-xs text-muted-foreground font-medium">Expires at</p>
            <p className={isExpired ? 'text-red-500' : ''}>{formatDateTime(session.expires_at)}</p>
            <p className={`text-xs ${isExpired ? 'text-red-400' : 'text-muted-foreground'}`}>
              {formatRelativeTime(session.expires_at)}
            </p>
          </div>
          {session.client_ip && (
            <div className="space-y-1 col-span-2">
              <p className="text-xs text-muted-foreground font-medium">Client IP</p>
              <p className="font-mono text-xs">{session.client_ip}</p>
            </div>
          )}
          {session.user_agent && (
            <div className="space-y-1 col-span-2">
              <p className="text-xs text-muted-foreground font-medium">User agent</p>
              <p className="text-xs break-all text-muted-foreground">{session.user_agent}</p>
            </div>
          )}
        </div>
      </div>

      {/* Footer */}
      {hasUserUpdate && (
        <div className="px-4 py-3 border-t shrink-0">
          <Button
            variant="outline"
            size="sm"
            onClick={onRevoke}
            className="w-full h-8 text-xs gap-1 text-destructive hover:text-destructive border-destructive/30 hover:border-destructive"
          >
            <LogOut className="h-3 w-3" />
            Revoke Session
          </Button>
        </div>
      )}
    </div>
  );
}
