/**
 * DataGrid
 *
 * 통합 테이블 컴포넌트
 * - Dashboard Panel: useData 콜백 사용
 * - 일반 Page: data props 직접 전달
 * - 모든 테이블 기능 통합 (Filter, Pagination, Export 등)
 */

import {useCallback, useEffect, useMemo, useRef, useState} from 'react';
import {
  ColumnDef,
  ColumnSizingState,
  PaginationState,
  getCoreRowModel,
  getExpandedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table';
import {DataTablePagination, LoadingIndicator} from '@pharos/shared/components/ui-extension';
import {Alert, AlertDescription, AlertTitle} from '../../ui';
import {AlertTriangle} from 'lucide-react';
import {DataGridProps} from './types';
import {DataGridFilters} from './DataGridFilters';
import {DataGridTableView} from './DataGridTableView';
import {createCheckboxColumn} from './checkbox-column';
import {useDataGridCore, useTableHeight, useFlexColumnSizing} from './hooks';
import {cn} from '../../../lib';

const DEFAULT_PAGE_SIZE = 10;

export function DataGrid<T>(props: DataGridProps<T>) {
  const {
    // 데이터 소스
    useData,
    data: propsData,
    columns,
    getRowId,
    getSubRows,
    tableKey,
    persistState = false,
    // 필터
    leftFilters,
    rightFilters,
    columnFilters: propsColumnFilters,
    // 검색
    searchableColumns,
    globalFilterFn: customGlobalFilterFn,
    // 페이지네이션
    enablePagination = true,
    paginationConfig = {pageSize: DEFAULT_PAGE_SIZE, initialPageIndex: 0},
    pageSizes = [10, 20, 50],
    paginationVariant = 'default',
    totalCount,
    onPageChange,
    onPageSizeChange,
    // 테이블 기능
    enableColumnResizing = true,
    enableExpandingColumn,
    treeColumnId,
    useTableSorting = true,
    initialSorting,
    visibilityState,
    excludeRowClickColumns,
    // UI
    tableHeight: propsTableHeight,
    description,
    emptyCustomMessage,
    rowCursor,
    tableWidthMode,
    className,
    isLoading: propsIsLoading,
    // 체크박스
    checkboxConfig,
    // 이벤트 핸들러
    onRowClick,
    dataCellOnClick,
    onDataLoaded,
    onTableReady,
    onColumnSizingChange,
  } = props;

  // ============================================================================
  // Validation: persistState=true일 때 tableKey 필수
  // ============================================================================
  if (persistState && !tableKey) {
    throw new Error(
      '[DataGrid] persistState=true requires tableKey. ' +
        'Please provide a unique tableKey or set persistState=false.',
    );
  }

  // ============================================================================
  // Refs & State
  // ============================================================================
  const tableRef = useRef<HTMLDivElement>(null);
  const filterRef = useRef<HTMLDivElement>(null);
  const paginationRef = useRef<HTMLDivElement>(null);
  const tableBodyRef = useRef<HTMLTableSectionElement | null>(null);

  // ============================================================================
  // 공통 코어 (전역 상태, 데이터 소스, 필터/정렬 상태, globalFilter 핸들러)
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
    updateTableState,
    tableStatePagination,
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

  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: paginationConfig.initialPageIndex || 0,
    pageSize: paginationConfig.pageSize || DEFAULT_PAGE_SIZE,
  });
  // document pointerup 시 onColumnSizingChange 호출을 위한 ref
  const latestSizingRef = useRef<Record<string, number>>({});
  const pendingResizeCallbackRef = useRef(false);

  const [localColumnVisibility, setLocalColumnVisibility] = useState(visibilityState || {});

  // ============================================================================
  // Custom Hooks & Handlers
  // ============================================================================

  // Checkbox 전용 컬럼 (checkboxConfig 제공 시 맨 앞에 자동 삽입)
  const columnsWithCheckbox = useMemo((): ColumnDef<T>[] => {
    if (!checkboxConfig) return finalColumns;
    return [createCheckboxColumn<T>(checkboxConfig), ...finalColumns];
  }, [checkboxConfig, finalColumns]);

  const {columnSizing, handleColumnSizingChange: flexHandleColumnSizingChange} =
    useFlexColumnSizing({
      columns: columnsWithCheckbox,
      columnVisibility: localColumnVisibility,
      containerRef: tableRef,
      isLoading,
    });

  // tanstack resize handle이 document level에서 pointerup을 처리할 수 있으므로
  // document listener로 등록해야 안정적으로 감지됨
  const handleColumnSizingChange = useCallback(
    (updater: ColumnSizingState | ((old: ColumnSizingState) => ColumnSizingState)) => {
      const nextSizing = typeof updater === 'function' ? updater(latestSizingRef.current) : updater;
      latestSizingRef.current = nextSizing;
      if (onColumnSizingChange) pendingResizeCallbackRef.current = true;
      flexHandleColumnSizingChange(updater);
    },
    [onColumnSizingChange, flexHandleColumnSizingChange],
  );

  useEffect(() => {
    if (!onColumnSizingChange) return;
    const handleDocPointerUp = () => {
      if (pendingResizeCallbackRef.current) {
        pendingResizeCallbackRef.current = false;
        onColumnSizingChange(latestSizingRef.current);
      }
    };
    document.addEventListener('pointerup', handleDocPointerUp);
    return () => document.removeEventListener('pointerup', handleDocPointerUp);
  }, [onColumnSizingChange]);

  // Pagination 핸들러
  const handlePaginationChange = useCallback(
    (updater: PaginationState | ((prev: PaginationState) => PaginationState)) => {
      if (!enablePagination) return;
      if (persistState && tableKey) {
        const newPagination =
          typeof updater === 'function' ? updater(tableStatePagination) : updater;
        updateTableState(tableKey, {pagination: newPagination});
        if (onPageChange && newPagination.pageIndex !== tableStatePagination.pageIndex)
          onPageChange(newPagination.pageIndex + 1);
        if (onPageSizeChange && newPagination.pageSize !== tableStatePagination.pageSize)
          onPageSizeChange(newPagination.pageSize);
      } else {
        const prevPagination = pagination;
        const newPagination = typeof updater === 'function' ? updater(prevPagination) : updater;
        setPagination(updater);
        if (onPageChange && newPagination.pageIndex !== prevPagination.pageIndex)
          onPageChange(newPagination.pageIndex + 1);
        if (onPageSizeChange && newPagination.pageSize !== prevPagination.pageSize)
          onPageSizeChange(newPagination.pageSize);
      }
    },
    [
      enablePagination,
      persistState,
      tableKey,
      tableStatePagination,
      pagination,
      updateTableState,
      onPageChange,
      onPageSizeChange,
    ],
  );

  // persistState=true && tableKey가 있으면 전역 상태, 아니면 로컬 상태 사용
  const currentPagination = enablePagination
    ? persistState && tableKey
      ? tableStatePagination
      : pagination
    : {pageIndex: 0, pageSize: DEFAULT_PAGE_SIZE};

  const table = useReactTable<T>({
    data,
    columns: columnsWithCheckbox,
    state: {
      pagination: enablePagination ? currentPagination : undefined,
      columnFilters,
      sorting,
      globalFilter,
      columnSizing,
      columnVisibility: localColumnVisibility,
    },
    // 서버 사이드 pagination 설정
    ...(totalCount !== undefined && {
      manualPagination: true,
      pageCount: Math.ceil(totalCount / (currentPagination.pageSize || DEFAULT_PAGE_SIZE)),
    }),
    getRowId,
    getSubRows,
    filterFromLeafRows: !!getSubRows,
    onPaginationChange: enablePagination ? handlePaginationChange : undefined,
    onColumnFiltersChange: setColumnFilters,
    onSortingChange: useTableSorting ? setSorting : undefined,
    onGlobalFilterChange: handleGlobalFilterChange,
    onColumnSizingChange: handleColumnSizingChange,
    onColumnVisibilityChange: setLocalColumnVisibility,
    getCoreRowModel: getCoreRowModel(),
    getExpandedRowModel: getExpandedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: enablePagination ? getPaginationRowModel() : undefined,
    getFacetedUniqueValues: getFacetedUniqueValues(),
    ...(globalFilterFn !== undefined && { globalFilterFn }),
    enableColumnResizing,
    columnResizeMode: 'onChange',
    enableSorting: useTableSorting,
    meta: {
      dataCellOnClick,
      enableExpandingColumn,
      treeColumnId,
    },
  });

  // 높이 계산
  const calcTableHeight = useTableHeight({
    tableHeight: propsTableHeight,
    useData: !!useData,
    dataLength: data.length,
    tableRef,
    filterRef,
    paginationRef,
    tableBodyRef,
    isLoading,
  });

  useEffect(() => {
    if (visibilityState) {
      setLocalColumnVisibility(visibilityState);
    }
  }, [visibilityState]);

  useEffect(() => {
    if (!enablePagination) return;
    const pageIndex = table.getState().pagination?.pageIndex || 0;
    const pageCount = table.getPageCount();
    if (pageCount > 0 && pageIndex >= pageCount) {
      table.setPageIndex(0);
    }
  }, [table, enablePagination]);

  // 필터(globalFilter, columnFilters) 변경 시 첫 페이지로 리셋
  useEffect(() => {
    if (!enablePagination) return;
    if ((table.getState().pagination?.pageIndex ?? 0) === 0) return;
    table.setPageIndex(0);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [columnFilters, globalFilter, enablePagination]);

  // paginationConfig.pageIndex 변경 시 내부 pagination 동기화 (서버 사이드 외부 제어)
  useEffect(() => {
    if (!enablePagination) return;
    const targetIndex = paginationConfig.pageIndex;
    if (targetIndex === undefined) return;
    if (targetIndex === (table.getState().pagination?.pageIndex ?? 0)) return;
    table.setPageIndex(targetIndex);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [paginationConfig.pageIndex, enablePagination]);

  // error 시 early return
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

  return (
    <div
      ref={tableRef}
      className={cn(
        'flex flex-col gap-0',
        propsTableHeight === 'flex' && 'h-full min-h-0',
        className,
      )}
    >
      {(leftFilters || rightFilters || description) && (
        <div ref={filterRef}>
          <DataGridFilters
            leftFilters={leftFilters}
            rightFilters={rightFilters}
            description={description}
            table={table}
          />
        </div>
      )}

      {isLoading ? (
        <LoadingIndicator />
      ) : (
        <DataGridTableView
          table={table}
          tableHeight={calcTableHeight}
          onRowClick={onRowClick}
          dataCellOnClick={dataCellOnClick}
          enableExpandingColumn={enableExpandingColumn}
          treeColumnId={treeColumnId}
          excludeRowClickColumns={excludeRowClickColumns}
          emptyCustomMessage={emptyCustomMessage}
          rowCursor={rowCursor}
          tableWidthMode={tableWidthMode}
          onTableReady={onTableReady}
          tableBodyRef={tableBodyRef}
          className={'flex-1 min-h-0'}
        />
      )}

      {enablePagination && !isLoading && (
        <div className={cn('mt-3')} ref={paginationRef}>
          <DataTablePagination
            table={table}
            additionalPageSizeItem={pageSizes}
            totalCount={totalCount}
            variant={paginationVariant}
          />
        </div>
      )}
    </div>
  );
}
