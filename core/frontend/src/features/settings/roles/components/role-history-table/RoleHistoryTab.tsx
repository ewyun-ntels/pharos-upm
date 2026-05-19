import React, { useMemo, useCallback, useState } from 'react';
import { useOne, useCreate, HttpError, BaseRecord } from '@/lib/data-provider';
import { DataGrid } from '@pharos/shared/components/ui-extension/data-grid/DataGrid';
import { SearchInput } from '@pharos/shared/components/ui-extension';
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
import { useSheetStore } from '@components/sheet/sheetStore';
import { ROLE_PROVIDER_NAME, ROLE_RESOURCES } from '@providers/role-provider';
import { useToast } from '@hooks/use-toast';
import { getRoleHistoryColumns } from './columns';
import { RoleHistoryViewSheet } from './RoleHistoryViewSheet';
import type { RoleMetadataConfigEntity } from '@pharos/shared/types/role';

export function RoleHistoryTab() {
  const { toast } = useToast();
  const { setSheet } = useSheetStore();
  const [rollbackTarget, setRollbackTarget] = useState<string | null>(null);
  const [isRollingBack, setIsRollingBack] = useState(false);

  const {
    query: { data, isLoading, refetch },
  } = useOne<RoleMetadataConfigEntity[]>({
    dataProviderName: ROLE_PROVIDER_NAME,
    resource: ROLE_RESOURCES.CONFIG_HISTORY,
    id: 'history',
    queryOptions: { retry: false, staleTime: 0 },
  });

  const { mutateAsync: rollback } = useCreate<BaseRecord, HttpError, { id: string }>();

  const historyList: RoleMetadataConfigEntity[] = useMemo(() => {
    const d = data?.data as unknown;
    return Array.isArray(d) ? (d as RoleMetadataConfigEntity[]) : [];
  }, [data]);

  const handleRollbackConfirm = useCallback(async () => {
    if (!rollbackTarget) return;
    setIsRollingBack(true);
    try {
      await rollback({
        resource: ROLE_RESOURCES.CONFIG_ROLLBACK,
        values: { id: rollbackTarget },
        dataProviderName: ROLE_PROVIDER_NAME,
      });
      toast({ description: 'Configuration rolled back successfully.' });
      setRollbackTarget(null);
      refetch();
    } catch {
      toast({ description: 'Failed to rollback configuration.', variant: 'destructive' });
    } finally {
      setIsRollingBack(false);
    }
  }, [rollbackTarget, rollback, toast, refetch]);

  const handleView = useCallback(
    (record: RoleMetadataConfigEntity) => {
      setSheet(<RoleHistoryViewSheet record={record} />);
    },
    [setSheet],
  );

  const columns = useMemo(
    () => getRoleHistoryColumns({ onRollback: setRollbackTarget }),
    [],
  );

  return (
    <>
      <div className="px-5 h-full">
        <DataGrid<RoleMetadataConfigEntity>
          tableKey="role-config-history"
          data={historyList}
          columns={columns}
          useTableSorting={false}
          searchableColumns={['changed-by']}
          rowCursor={true}
          isLoading={isLoading}
          onRowClick={handleView}
          excludeRowClickColumns={['rollback']}
          leftFilters={(table) => [
            <SearchInput
              key="search"
              size="hsmall"
              placeholder="Changed By..."
              autoComplete="off"
              table={table}
            />,
          ]}
        />
      </div>

      <AlertDialog open={!!rollbackTarget} onOpenChange={(v) => !v && setRollbackTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Rollback to this version?</AlertDialogTitle>
            <AlertDialogDescription>
              The current configuration will be replaced with the selected version. A new history
              entry will be created.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleRollbackConfirm} disabled={isRollingBack}>
              {isRollingBack ? 'Rolling back...' : 'Rollback'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
