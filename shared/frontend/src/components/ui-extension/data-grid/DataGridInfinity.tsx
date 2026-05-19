/**
 * DataGridInfinity
 *
 * 무한 스크롤 테이블 컴포넌트
 * - DataGrid에서 페이지네이션 제거
 * - IntersectionObserver 기반 자동 로드
 * - 데이터 누적 방식
 */

import { useRef } from 'react';
import {
  getCoreRowModel,
  getExpandedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table';
import { DataGridInfinityProps } from './types';
import { DataGridFilters } from './DataGridFilters';
import { DataGridTableView } from './DataGridTableView';
import { useDataGridCore, useInfiniteScroll, useFlexColumnSizing } from './hooks';
import { cn } from '../../../lib';
import { LoadingIndicator } from '../loading-indicator';
import { Alert, AlertDescription, AlertTitle } from '../../ui';
import { AlertTriangle } from 'lucide-react';

export function DataGridInfinity<T>(props: DataGridInfinityProps<T>) {
  const {
    // 데이터 소스
    useData,
    data: propsData,
    columns,
    getRowId,
    getSubRows,
    tableKey,
    persistState = false,

    // 무한 스크롤
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
    rootMargin = '100px',

    // 필터
    leftFilters,
    rightFilters,
    columnFilters: propsColumnFilters,

    // 검색
    searchableColumns,
    globalFilterFn: customGlobalFilterFn,

    // 테이블 기능
    enableColumnResizing = true,
    enableExpandingColumn,
    useTableSorting = true,
    initialSorting,
    visibilityState,
    excludeRowClickColumns,

    // UI
    tableHeight: propsTableHeight = 'flex',
    emptyCustomMessage,
    rowCursor,
    tableWidthMode,
    className,
    isLoading: propsIsLoading,

    // 이벤트 핸들러
    onRowClick,
    dataCellOnClick,
    onDataLoaded,
    onTableReady,
  } = props;

  // ============================================================================
  // Validation
  // ============================================================================
  if (persistState && !tableKey) {
    throw new Error(
      '[DataGridInfinity] persistState=true requires tableKey. ' +
        'Please provide a unique tableKey or set persistState=false.'
    );
  }

  // ============================================================================
  // Refs & State
  // ============================================================================
  const tableRef = useRef<HTMLDivElement>(null);
  const filterRef = useRef<HTMLDivElement>(null);
  const tableBodyRef = useRef<HTMLTableSectionElement | null>(null);

  // ============================================================================
  // 공통 코어
  // ============================================================================
  const {
    data,
    finalColumns,
    isLoading,
    error,
    columnFilters,
    setColumnFilters,
    sorting,
    setSorting,
    globalFilter,
    handleGlobalFilterChange,
    resolvedGlobalFilterFn: globalFilterFn,
  } = useDataGridCore<T>({
    columns,
    useData,
    data: propsData,
    tableKey,
    persistState,
    searchableColumns,
    globalFilterFn: customGlobalFilterFn,
    columnFilters: propsColumnFilters,
    initialSorting,
    isLoading: propsIsLoading,
    onDataLoaded,
  });

  const { columnSizing, handleColumnSizingChange } = useFlexColumnSizing({
    columns: finalColumns,
    columnVisibility: visibilityState,
    containerRef: tableRef,
    isLoading,
  });

  // ============================================================================
  // Custom Hooks
  // ============================================================================

  // 무한 스크롤 hook
  const { loadMoreRef } = useInfiniteScroll({
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
    rootMargin,
    enabled: !isLoading && data.length > 0,
  });

  // ============================================================================
  // TanStack Table
  // ============================================================================

  const table = useReactTable<T>({
    data,
    columns: finalColumns,
    state: {
      columnFilters,
      sorting,
      globalFilter,
      columnSizing,
      ...(visibilityState && { columnVisibility: visibilityState }),
    },
    getRowId,
    getSubRows,
    onColumnFiltersChange: setColumnFilters,
    onSortingChange: useTableSorting ? setSorting : undefined,
    onGlobalFilterChange: handleGlobalFilterChange,
    onColumnSizingChange: handleColumnSizingChange,
    getCoreRowModel: getCoreRowModel(),
    getExpandedRowModel: getExpandedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    ...(globalFilterFn !== undefined && { globalFilterFn }),
    enableColumnResizing,
    columnResizeMode: 'onChange',
    enableSorting: useTableSorting,
    meta: {
      dataCellOnClick,
      enableExpandingColumn,
    },
  });

  // ============================================================================
  // Render
  // ============================================================================

  // 초기 로딩 (데이터 없을 때만)
  if (isLoading && data.length === 0) {
    return <LoadingIndicator />;
  }

  if (error) {
    return (
      <div className="p-3">
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error.message}</AlertDescription>
        </Alert>
      </div>
    );
  }

  const isFlex = propsTableHeight === 'flex';

  return (
    <div ref={tableRef} className={cn('flex flex-col gap-0', isFlex && 'h-full min-h-0', className)}>
      {(leftFilters || rightFilters) && (
        <div ref={filterRef}>
          <DataGridFilters leftFilters={leftFilters} rightFilters={rightFilters} table={table} />
        </div>
      )}

      <DataGridTableView
        table={table}
        tableHeight={propsTableHeight}
        onRowClick={onRowClick}
        dataCellOnClick={dataCellOnClick}
        enableExpandingColumn={enableExpandingColumn}
        excludeRowClickColumns={excludeRowClickColumns}
        emptyCustomMessage={emptyCustomMessage}
        rowCursor={rowCursor}
        tableWidthMode={tableWidthMode}
        onTableReady={onTableReady}
        tableBodyRef={tableBodyRef}
        className="flex-1 min-h-0"
        infiniteScroll={{
          isFetching: isFetchingNextPage,
          loadMoreRef,
          hasNextPage,
        }}
      />
    </div>
  );
}
