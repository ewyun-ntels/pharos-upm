/**
 * useDataSource Hook
 *
 * 데이터 소스 처리 및 자동 컬럼 생성 로직
 */

import { useMemo } from 'react';
import { ColumnDef } from '@tanstack/react-table';
import { UseDataResult } from '../types';

interface UseDataSourceProps<T> {
  useData?: () => UseDataResult<T>;
  data?: T[];
  columns: ColumnDef<T>[];
}

interface UseDataSourceReturn<T> {
  data: T[];
  finalColumns: ColumnDef<T>[];
  isLoading: boolean;
  error?: Error;
}

export function useDataSource<T = unknown>({
  useData,
  data: propsData,
  columns,
}: UseDataSourceProps<T>): UseDataSourceReturn<T> {

  // 데이터 소스 처리 - Hook은 항상 최상위에서 호출
  const fetchedData = useData?.();

  const data = useMemo(() => {
    if (fetchedData) {
      return fetchedData.data || [];
    }
    return propsData || [];
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [fetchedData?.data, propsData]);

  const isLoading = fetchedData?.isLoading || false;
  const error = fetchedData?.error;

  // 자동 컬럼 생성 로직
  const finalColumns = useMemo(() => {
    if (columns && columns.length > 0) {
      return columns;
    }

    // 데이터가 있으면 첫 번째 row의 key들로 columns 생성
    if (data && data.length > 0) {
      const firstRow = data[0] as Record<string, unknown>;
      const autoColumns = Object.keys(firstRow).map((key) => ({
        accessorKey: key,
        header: key,
        id: key,
      }));

      return autoColumns as ColumnDef<T>[];
    }

    return columns;
  }, [columns, data]);

  return {
    data,
    finalColumns,
    isLoading,
    error,
  };
}
