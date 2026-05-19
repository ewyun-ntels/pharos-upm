'use client';

import { ColumnDef } from '@tanstack/react-table';
import { DataGrid, DataGridProps } from '../../../ui-extension';

interface TablePaginationProps<T> extends Omit<DataGridProps<T>, 'columns' | 'data' | 'children'> {
  data: T[];
  columns: ColumnDef<T>[];
  onPageChange?: (page: number) => void;
  onPageSizeChange?: (pageSize: number) => void;
  pageSizes?: number[];
  tableKey?: string;
  className?: string;
  paginationVariant?: 'default' | 'numbered';
}

export function TablePagination<T>({
  data,
  columns,
  onPageChange,
  onPageSizeChange,
  pageSizes = [10, 20, 50],
  tableKey = 'table-pagination',
  className,
  paginationVariant = 'default',
  ...rest
}: TablePaginationProps<T>) {
  return (
    <DataGrid<T>
      data={data}
      columns={columns}
      tableKey={tableKey}
      className={className}
      enablePagination={true}
      onPageChange={onPageChange}
      onPageSizeChange={onPageSizeChange}
      paginationVariant={paginationVariant}
      pageSizes={pageSizes}
      {...rest}
    />
  );
}
