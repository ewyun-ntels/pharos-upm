'use client';

import * as React from 'react';
import {flexRender, Table as TanstackTable, Row, Cell} from '@tanstack/react-table';
import {ChevronRight, GripVertical} from 'lucide-react';
import {DoubleArrowRightIcon} from '@radix-ui/react-icons';

import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core';
import {
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import {CSS} from '@dnd-kit/utilities';
import {Button} from '@pharos/shared/components/ui';
import {cn} from '../../lib';

import {ColumnResizer} from './column-resizer';
import ScrollTable from './scroll-table';
import {TableBody, TableCell, TableHead, TableHeader, TableRow} from '../ui';

type TableWidthMode = 'independent' | 'spacer' | 'fill-last';

interface DataTableRowProps<TData> {
  row: Row<TData>;
  enableExpandingColumn?: boolean;
  treeColumnId?: string;
  rowClassName?: string;
  depth: number;
  dataRowOnClick?: (row: Row<TData>) => void;
  dataCellOnClick?: (cell: Cell<TData, unknown>) => void;
  rowCursor?: boolean;
  tdHeight?: string | number | undefined;
  excludeRowClickColumns?: string[];
  isLastRow?: boolean;
  tableWidthMode?: TableWidthMode;
  // drag 관련 (opt-in)
  enableDragRow?: boolean;
  dragRowRef?: (node: HTMLElement | null) => void;
  dragRowStyle?: React.CSSProperties;
  dragHandleProps?: Record<string, any>;
}

//TableCell right border 라인 + 텍스트 오버플로우 처리
// - overflow-hidden: 셀 밖으로 텍스트 튀어나가지 않도록
// - text-ellipsis: 넘치는 텍스트는 ... 으로 표시
// - whitespace-nowrap은 ui/table.tsx의 TableCell 기본값으로 이미 적용됨
const cellBorder = 'border-r [&:last-child]:border-r-0 py-1 px-3 overflow-hidden text-ellipsis';

// data-table size constants
const thHeight = 8; // Tailwind h-8 class
export const TABLE_TD_HEIGHT = 32 + 1; // 1px border-b

function DataTableRow<TData>({
  row,
  enableExpandingColumn,
  treeColumnId,
  rowClassName,
  depth,
  dataRowOnClick,
  dataCellOnClick,
  rowCursor = false,
  tdHeight,
  excludeRowClickColumns = [],
  isLastRow = false,
  tableWidthMode = 'spacer',
  enableDragRow = false,
  dragRowRef,
  dragRowStyle,
  dragHandleProps,
}: DataTableRowProps<TData>) {
  const handleSelectedButton = () => {
    row.toggleExpanded();
  };

  const depthClass = 'pl-' + (depth * 4).toString();
  const visibleCells = row.getVisibleCells();

  return (
    <>
      <TableRow
        key={`table-row-${row.id}`}
        ref={enableDragRow ? dragRowRef : undefined}
        style={{height: tdHeight, ...(enableDragRow && dragRowStyle ? dragRowStyle : {})}}
        className={cn(
          rowClassName,
          `bg-background data-[state=selected]:bg-muted hover:bg-muted/30`,
          '[&>td]:border-b',
          isLastRow && '[&>td]:border-b-0',
          rowCursor ? 'cursor-pointer' : 'cursor-default',
        )}
        onClick={() => {
          dataRowOnClick && dataRowOnClick(row);
        }}
        data-state={row.getIsSelected() && 'selected'}
      >
        {enableDragRow && (
          <TableCell
            className={cn(cellBorder, 'cursor-grab')}
            style={{width: '33px', minWidth: '33px', maxWidth: '33px', padding: '0'}}
          >
            <div className="flex items-center justify-center">
              <GripVertical size={16} className="text-muted-foreground" {...dragHandleProps} />
            </div>
          </TableCell>
        )}
        {enableExpandingColumn && (
          <TableCell className={cn(cellBorder)}>
            {/* TODO CellOnClick이 필요할경우 추가 개발 필요 (group 되어있는 cell click이 필요할때는 따로 props빼는게 좋지 않을지 */}
            <div className={cn(depthClass, 'overflow-hidden')}>
              {row.subRows.length > 0 ? (
                <Button className={'h-7 w-7 p-0'} variant={'ghost'} onClick={handleSelectedButton}>
                  {row.getIsExpanded() ? (
                    <ChevronRight className={'h-4 w-4 rotate-90 transition-all'} />
                  ) : (
                    <ChevronRight className={'h-4 w-4 transition-all'} />
                  )}
                </Button>
              ) : (
                <Button
                  disabled={true}
                  className={'h-7 w-7 p-0 text-transparent'}
                  variant={'ghost'}
                >
                  <ChevronRight className={'h-4 w-4'} />
                </Button>
              )}
            </div>
          </TableCell>
        )}
        {visibleCells.map((cell, cellIndex) => {
          const colSize = cell.column.getSize();
          const isLastCell = cellIndex === visibleCells.length - 1;
          const isFillCell = tableWidthMode === 'fill-last' && isLastCell;
          const hasFixedSize = cell.column.columnDef.maxSize !== undefined;
          const isTreeCell = !!treeColumnId && cell.column.id === treeColumnId;

          const tdWidth = isFillCell ? 'auto' : `${colSize}px`;

          return (
            <TableCell
              onClick={
                dataCellOnClick
                  ? () => {
                      dataCellOnClick(cell);
                    }
                  : excludeRowClickColumns.includes(cell.column.id)
                    ? (e) => {
                        e.stopPropagation();
                      }
                    : undefined
              }
              key={cell.id}
              style={{
                width: tdWidth,
                minWidth: `${colSize}px`,
                maxWidth: hasFixedSize && !isFillCell ? `${colSize}px` : undefined,
              }}
              className={cn(cellBorder, cell.column.columnDef.meta?.cellClassName)}
            >
              {isTreeCell ? (
                <div
                  className="flex items-center overflow-hidden"
                  style={{paddingLeft: `${depth * 16}px`}}
                >
                  {row.subRows.length > 0 ? (
                    <Button
                      className="h-6 w-6 p-0 mr-0.5 shrink-0"
                      variant="ghost"
                      onClick={(e) => {
                        e.stopPropagation();
                        row.toggleExpanded();
                      }}
                    >
                      <ChevronRight
                        className={cn(
                          'h-3 w-3 transition-all',
                          row.getIsExpanded() && 'rotate-90',
                        )}
                      />
                    </Button>
                  ) : (
                    <span className="inline-block w-6 mr-0.5 shrink-0" />
                  )}
                  <span className="overflow-hidden text-ellipsis whitespace-nowrap">
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </span>
                </div>
              ) : (
                <div className={cn(treeColumnId ? '' : depthClass, 'overflow-hidden text-ellipsis')}>
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </div>
              )}
            </TableCell>
          );
        })}
        {/* spacer 모드: 각 row 끝에 빈 td로 나머지 공간 채움 */}
        {tableWidthMode === 'spacer' && (
          <TableCell className="border-r-0 p-0" style={{width: 'auto'}} />
        )}
      </TableRow>
    </>
  );
}

// dnd-kit useSortable을 적용한 DataTableRow wrapper
function SortableDataTableRow<TData>(
  props: Omit<DataTableRowProps<TData>, 'enableDragRow' | 'dragRowRef' | 'dragRowStyle' | 'dragHandleProps'>,
) {
  const rowId = (props.row.original as any)?.id ?? props.row.id;
  const {attributes, listeners, setNodeRef, transform, transition, isDragging} = useSortable({
    id: rowId,
  });

  return (
    <DataTableRow
      {...props}
      enableDragRow
      dragRowRef={setNodeRef}
      dragRowStyle={{
        transform: CSS.Transform.toString(transform),
        transition,
        opacity: isDragging ? 0.5 : 1,
        position: isDragging ? 'relative' : undefined,
        zIndex: isDragging ? 1 : undefined,
      }}
      dragHandleProps={{...attributes, ...listeners}}
    />
  );
}

interface DataTableProps<TData> {
  table: TanstackTable<TData>;
  enableExpandingColumn?: boolean;
  /** Expand toggle을 주입할 컬럼 ID. 지정 시 해당 컬럼에 chevron + indent 자동 주입 */
  treeColumnId?: string;
  headerClassName?: string;
  tableWidth?: string | number | undefined;
  tableHeight?: string | number | undefined;
  dataRowOnClick?: (row: Row<TData>) => void;
  dataCellOnClick?: (cell: Cell<TData, unknown>) => void;
  rowCursor?: boolean;
  className?: string;
  emptyCustomMessage?: string;
  infiniteScroll?: {
    isFetching: boolean;
    loadMoreRef: React.RefObject<HTMLDivElement | null>;
    hasNextPage: boolean;
  };
  tdHeight?: string | number | undefined;
  minTableHeight?: number; // 최소 테이블 높이
  excludeRowClickColumns?: string[];
  tableBodyRef?: React.RefObject<HTMLTableSectionElement | null>;
  /** 행 드래그앤드롭 재정렬 콜백. 제공 시 GripVertical 핸들 컬럼이 자동 추가됨 */
  onRowReorder?: (fromIndex: number, toIndex: number) => void;
  /** 테이블 너비 처리 방식 */
  tableWidthMode?: TableWidthMode;
}

export function DataTable<TData>({
  table,
  enableExpandingColumn,
  treeColumnId,
  headerClassName = 'bg-background',
  tableHeight,
  tableWidth,
  dataRowOnClick,
  dataCellOnClick,
  rowCursor = false,
  className,
  emptyCustomMessage,
  infiniteScroll,
  tdHeight = TABLE_TD_HEIGHT,
  excludeRowClickColumns = [],
  tableBodyRef,
  onRowReorder,
  tableWidthMode = 'fill-last',
}: DataTableProps<TData>) {
  const setAllExpanding = () => {
    table.toggleAllRowsExpanded();
  };

  // ✅ table.getRowModel().rows를 직접 사용 (useMemo 제거 - table 객체 참조가 바뀌지 않아 업데이트 안 됨)
  const rows = table.getRowModel().rows;

  // dnd-kit sensors (onRowReorder 없어도 hooks는 항상 호출)
  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {coordinateGetter: sortableKeyboardCoordinates}),
  );

  const handleDragEnd = (event: DragEndEvent) => {
    const {active, over} = event;
    if (!over || active.id === over.id || !onRowReorder) return;

    const oldIndex = rows.findIndex((row) => {
      const rowId = (row.original as any)?.id ?? row.id;
      return rowId === active.id;
    });
    const newIndex = rows.findIndex((row) => {
      const rowId = (row.original as any)?.id ?? row.id;
      return rowId === over.id;
    });

    if (oldIndex !== -1 && newIndex !== -1) {
      onRowReorder(oldIndex, newIndex);
    }
  };

  // spacer 모드일 때 empty row의 colSpan에 +1
  const spacerOffset = tableWidthMode === 'spacer' ? 1 : 0;

  const getTableRows = () => {
    const countTotalColumns =
      (onRowReorder ? 1 : 0) +
      table.getAllColumns().reduce((sum, item) => sum + 1 + (item.columns?.length ?? 0), 0) +
      spacerOffset;

    if (!rows || rows.length === 0) {
      return (
        <TableRow className="hover:bg-transparent cursor-default border-none">
          <TableCell
            colSpan={countTotalColumns}
            className="border-none py-10 text-center text-muted-foreground"
          >
            {emptyCustomMessage || 'No results.'}
          </TableCell>
        </TableRow>
      );
    }

    if (onRowReorder) {
      const sortableIds = rows.map((row) => (row.original as any)?.id ?? row.id);
      return (
        <SortableContext items={sortableIds} strategy={verticalListSortingStrategy}>
          {rows.map((row, index) => (
            <SortableDataTableRow
              key={`table-row-${index}`}
              row={row}
              depth={row.depth}
              enableExpandingColumn={enableExpandingColumn}
              treeColumnId={treeColumnId}
              dataRowOnClick={dataRowOnClick}
              dataCellOnClick={dataCellOnClick}
              rowCursor={rowCursor}
              tdHeight={tdHeight}
              excludeRowClickColumns={excludeRowClickColumns}
              isLastRow={index === rows.length - 1}
              tableWidthMode={tableWidthMode}
            />
          ))}
        </SortableContext>
      );
    }

    return (
      <>
        {rows.map((row, index) => (
          <DataTableRow
            key={`table-row-${index}`}
            row={row}
            depth={row.depth}
            enableExpandingColumn={enableExpandingColumn}
            treeColumnId={treeColumnId}
            dataRowOnClick={dataRowOnClick}
            dataCellOnClick={dataCellOnClick}
            rowCursor={rowCursor}
            tdHeight={tdHeight}
            excludeRowClickColumns={excludeRowClickColumns}
            isLastRow={index === rows.length - 1}
            tableWidthMode={tableWidthMode}
          />
        ))}
        {infiniteScroll && (
          <tr>
            <td colSpan={table.getAllColumns().length + spacerOffset}>
              <div
                ref={infiniteScroll.loadMoreRef}
                className="py-3 text-center text-sm text-muted-foreground"
              >
                {infiniteScroll.isFetching && <span>Loading...</span>}
                {!infiniteScroll.hasNextPage && !infiniteScroll.isFetching && (
                  <span>No more data</span>
                )}
              </div>
            </td>
          </tr>
        )}
      </>
    );
  };

  const isFlex = tableHeight === 'flex';

  // 테이블 너비 결정
  // - independent / fill-last: 컬럼 합계 px (최소 100%)
  // - spacer: 항상 100%
  const resolvedTableWidth =
    tableWidthMode === 'spacer'
      ? '100%'
      : `max(${table.getTotalSize()}px, 100%)`;

  const tableContent = (
    // border를 absolute overlay로 분리
    // → scroll 컨테이너가 border 영향을 받지 않아 컬럼이 딱 맞아도 가로 스크롤 없음
    <div
      style={{
        width: tableWidth || '100%',
        overscrollBehavior: 'contain',
      }}
      className={cn(
        'relative rounded-md',
        isFlex ? 'flex-1 min-h-0 flex flex-col h-full' : '',
        table.getState().columnSizingInfo.isResizingColumn && 'select-none',
        className,
      )}
    >
      {/* 순수 시각용 border — absolute라 레이아웃에 영향 없음, z-20으로 sticky header(z-10)보다 위에 */}
      <div className="absolute inset-0 rounded-md border pointer-events-none z-20" />
      <div
        style={{
          ...(isFlex ? {} : {height: tableHeight || '400px'}),
          backgroundColor: 'var(--background/30)',
        }}
        className={cn('rounded-md relative overflow-auto', isFlex ? 'flex-1 min-h-0 h-full' : '')}
      >
        <ScrollTable
          style={{
            width: resolvedTableWidth,
            tableLayout: 'fixed',
          }}
        >
          <TableHeader
            className={cn(`sticky top-0 z-10 h-${thHeight}`, headerClassName)}
            style={{
              transform: 'translateZ(0)',
              willChange: 'transform',
            }}
          >
            {(() => {
              const headerGroups = table.getHeaderGroups();
              return (
                <>
                  {headerGroups.map((headerGroup, groupIndex) => {
                    const filteredHeaders = headerGroup.headers.filter((header) => {
                      if (header.isPlaceholder) return false;
                      const isGroupHeader = header.subHeaders && header.subHeaders.length > 0;
                      const hasParent = header.column.parent;
                      const isIndependentColumn = !isGroupHeader && !hasParent;
                      const metaRowSpan = header.column.columnDef.meta?.rowSpan || 1;

                      if (groupIndex === 0) {
                        return isGroupHeader || isIndependentColumn;
                      } else {
                        if (isIndependentColumn && metaRowSpan > 1) return false;
                        return hasParent || (isIndependentColumn && metaRowSpan === 1);
                      }
                    });

                    return (
                      <TableRow
                        className={cn(
                          'bg-muted/60 [&>th]:border-b [&>th]:border-r [&>th:last-child]:border-r-0! [&>th:first-child]:pl-3 [&>th:last-child]:pr-3!',
                        )}
                        key={headerGroup.id}
                      >
                        {/* drag 핸들 컬럼 헤더 (onRowReorder 있을 때만) */}
                        {onRowReorder && groupIndex === 0 && (
                          <TableHead
                            className={`relative h-${thHeight}`}
                            style={{width: '33px', minWidth: '33px', maxWidth: '33px', padding: '0'}}
                            rowSpan={headerGroups.length}
                          />
                        )}
                        {enableExpandingColumn && groupIndex === 0 && (
                          <TableHead
                            className="border-collapse border-spacing-0 [&:has([role=checkbox])]:pr-0 *:[[role=checkbox]]:translate-y-0.5"
                            style={{width: '4%'}}
                            rowSpan={headerGroups.length}
                          >
                            <Button className="w-7 p-0" variant="default" onClick={setAllExpanding}>
                              {table.getIsAllRowsExpanded() ? (
                                <DoubleArrowRightIcon className="h-4 w-4 rotate-90 transition-all" />
                              ) : (
                                <DoubleArrowRightIcon className="transition-all" />
                              )}
                            </Button>
                          </TableHead>
                        )}

                        {filteredHeaders.map((header, headerIndex) => {
                          const currentSize = header.getSize();
                          const columnDef = header.column.columnDef;
                          const hasFixedSize = columnDef.maxSize !== undefined;

                          const metaRowSpan = columnDef.meta?.rowSpan || 1;
                          const isGroupHeader = header.subHeaders && header.subHeaders.length > 0;
                          const calculatedColSpan = isGroupHeader ? header.subHeaders.length : 1;

                          const isLastHeader = headerIndex === filteredHeaders.length - 1;
                          const isFillHeader = tableWidthMode === 'fill-last' && isLastHeader;
                          const thWidth = isFillHeader ? 'auto' : `${currentSize}px`;

                          return (
                            <TableHead
                              key={header.id}
                              colSpan={calculatedColSpan}
                              rowSpan={metaRowSpan}
                              className={cn(
                                `relative h-${thHeight} text-[13px] font-semibold text-foreground! ${calculatedColSpan > 1 && 'text-center'}`,
                                columnDef.meta?.cellClassName,
                              )}
                              style={{
                                width: thWidth,
                                minWidth: `${currentSize}px`,
                                maxWidth: hasFixedSize && !isFillHeader ? `${currentSize}px` : undefined,
                              }}
                            >
                              <div className="w-full flex items-center overflow-hidden">
                                <div className="flex-1 overflow-hidden text-ellipsis whitespace-nowrap">
                                  {flexRender(columnDef.header, header.getContext())}
                                </div>
                              </div>
                              {table.options.enableColumnResizing && (
                                <ColumnResizer header={header} />
                              )}
                            </TableHead>
                          );
                        })}

                        {/* spacer 모드: 헤더 끝에 빈 th로 나머지 공간 채움 */}
                        {tableWidthMode === 'spacer' && groupIndex === 0 && (
                          <TableHead
                            rowSpan={headerGroups.length}
                            className={`relative h-${thHeight}`}
                            style={{width: 'auto'}}
                          />
                        )}
                      </TableRow>
                    );
                  })}
                </>
              );
            })()}
          </TableHeader>
          <TableBody ref={tableBodyRef} className="h-auto">
            {getTableRows()}
          </TableBody>
        </ScrollTable>
      </div>
    </div>
  );

  if (onRowReorder) {
    return (
      <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
        {tableContent}
      </DndContext>
    );
  }

  return tableContent;
}
