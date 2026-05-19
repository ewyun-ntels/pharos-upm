'use client';

import {useMemo, useState, useEffect, useCallback} from 'react';
import {useList} from '@lib/data-provider';
import {ColumnDef, Table} from '@tanstack/react-table';
import {DataGrid} from '@pharos/shared/components/ui-extension/data-grid';
import {SearchInput} from '@pharos/shared/components/ui-extension';
import {useSheetStore} from '@components/sheet/sheetStore';
import {
  CATV_PROVIDER_NAME,
  CATV_RESOURCES,
  ScheduleResultDetailRecord,
} from '../../../../providers';
import {
  BASIC_COLUMN_DEFS,
  BASIC_COLUMN_IDS,
  DEFAULT_HIDDEN_BASIC_COLUMN_IDS,
  DEFAULT_HIDDEN_BASIC_COLUMN_IDS_BY_WORK_TYPE,
  PROGRESS_STATUS_CONFIG,
  StbInfoRow,
} from './columns';
import {ColumnVisibilityPanel} from './ColumnVisibilityPanel';
import {ScheduleResultDetailSheet} from './ScheduleResultDetailSheet';
import {POLL_INTERVAL_MS, SEARCH_DEBOUNCE_MS} from '../constants';
import {STATUS_RUNNING, STATUS_PENDING} from '../constants';

const PROGRESS_STATUS_KOREAN_MAP: Record<string, string> = Object.fromEntries(
  Object.entries(PROGRESS_STATUS_CONFIG).map(([key, {label}]) => [label, key]),
);

function buildFilters(search: string) {
  if (!search) return [];
  const trimmed = search.trim();
  // 한글 입력 → 영문 상태값 변환
  const resolvedFromKorean = PROGRESS_STATUS_KOREAN_MAP[trimmed];
  if (resolvedFromKorean) {
    return [{field: 'progress_status', operator: 'eq', value: resolvedFromKorean}];
  }
  // 영문 상태값 직접 입력 (pending/running/succeeded/failed)
  const validStatuses = Object.keys(PROGRESS_STATUS_CONFIG);
  if (validStatuses.includes(trimmed)) {
    return [{field: 'progress_status', operator: 'eq', value: trimmed}];
  }
  return [{field: 'search', operator: 'eq', value: trimmed}];
}

interface ScheduleResultDetailProps {
  resultId: string;
  workType: string;
}

