import React, { useMemo, useState, useEffect, useCallback } from 'react';
import { useList, useUpdate, useInvalidate, BaseRecord, HttpError } from '@/lib/data-provider';
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
import { DASHBOARD_PROVIDER_NAME, DASHBOARD_RESOURCES } from '@/providers/dashboard-provider';
import { useToast } from '@hooks/use-toast';
import { getDashboardHistoryColumns, type DashboardHistoryRecord } from './columns';
import { DashboardHistoryViewSheet } from './DashboardHistoryViewSheet';

const DEFAULT_PAGE_SIZE = 20;
const DEBOUNCE_MS = 300;

export function DashboardHistoryTab() {
  const { toast } = useToast();
  const { setSheet } = useSheetStore();
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [searchInput, setSearchInput] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [rollbackTarget, setRollbackTarget] = useState<DashboardHistoryRecord | null>(null);
  const [isRollingBack, setIsRollingBack] = useState(false);
  const invalidate = useInvalidate();

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchInput);
      setCurrentPage(1);
    }, DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [searchInput]);

  const {
    query: { data, isLoading, refetch },
  } = useList<DashboardHistoryRecord>({
    dataProviderName: DASHBOARD_PROVIDER_NAME,
    resource: DASHBOARD_RESOURCES.HISTORY,
    pagination: { currentPage, pageSize, mode: 'server' },
    meta: { query: debouncedSearch ? { search: debouncedSearch } : {} },
    queryOptions: { retry: false },
  });

  const { mutateAsync: rollback } = useUpdate<BaseRecord, HttpError>();

  const tableData = useMemo(() => data?.data ?? [], [data]);
  const total = data?.total ?? 0;

  const handleRollbackConfirm = useCallback(async () => {
    if (!rollbackTarget) return;
    setIsRollingBack(true);
    try {
      await rollback({
        resource: DASHBOARD_RESOURCES.DASHBOARD,
        id: rollbackTarget.dashboardId,
        values: JSON.parse(rollbackTarget.configJson!),
        dataProviderName: DASHBOARD_PROVIDER_NAME,
      });
      toast({ description: 'Dashboard rolled back successfully.' });
      setRollbackTarget(null);
      refetch();
      invalidate({
        resource: DASHBOARD_RESOURCES.DASHBOARD,
        dataProviderName: DASHBOARD_PROVIDER_NAME,
        invalidates: ['list'],
      });
    } catch {
      toast({ description: 'Failed to rollback dashboard.', variant: 'destructive' });
    } finally {
      setIsRollingBack(false);
    }
  }, [rollbackTarget, rollback, toast, refetch, invalidate]);

  const handleView = useCallback(
    (record: DashboardHistoryRecord) => {
      setSheet(<DashboardHistoryViewSheet record={record} onRollback={setRollbackTarget} />);
    },
    [setSheet, setRollbackTarget],
  );

  const columns = useMemo(
    () => getDashboardHistoryColumns({ onRollback: setRollbackTarget }),
    [],
  );

  return (
    <>
      <div className="px-5 h-full">
        <DataGrid<DashboardHistoryRecord>
          tableKey="dashboard-history"
          data={tableData}
          columns={columns}
          useTableSorting={true}
          rowCursor={true}
          isLoading={isLoading}
          onRowClick={handleView}
          excludeRowClickColumns={['rollback']}
          totalCount={total}
          paginationConfig={{ pageSize, pageIndex: currentPage - 1 }}
          onPageChange={(page) => setCurrentPage(page)}
          onPageSizeChange={(size) => {
            setPageSize(size);
            setCurrentPage(1);
          }}
          leftFilters={() => [
            <SearchInput
              key="search"
              size="hsmall"
              placeholder="Dashboard, Action, Changed By..."
              searchValue={searchInput}
              onSearchChange={(v) => setSearchInput(v)}
            />,
          ]}
        />
      </div>

      <AlertDialog open={!!rollbackTarget} onOpenChange={(v) => !v && setRollbackTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Rollback to this version?</AlertDialogTitle>
            <AlertDialogDescription>
              The dashboard &quot;{rollbackTarget?.dashboardTitle}&quot; will be restored to this
              version. A new history entry will be created.
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
