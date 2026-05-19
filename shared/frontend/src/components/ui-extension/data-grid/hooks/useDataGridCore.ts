import { useCallback, useEffect, useMemo, useState } from 'react';
import { ColumnFiltersState, FilterFnOption, SortingState } from '@tanstack/react-table';
import { useDataSource } from './useDataSource';
import { useTableStore } from './useTableStore';
import type { TableStoreState } from './useTableStore';
import { createGlobalFilterFn } from '../utils/globalFilterUtils';
import type { ColumnDef } from '@tanstack/react-table';
import type { UseDataResult } from '../types';

interface UseDataGridCoreParams<T> {
  columns: ColumnDef<T>[];
  useData?: () => UseDataResult<T>;
  data?: T[];
  tableKey?: string;
  persistState?: boolean;
  searchableColumns?: string[];
  globalFilterFn?: FilterFnOption<T>;
  columnFilters?: ColumnFiltersState;
  initialSorting?: SortingState;
  isLoading?: boolean;
  onDataLoaded?: (data: T[]) => void;
}

export function useDataGridCore<T>({
  columns,
  useData,
  data: propsData,
  tableKey,
  persistState = false,
  searchableColumns,
  globalFilterFn: customGlobalFilterFn,
  columnFilters: propsColumnFilters,
  initialSorting,
  isLoading: propsIsLoading,
  onDataLoaded,
}: UseDataGridCoreParams<T>) {
  // ============================================================================
  // 전역 상태 관리 (Zustand tableStore)
  // ============================================================================
  const { getTableState, updateTableState, registerTableKey, unregisterTableKey } = useTableStore();
  const tableState =
    persistState && tableKey
      ? getTableState(tableKey)
      : { pagination: { pageIndex: 0, pageSize: 10 }, searchTerm: '' };

  useEffect(() => {
    if (!persistState || !tableKey) return;
    registerTableKey(tableKey);
    return () => unregisterTableKey(tableKey);
  }, [persistState, tableKey, registerTableKey, unregisterTableKey]);

  // ============================================================================
  // State
  // ============================================================================
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>(propsColumnFilters || []);
  const [sorting, setSorting] = useState<SortingState>(initialSorting || []);
  const [globalFilter, setGlobalFilter] = useState<string>(tableState.searchTerm || '');

  // ============================================================================
  // 데이터 소스
  // ============================================================================
  const { data, finalColumns, isLoading: dataSourceLoading, error } = useDataSource({
    useData,
    data: propsData,
    columns,
  });

  const isLoading = propsIsLoading ?? dataSourceLoading;

  // ============================================================================
  // Handlers
  // ============================================================================
  const handleGlobalFilterChange = useCallback(
    (value: string) => {
      setGlobalFilter(value);
      if (persistState && tableKey) {
        updateTableState(tableKey, { searchTerm: value });
      }
    },
    [persistState, tableKey, updateTableState],
  );

  const defaultGlobalFilterFn = useMemo(() => {
    if (!searchableColumns) return undefined;
    return createGlobalFilterFn<T>(searchableColumns);
  }, [searchableColumns]);

  const resolvedGlobalFilterFn: FilterFnOption<T> | undefined =
    customGlobalFilterFn || defaultGlobalFilterFn;

  // ============================================================================
  // Effects
  // ============================================================================
  useEffect(() => {
    if (onDataLoaded && data) onDataLoaded(data as T[]);
  }, [data, onDataLoaded]);

  useEffect(() => {
    if (propsColumnFilters) setColumnFilters(propsColumnFilters);
  }, [propsColumnFilters]);

  useEffect(() => {
    setGlobalFilter(tableState.searchTerm || '');
  }, [tableState.searchTerm]);

  return {
    // 데이터
    data,
    finalColumns,
    isLoading,
    error,
    // 상태 (tableState는 내부에서만 searchTerm 참조용으로 쓰이므로 직접 반환하지 않음)
    columnFilters,
    setColumnFilters,
    sorting,
    setSorting,
    globalFilter,
    // 핸들러
    handleGlobalFilterChange,
    resolvedGlobalFilterFn,
    // store
    updateTableState,
    // pagination을 위한 tableStatePagination
    tableStatePagination: tableState.pagination as TableStoreState['pagination'],
    tableStateSearchTerm: tableState.searchTerm,
  };
}
