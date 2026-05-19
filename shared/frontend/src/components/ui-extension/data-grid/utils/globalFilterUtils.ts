/**
 * Global Filter 유틸리티
 */

import {Row, FilterFn} from '@tanstack/react-table';

export function createGlobalFilterFn<T>(searchableColumns: string[]): FilterFn<T> {
  return (row: Row<T>, _columnId: string, filterValue: string, _addMeta) => {
    if (!filterValue) return true;

    const searchLower = filterValue.toLowerCase();

    return searchableColumns.some((col) => {
      const value = row.getValue(col);
      if (value == null) return false;

      const stringValue = String(value).toLowerCase();
      return stringValue.includes(searchLower);
    });
  };
}
