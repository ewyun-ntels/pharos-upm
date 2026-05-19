import React, {useState} from 'react';
import {Ban, X, Pencil, Trash2, KeyRound} from 'lucide-react';
import {
  Badge,
  Button,
  Separator,
  AlertDialog,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogCancel,
  AlertDialogAction,
} from '@pharos/shared/components/ui';
import {useHasPermission} from '@pharos/shared/features/auth';
import {PermissionKeysSchema} from '@pharos/shared/types/role';
import {formatDateTime, formatRelativeTime} from '@lib/format-date';
import type {User, Role} from '@pharos/shared/types/user';
import {useUserSheet} from '@features/user/hooks/use-user-sheet';
import {sortRoles, getRoleBadgeDisplay} from '@features/user/utils/role-badge';
import {ChangePasswordDialog} from './ChangePasswordDialog';

interface UserDetailSheetProps {
  user: User;
  refetch: () => void;
  onEdit: (user: User) => void;
}

function RoleAttrBadges({roles}: {roles: Role[] | undefined}) {
  if (!roles || roles.length === 0) return <span className="text-muted-foreground text-sm">-</span>;

  const sorted = sortRoles(roles);

  return (
    <div className="flex flex-wrap gap-1">
      {sorted.map((r) => {
        const {isAttr, prefix, displayName} = getRoleBadgeDisplay(r);
        return (
          <Badge key={r.role} variant="secondary" className="text-xs">
            {isAttr ? (
              <span className="text-muted-foreground">Permission:</span>
            ) : prefix ? (
              <span className="text-muted-foreground">{prefix}:</span>
            ) : null}
            <span>{displayName}</span>
          </Badge>
        );
      })}
    </div>
  );
}

