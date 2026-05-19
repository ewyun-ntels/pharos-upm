import React, {useMemo} from 'react';
import {TablePanelOptions, FilterItem} from './types';
import {DataGrid, IconButton, SearchInput} from '@pharos/shared/components/ui-extension';
import {Download} from '@pharos/shared/components';
import {ColumnDef, Table} from '@tanstack/react-table';
import {useTableColumns} from '@pharos/shared/hooks/table-columns';
import type {ColumnConfig, TableList} from '@pharos/shared/hooks/table-columns';
import type {PanelContentProps} from '@features/dashboard/panels/plugins/common/withChartPanel';
import {withCardChartPanel} from '@features/dashboard/panels/plugins/common/withCardChartPanel';
import {DataLinkCell} from './DataLinkCell';
import {usePanelAction} from '@features/dashboard/panels/plugins/common/usePanelAction';

/**
 * TableContent - Table 패널 내부 컴포넌트
 *
 * withRawCard HOC가 Card UI, 데이터 fetching, ResizeObserver를 처리하므로
 * 이 컴포넌트는 DataGrid 렌더링에만 집중합니다.
 */
function TableContent({
  options: tableOptions,
  dashboardId,
  id,
  title,
  displayName,
  chartData,
  filterMetas,
  onPanelAction,
}: PanelContentProps<TablePanelOptions, Record<string, unknown>[]>) {

  const dispatch = usePanelAction(onPanelAction);

  // columnDataLinks가 있는 컬럼에 customCell(DataLinkCell) 주입
  const resolvedColumnConfigs = useMemo((): ColumnConfig<TableList>[] => {
    const configs = (tableOptions?.columnConfigs || []) as ColumnConfig<TableList>[];
    const dataLinks = tableOptions?.columnDataLinks || {};

    return configs.map((config) => {
      const links = dataLinks[String(config.key)];
      if (!links?.length) return config;

      return {
        ...config,
        customCell: (value: unknown, row: Record<string, unknown>) => (
          <DataLinkCell
            value={value}
            row={row}
            links={links}
            filterMetas={filterMetas}
            cellContent={String(value ?? '')}
          />
        ),
      } as ColumnConfig<TableList>;
    });
  }, [tableOptions?.columnConfigs, tableOptions?.columnDataLinks, filterMetas]);

  const {columns} = useTableColumns({
    columnConfigs: resolvedColumnConfigs,
    showCheckbox: tableOptions?.checkboxConfig?.showCheckbox,
  });

  const renderLeftFilters = (table: Table<Record<string, unknown>>) => {
    return (tableOptions?.leftItems || [])
      .map((filter: FilterItem, index: number) => {
        if (filter.type === 'search') {
          return (
            <SearchInput
              key={`left-${index}`}
              size="hsmall"
              table={table}
              placeholder={filter.placeholder}
            />
          );
        }
        return null;
      })
      .filter(Boolean);
  };

  const handleExport = (table: Table<Record<string, unknown>>) => {
    const rows = table.getFilteredRowModel().rows;
    const visibleColumns = table.getAllColumns().filter((col) => col.getIsVisible() && col.id !== 'select');
    const headers = visibleColumns.map((col) => col.id);

    const csvContent = [
      headers.join(','),
      ...rows.map((row) =>
        headers.map((h) => JSON.stringify(row.original[h] ?? '')).join(','),
      ),
    ].join('\n');

    const blob = new Blob([csvContent], {type: 'text/csv;charset=utf-8;'});
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${title || displayName || 'export'}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const renderRightFilters = (table: Table<Record<string, unknown>>) => {
    return (tableOptions?.rightItems || [])
      .map((filter: FilterItem, index: number) => {
        if (filter.type === 'export') {
          return (
            <IconButton
              key={`right-${index}`}
              icon={<Download className="size-4" />}
              variant="outline"
              size="icon"
              onClick={() => handleExport(table)}
              disabled={filter.disabled}
            />
          );
        }
        return null;
      })
      .filter(Boolean);
  };

  return (
    <DataGrid<Record<string, unknown>>
      tableKey={`dashboard-${dashboardId}-panel-${id}`}
      persistState={true}
      data={chartData ?? []}
      columns={columns as ColumnDef<Record<string, unknown>>[]}
      leftFilters={renderLeftFilters}
      rightFilters={renderRightFilters}
      tableHeight="flex"
      enablePagination={tableOptions?.usePagenation}
      checkboxConfig={tableOptions?.checkboxConfig}
      columnFilters={tableOptions?.columnFilters}
      title={title}
      displayName={displayName}
      className="px-4 pb-4 pt-[2px]"
      onColumnSizingChange={(sizing) => dispatch('resize-column', sizing)}
    />
  );
}

const TableCardChart = withCardChartPanel<TablePanelOptions, Record<string, unknown>[]>(TableContent);

export {TableCardChart};
