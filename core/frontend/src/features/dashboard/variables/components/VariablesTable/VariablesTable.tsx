'use client';

import React, {useMemo, useRef} from 'react';
import {getCoreRowModel, useReactTable} from '@tanstack/react-table';
import {DataGridTableView} from '@pharos/shared/components/ui-extension/data-grid/DataGridTableView';
import {useFlexColumnSizing} from '@pharos/shared/components/ui-extension/data-grid/hooks';
import {TABLE_TD_HEIGHT} from '@pharos/shared/components/ui-extension';
import {getVariablesColumns} from './columns';
import { Title } from "@pharos/shared/components/ui-extension";

export interface ToggleTableRow {
  id: string;
  type: string;
  description: string;
  [key: string]: any;
}

interface ToggleTableProps {
  title: string;
  rows: ToggleTableRow[];
  onRowsChange: (newRows: ToggleTableRow[]) => void;
  onDeleteRow: (rowId: string) => void;
  dashboardId?: string;
}

export default function VariablesTable({
  title,
  rows,
  onRowsChange,
  onDeleteRow,
  dashboardId,
}: ToggleTableProps) {
  const containerRef = useRef<HTMLDivElement>(null);

  const handleRowReorder = (fromIndex: number, toIndex: number) => {
    const newRows = [...rows];
    const [moved] = newRows.splice(fromIndex, 1);
    newRows.splice(toIndex, 0, moved);
    onRowsChange(newRows);
  };

  const columns = useMemo(
    () => getVariablesColumns({dashboardId, onDeleteRow}),
    [dashboardId, onDeleteRow],
  );

  const { columnSizing, handleColumnSizingChange } = useFlexColumnSizing({
    columns,
    containerRef,
  });

  const adjustedColumnSizing = useMemo(() => {
    const sizing = { ...columnSizing };
    if (sizing['description'] !== undefined) {
      sizing['description'] = Math.max(sizing['description'] - 33, 100);
    }
    return sizing;
  }, [columnSizing]);

  const table = useReactTable({
    data: rows,
    columns,
    state: { columnSizing: adjustedColumnSizing },
    onColumnSizingChange: handleColumnSizingChange,
    getCoreRowModel: getCoreRowModel(),
    getRowId: (row) => row.id,
    enableColumnResizing: true,
    columnResizeMode: 'onChange',
  });

  return (
    <div className={`rounded-[var(--radius)] overflow-hidden`}>
      <div className="flex items-center">
        <div className="flex items-center gap-3 ">
          <Title variant="h3">{title}</Title>
          <div className="mb-3 flex items-center justify-center w-5 h-5 bg-muted rounded-[var(--radius)] font-semibold text-xs">
            {rows.length}
          </div>
        </div>
      </div>

      <div ref={containerRef}>
          <DataGridTableView
            table={table}
            tableHeight={rows.length === 0 ? 138 : Math.min(rows.length * TABLE_TD_HEIGHT + (TABLE_TD_HEIGHT - 1), 5 * TABLE_TD_HEIGHT + (TABLE_TD_HEIGHT - 1))}
            emptyCustomMessage="There are no variables added yet"
            onRowReorder={handleRowReorder}
            className="bg-background"
          />
        </div>
    </div>
  );
}
