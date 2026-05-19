import React, { useMemo, useState, useEffect, useCallback } from 'react';
import { useList, useOne } from '@/lib/data-provider';
import { VariableInput } from '@pharos/shared/components/ui-extension';
import { DataGrid } from '@pharos/shared/components/ui-extension/data-grid/DataGrid';
import { USER_PROVIDER_NAME, USER_RESOURCES } from '@providers/user-provider';
import { useSheetStore } from '@components/sheet/sheetStore';
import { buildHistoryColumns, type LoginHistoryRecord } from './columns';
import { HistoryDetailSheet } from './HistoryDetailSheet';

const DEFAULT_PAGE_SIZE = 10;
const DEBOUNCE_MS = 300;

export function HistoryTab() {
  const { setSheet } = useSheetStore();
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [inputValue, setInputValue] = useState('');
  const [debouncedUsername, setDebouncedUsername] = useState('');

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedUsername(inputValue);
      setCurrentPage(1);
    }, DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [inputValue]);

  const {
    query: { data, isLoading },
  } = useList<LoginHistoryRecord>({
    dataProviderName: USER_PROVIDER_NAME,
    resource: USER_RESOURCES.AUTH_HISTORY,
    pagination: { currentPage, pageSize, mode: 'server' },
    meta: { query: debouncedUsername ? { username: debouncedUsername } : {} },
    queryOptions: { retry: false },
  });

  const { query: { data: authConfigData } } = useOne({
    dataProviderName: USER_PROVIDER_NAME,
    resource: 'auth',
    id: 'config',
    queryOptions: { retry: false },
  });
  const collectClientInfo: boolean = authConfigData?.data?.common?.collect_client_info ?? false;
  const historyColumns = useMemo(() => buildHistoryColumns(collectClientInfo), [collectClientInfo]);

  const tableData = useMemo(() => data?.data ?? [], [data]);
  const total = data?.total ?? 0;

  const openHistorySheet = useCallback(
    (record: LoginHistoryRecord) => {
      setSheet(<HistoryDetailSheet record={record} />);
    },
    [setSheet],
  );

  return (
    <div className="px-5 h-full">
      <DataGrid<LoginHistoryRecord>
        tableKey="user-history"
        data={tableData}
        columns={historyColumns}
        useTableSorting={true}
        onRowClick={openHistorySheet}
        rowCursor={true}
        isLoading={isLoading}
        // tableHeight="flex"
        totalCount={total}
        paginationConfig={{ pageSize, pageIndex: currentPage - 1 }}
        onPageChange={(page) => setCurrentPage(page)}
        onPageSizeChange={(size) => {
          setPageSize(size);
          setCurrentPage(1);
        }}
        leftFilters={() => [
          <VariableInput
            key="username-filter"
            label="Username"
            placeholder="Search..."
            value={inputValue}
            onChange={(v) => setInputValue(v)}
          />,
        ]}
      />
    </div>
  );
}