export function ScheduleResultDetail({resultId, workType}: ScheduleResultDetailProps) {
  const {setSheet} = useSheetStore();
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(100);
  const [searchInput, setSearchInput] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');

  // 검색어 디바운스
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchInput.trim());
      setCurrentPage(1);
    }, SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [searchInput]);

  const listQuery = useList<ScheduleResultDetailRecord>({
    resource: CATV_RESOURCES.SCHEDULE_RESULT_DETAILS,
    dataProviderName: CATV_PROVIDER_NAME,
    meta: {
      resultId,
    },
    filters: buildFilters(debouncedSearch),
    pagination: {
      currentPage,
      pageSize,
      mode: 'server' as const,
    },
    queryOptions: {
      enabled: !!resultId,
    },
  });

  const isLoading = listQuery.query.isLoading;
  const isFetching = listQuery.query.isFetching;
  const data = listQuery.query.data;
  const refetch = listQuery.query.refetch;

  const isStbRequestInfo = workType === 'stb_request_info';

  // stb_request_info: result_message JSON 파싱
  const stbInfoRows = useMemo<StbInfoRow[]>(() => {
    if (!isStbRequestInfo) return [];
    return (data?.data || []).map((record) => ({
      ...record,
      parsedInfo: (() => {
        try {
          return JSON.parse(record.result_message || '{}');
        } catch {
          return {};
        }
      })(),
    }));
  }, [data, isStbRequestInfo]);

  // 모든 rows에서 JSON 키 목록 추출 (순서 보존)
  const allInfoKeys = useMemo(() => {
    const seen = new Set<string>();
    const keys: string[] = [];
    stbInfoRows.forEach((row) => {
      Object.keys(row.parsedInfo).forEach((k) => {
        if (!seen.has(k)) {
          seen.add(k);
          keys.push(k);
        }
      });
    });
    return keys;
  }, [stbInfoRows]);

  const dynamicColumns: ColumnDef<StbInfoRow>[] = allInfoKeys.map((key) => ({
    id: key,
    header: key,
    size: 130,
    enableHiding: true,
    accessorFn: (row) => row.parsedInfo[key] ?? '-',
    cell: ({getValue}) => {
      const value = getValue<string>();
      return <span className="font-mono text-xs">{value}</span>;
    },
  }));

  const columns = isStbRequestInfo
    ? ([...BASIC_COLUMN_DEFS, ...dynamicColumns] as ColumnDef<ScheduleResultDetailRecord>[])
    : BASIC_COLUMN_DEFS;

  const tableData = isStbRequestInfo
    ? (stbInfoRows as ScheduleResultDetailRecord[])
    : data?.data || [];

  // 워크타입별 기본 숨김 컬럼 설정
  const initialVisibility = useMemo(() => {
    const hiddenIds =
      DEFAULT_HIDDEN_BASIC_COLUMN_IDS_BY_WORK_TYPE[workType] ?? DEFAULT_HIDDEN_BASIC_COLUMN_IDS;
    return Object.fromEntries(hiddenIds.map((id) => [id, false]));
  }, [workType]);

  // active record(running/pending) 있을 때만 폴링
  const hasActiveRecords = useMemo(
    () =>
      (data?.data || []).some(
        (record) =>
          record.progress_status === STATUS_RUNNING || record.progress_status === STATUS_PENDING,
      ),
    [data],
  );

  useEffect(() => {
    if (!hasActiveRecords) return;
    const interval = setInterval(() => {
      refetch();
    }, POLL_INTERVAL_MS);
    return () => clearInterval(interval);
  }, [hasActiveRecords, refetch]);

  const openDetailSheet = useCallback(
    (record: ScheduleResultDetailRecord) => {
      setSheet(<ScheduleResultDetailSheet record={record} />);
    },
    [setSheet],
  );

  const handleSearchChange = useCallback((v: string) => setSearchInput(v), []);
  const renderLeftFilters = useCallback(
    () => [
      <SearchInput
        key="search"
        size="hsmall"
        placeholder="MAC, K8s Job, 결과 값..."
        searchValue={searchInput}
        onSearchChange={handleSearchChange}
      />,
    ],
    [handleSearchChange],
  );
  const renderRightFilters = useCallback(
    (table: Table<ScheduleResultDetailRecord>) => {
      const items = [];
      if (hasActiveRecords) {
        items.push(
          <span key="polling" className="flex items-center gap-1.5 text-xs text-muted-foreground">
            <span
              className={`w-1.5 h-1.5 rounded-full bg-green-500 ${isFetching ? 'animate-pulse' : ''}`}
            />
            실시간 갱신 중
          </span>,
        );
      }
      items.push(
        <ColumnVisibilityPanel
          key="view"
          table={table}
          basicColumnIds={BASIC_COLUMN_IDS}
          basicLabel="기본 컬럼"
          dynamicLabel="STB 정보 컬럼"
        />,
      );
      return items;
    },
    [hasActiveRecords, isFetching],
  );

  return (
    <div className="h-full flex flex-col">
      <div className="flex-1 p-4">
        <DataGrid
          data={tableData}
          columns={columns}
          isLoading={isLoading}
          getRowId={(row) => row.id}
          visibilityState={initialVisibility}
          onRowClick={openDetailSheet}
          rowCursor={true}
          enablePagination
          totalCount={data?.total}
          paginationConfig={{
            pageSize,
            pageIndex: currentPage - 1,
          }}
          pageSizes={[10, 50, 100, 500, 1000]}
          onPageChange={(pageIndex) => {
            setCurrentPage(pageIndex);
          }}
          onPageSizeChange={(newPageSize) => {
            setPageSize(newPageSize);
            setCurrentPage(1);
          }}
          leftFilters={renderLeftFilters}
          rightFilters={renderRightFilters}
        />
      </div>
    </div>
  );
}
