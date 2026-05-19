import React, { useMemo, useCallback, useState } from 'react';
import { useList, useDelete, useOne } from '@/lib/data-provider';
import { VariableInput, IconButton } from '@pharos/shared/components/ui-extension';
import { DataGrid } from '@pharos/shared/components/ui-extension/data-grid/DataGrid';
import { RefreshCcw, toast } from '@pharos/shared/components';
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogCancel,
  AlertDialogAction,
} from '@pharos/shared/components/ui';
import { USER_PROVIDER_NAME, USER_RESOURCES } from '@providers/user-provider';
import { useSheetStore } from '@components/sheet/sheetStore';
import { buildSessionColumns, type Session } from './columns';
import { SessionDetailSheet } from './SessionDetailSheet';

export function SessionsTab() {
  const { setSheet } = useSheetStore();
  const [revokeTarget, setRevokeTarget] = useState<Session | null>(null);

  const { query: { data: authConfigData } } = useOne({
    dataProviderName: USER_PROVIDER_NAME,
    resource: 'auth',
    id: 'config',
    queryOptions: { retry: false },
  });
  const collectClientInfo: boolean = authConfigData?.data?.common?.collect_client_info ?? false;
  const sessionColumns = useMemo(() => buildSessionColumns(collectClientInfo), [collectClientInfo]);

  const {
    query: { data, isLoading, refetch },
  } = useList<Session>({
    dataProviderName: USER_PROVIDER_NAME,
    resource: USER_RESOURCES.AUTH_SESSIONS,
    queryOptions: { retry: false },
  });

  const { mutateAsync: revokeSession } = useDelete();

  const tableData = useMemo(() => data?.data ?? [], [data]);

  const openSessionSheet = useCallback(
    (session: Session) => {
      setSheet(
        <SessionDetailSheet
          session={session}
          onRevoke={() => setRevokeTarget(session)}
        />,
      );
    },
    [setSheet],
  );

  const handleRevokeConfirm = async () => {
    if (!revokeTarget) return;
    try {
      await revokeSession({
        dataProviderName: USER_PROVIDER_NAME,
        resource: USER_RESOURCES.AUTH_SESSIONS,
        id: revokeTarget.username,
      });
      toast.success(`Session for ${revokeTarget.username} has been revoked.`);
      refetch();
    } catch {
      toast.error('Failed to revoke session.');
    } finally {
      setRevokeTarget(null);
    }
  };

  return (
    <div className="px-5">
      <DataGrid<Session>
        tableKey="user-sessions"
        data={tableData}
        columns={sessionColumns}
        useTableSorting={true}
        onRowClick={openSessionSheet}
        rowCursor={true}
        isLoading={isLoading}
        rightFilters={() => [
          <IconButton
            key="refresh"
            type="button"
            variant="outline"
            icon={<RefreshCcw />}
            onClick={() => refetch()}
          />,
        ]}
        leftFilters={(table) => [
          <VariableInput
            key="username-filter"
            label="Username"
            placeholder="Search..."
            onChange={(v) => {
              table.getColumn('username')?.setFilterValue(v || undefined);
            }}
          />,
        ]}
      />

      <AlertDialog open={!!revokeTarget} onOpenChange={(open) => !open && setRevokeTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Revoke session?</AlertDialogTitle>
            <AlertDialogDescription>
              The session for <strong>{revokeTarget?.username}</strong> will be terminated immediately. The user will need to log in again.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={handleRevokeConfirm}>
              Revoke
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
