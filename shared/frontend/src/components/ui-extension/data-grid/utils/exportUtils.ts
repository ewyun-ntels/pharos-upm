/**
 * Export 유틸리티
 */

import {Table, ColumnDef, AccessorKeyColumnDef} from '@tanstack/react-table';
import {ExportConfig} from '../types';

/** accessorKey를 가진 컬럼인지 확인하는 타입 가드 */
function hasAccessorKey<T>(col: ColumnDef<T>): col is AccessorKeyColumnDef<T> {
  return 'accessorKey' in col && col.accessorKey != null;
}

export function exportTableData<T>(
  table: Table<T>,
  columns: ColumnDef<T>[],
  config?: ExportConfig<T>,
  defaultFileName?: string
) {
  const fileName = config?.fileName || defaultFileName || 'table-export';
  const rows = table.getFilteredRowModel().rows;

  if (rows.length === 0) {
    console.warn('No data to export');
    return;
  }

  // accessorKey가 있는 컬럼만 필터링
  const accessorColumns = columns.filter(hasAccessorKey);

  // Header 추출
  const headers = accessorColumns
    .filter((col) => col.header)
    .map((col) => String(col.header));

  // Data 추출
  const data = rows.map((row) => {
    const rowData = row.original;
    const mappedData = config?.mapData ? config.mapData(rowData) : (rowData as Record<string, unknown>);

    return accessorColumns.map((col) => {
      const key = col.accessorKey as string;
      const value = mappedData[key];
      return value != null ? String(value) : '';
    });
  });

  // CSV 생성
  const csvContent = [
    headers.join(','),
    ...data.map((row) => row.map((cell) => `"${cell}"`).join(',')),
  ].join('\n');

  // Download
  const blob = new Blob([csvContent], {type: 'text/csv;charset=utf-8;'});
  const link = document.createElement('a');
  const url = URL.createObjectURL(blob);

  link.setAttribute('href', url);
  link.setAttribute('download', `${fileName}.csv`);
  link.style.visibility = 'hidden';

  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}
