import {ColumnDef} from '@tanstack/react-table';
import {Checkbox} from '../../ui';
import {CheckboxConfig} from './types';

export function createCheckboxColumn<T>(checkboxConfig: CheckboxConfig<T>): ColumnDef<T> {
  const {
    checkboxKey = 'id',
    selectedItems = [],
    isSelectable,
    handleSelectAll,
    handleSelectOne,
  } = checkboxConfig;

  return {
    id: 'checkbox',
    size: 36,
    minSize: 36,
    maxSize: 36,
    enableResizing: false,
    meta: {cellClassName: '!p-0'},
    header: ({table: t}) => {
      const selectableRows = t
        .getFilteredRowModel()
        .flatRows.filter((r) =>
          typeof isSelectable === 'function' ? isSelectable(r.original) : isSelectable !== false,
        );
      const allKeys = selectableRows.map(
        (r) => (r.original as Record<string, unknown>)[checkboxKey] as string,
      );
      const isAllSelected =
        allKeys.length > 0 && allKeys.every((k) => selectedItems.includes(k));
      return (
        <div className="flex items-center justify-center">
          <Checkbox
            checked={isAllSelected}
            disabled={allKeys.length === 0}
            onCheckedChange={(checked) =>
              handleSelectAll?.(!!checked, selectableRows.map((r) => r.original))
            }
            className="border-primary shadow-none cursor-pointer"
          />
        </div>
      );
    },
    cell: ({row}) => {
      const rowSelectable =
        typeof isSelectable === 'function' ? isSelectable(row.original) : isSelectable !== false;
      if (!rowSelectable) return null;
      const key = (row.original as Record<string, unknown>)[checkboxKey] as string;
      const isChecked = selectedItems.includes(key);
      return (
        <div className="flex items-center justify-center" onClick={(e) => e.stopPropagation()}>
          <Checkbox
            checked={isChecked}
            onCheckedChange={(checked) => handleSelectOne?.(row.original, !!checked)}
            className="border-primary shadow-none cursor-pointer"
          />
        </div>
      );
    },
  };
}