export function UserDetailSheet({user, refetch, onEdit}: UserDetailSheetProps) {
  const {
    liveUser,
    setClose,
    isBlocked,
    showDeleteAlert,
    setShowDeleteAlert,
    metaSchema,
    userInfoEntries,
    handleBlockToggle,
    handleDeleteConfirm,
  } = useUserSheet(user, refetch);

  const isSuperAdmin = useHasPermission(PermissionKeysSchema.enum['role:super_admin']);
  const hasUserUpdate = useHasPermission(PermissionKeysSchema.enum['role:user_update']);
  const hasUserDelete = useHasPermission(PermissionKeysSchema.enum['role:user_delete']);

  const isRowSuperAdmin = liveUser.roles?.some((r) => r.role === 'role:super_admin');
  const disableActions = !isSuperAdmin && isRowSuperAdmin;
  const [showPasswordDialog, setShowPasswordDialog] = useState(false);

  return (
    <>
      <div className="flex flex-col h-full w-full">
        {/* Header */}
        <div className="flex items-start justify-between px-4 py-3 border-b shrink-0">
          <div className="flex items-center gap-2 flex-wrap min-w-0">
            <span className="font-semibold text-base truncate">{liveUser.name}</span>
            {isBlocked && (
              <Badge variant="outline" className="text-red-500 border-red-500 text-xs shrink-0">
                Blocked
              </Badge>
            )}
            {liveUser.temporary_blocked_expires_at && !isBlocked && (
              <Badge variant="outline" className="text-orange-500 border-orange-500 text-xs shrink-0">
                Temp Locked
              </Badge>
            )}
          </div>
          <div className="flex items-center gap-1 shrink-0 ml-2">
            {hasUserUpdate && (
              <>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => {
                    setClose();
                    onEdit(liveUser);
                  }}
                  disabled={disableActions}
                  className="h-7 text-xs gap-1"
                >
                  <Pencil className="h-3 w-3" />
                  Edit
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => setShowPasswordDialog(true)}
                  disabled={disableActions}
                  className="h-7 text-xs gap-1"
                >
                  <KeyRound className="h-3 w-3" />
                  Password
                </Button>
                <Button
                  size="sm"
                  variant={isBlocked ? 'default' : 'outline'}
                  onClick={handleBlockToggle}
                  disabled={disableActions}
                  className="h-7 text-xs gap-1"
                >
                  <Ban className="h-3 w-3" />
                  {isBlocked ? 'Unblock' : 'Block'}
                </Button>
              </>
            )}
            <Button size="icon" variant="ghost" className="h-7 w-7" onClick={() => setClose()}>
              <X className="h-4 w-4" />
            </Button>
          </div>
        </div>

        {/* Body */}
        <div className="flex-1 overflow-y-auto px-4 py-4 space-y-4">
          <div className="grid grid-cols-2 gap-3 text-sm">
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground font-medium">Username</p>
              <p className="font-medium">{liveUser.name}</p>
            </div>
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground font-medium">Status</p>
              <div>
                {isBlocked ? (
                  <Badge variant="outline" className="text-red-500 border-red-500 text-xs">Blocked</Badge>
                ) : liveUser.temporary_blocked_expires_at ? (
                  <Badge variant="outline" className="text-orange-500 border-orange-500 text-xs">Temp Locked</Badge>
                ) : (
                  <Badge variant="outline" className="text-green-600 border-green-600 text-xs">Active</Badge>
                )}
              </div>
            </div>
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground font-medium">Created</p>
              <p>{formatDateTime(liveUser.created_at)}</p>
              {liveUser.created_at && (
                <p className="text-xs text-muted-foreground">{formatRelativeTime(liveUser.created_at)}</p>
              )}
            </div>
            <div className="space-y-1">
              <p className="text-xs text-muted-foreground font-medium">Password expiration</p>
              <p className={liveUser.password_expired_at && new Date(liveUser.password_expired_at) < new Date() ? 'text-red-500' : ''}>
                {liveUser.password_expired_at ? formatDateTime(liveUser.password_expired_at) : '-'}
              </p>
              {liveUser.password_expired_at && (
                <p className={`text-xs ${new Date(liveUser.password_expired_at) < new Date() ? 'text-red-400' : 'text-muted-foreground'}`}>
                  {formatRelativeTime(liveUser.password_expired_at)}
                </p>
              )}
            </div>
            {liveUser.temporary_blocked_expires_at && (
              <div className="space-y-1 col-span-2">
                <p className="text-xs text-muted-foreground font-medium">Temp Lock Expires</p>
                <p className="text-orange-500">{formatDateTime(liveUser.temporary_blocked_expires_at)}</p>
              </div>
            )}
          </div>

          <div className="space-y-1">
            <p className="text-xs text-muted-foreground font-medium">Pending actions</p>
            {liveUser.prepare_labels && liveUser.prepare_labels.length > 0 ? (
              <div className="flex flex-wrap gap-1">
                {liveUser.prepare_labels.map((label, i) => (
                  <Badge key={i} variant="outline" className="text-xs text-orange-500 border-orange-500">
                    {label}
                  </Badge>
                ))}
              </div>
            ) : (
              <span className="text-muted-foreground text-sm">-</span>
            )}
          </div>

          <div className="space-y-1">
            <p className="text-xs text-muted-foreground font-medium">Assignments</p>
            <RoleAttrBadges roles={liveUser.roles} />
          </div>

          {userInfoEntries.length > 0 && (
            <>
              <Separator />
              <div className="space-y-2">
                <p className="text-xs text-muted-foreground font-medium">User Fields</p>
                <div className="grid grid-cols-2 gap-3 text-sm">
                  {userInfoEntries.map(([key, value]) => {
                    const label = metaSchema?.properties?.[key]?.title ?? key;
                    return (
                      <div key={key} className="space-y-1">
                        <p className="text-xs text-muted-foreground font-medium">{label}</p>
                        <p className="wrap-break-word">{String(value)}</p>
                      </div>
                    );
                  })}
                </div>
              </div>
            </>
          )}
        </div>

        {/* Footer - Danger Zone */}
        {hasUserDelete && (
          <div className="px-4 py-3 border-t shrink-0">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowDeleteAlert(true)}
              disabled={disableActions}
              className="w-full h-8 text-xs gap-1 text-destructive hover:text-destructive border-destructive/30 hover:border-destructive"
            >
              <Trash2 className="h-3 w-3" />
              Delete User
            </Button>
          </div>
        )}
      </div>

      <ChangePasswordDialog
        open={showPasswordDialog}
        onOpenChange={setShowPasswordDialog}
        username={liveUser.name}
      />

      <AlertDialog open={showDeleteAlert} onOpenChange={setShowDeleteAlert}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Are you sure you want to delete?</AlertDialogTitle>
            <AlertDialogDescription>
              User <strong>{liveUser.name}</strong> will be deleted. This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={handleDeleteConfirm}>
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
