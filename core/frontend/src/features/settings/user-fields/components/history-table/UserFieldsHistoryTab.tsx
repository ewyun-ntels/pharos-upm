import React, { useCallback, useMemo, useState } from 'react';
import { useList, useCreate, useInvalidate, HttpError, BaseRecord } from '@/lib/data-provider';
import { DataGrid } from '@pharos/shared/components/ui-extension/data-grid/DataGrid';
import { SearchInput } from '@pharos/shared/components/ui-extension';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@pharos/shared/components/ui';
import { useSheetStore } from '@components/sheet/sheetStore';
import { USER_PROVIDER_NAME, USER_RESOURCES } from '@providers/user-provider';
import { useToast } from '@hooks/use-toast';
import {
  getUserFieldsHistoryColumns,
  type UserMetadataConfigHistoryEntry,
} from './columns';
import { UserFieldsHistoryViewSheet } from './UserFieldsHistoryViewSheet';

export function UserFieldsHistoryTab() {
  const { toast } = useToast();
  const { setSheet } = useSheetStore();
  const [rollbackTarget, setRollbackTarget] = useState<string | null>(null);
  const [isRollingBack, setIsRollingBack] = useState(false);
  const invalidate = useInvalidate();

  const {
    query: { data, isLoading, refetch },
  } = useList<UserMetadataConfigHistoryEntry>({
    resource: USER_RESOURCES.USER_METADATA_CONFIG_HISTORY,
    dataProviderName: USER_PROVIDER_NAME,
    pagination: { pageSize: 50 },
    queryOptions: { retry: false },
  });

  const { mutateAsync: rollback } = useCreate<BaseRecord, HttpError, { id: string }>();

  const historyList: UserMetadataConfigHistoryEntry[] = useMemo(
    () => (data?.data as UserMetadataConfigHistoryEntry[]) ?? [],
    [data],
  );

  const handleRollbackConfirm = useCallback(async () => {
    if (!rollbackTarget) return;
    setIsRollingBack(true);
    try {
      await rollback({
        resource: USER_RESOURCES.USER_METADATA_CONFIG_ROLLBACK,
        values: { id: rollbackTarget },
        dataProviderName: USER_PROVIDER_NAME,
      });
      toast({ description: 'Configuration rolled back successfully.' });
      setRollbackTarget(null);
      refetch();
      invalidate({
        resource: USER_RESOURCES.USER_METADATA_CONFIG,
        dataProviderName: USER_PROVIDER_NAME,
        invalidates: ['detail'],
      });
    } catch {
      toast({ description: 'Failed to rollback configuration.', variant: 'destructive' });
    } finally {
      setIsRollingBack(false);
    }
  }, [rollbackTarget, rollback, toast, refetch, invalidate]);

  const handleView = useCallback(
    (record: UserMetadataConfigHistoryEntry) => {
      setSheet(<UserFieldsHistoryViewSheet record={record} />);
    },
    [setSheet],
  );

  const columns = useMemo(
    () => getUserFieldsHistoryColumns({ onRollback: setRollbackTarget }),
    [],
  );

  return (
    <>
      <div className="px-5 h-full">
        <DataGrid<UserMetadataConfigHistoryEntry>
          tableKey="user-fields-config-history"
          data={historyList}
          columns={columns}
          useTableSorting={false}
          searchableColumns={['changed-by', 'description']}
          rowCursor={true}
          isLoading={isLoading}
          onRowClick={handleView}
          excludeRowClickColumns={['rollback']}
          leftFilters={(table) => [
            <SearchInput
              key="search"
              size="hsmall"
              placeholder="Changed By, Description..."
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
