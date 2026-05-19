/**
 * DataGridTableView
 *
 * DataGrid의 테이블 렌더링 전용 컴포넌트
 * - DataTable 래퍼
 * - 테이블 관련 props만 관리
 */

import { useEffect } from 'react';
import { Table, Cell } from '@tanstack/react-table';
import { DataTable } from '@pharos/shared/components/ui-extension';
import { cn } from '../../../lib';

interface DataGridTableViewProps<T> {
  table: Table<T>;
  tableHeight?: string | number | 'flex';
  onRowClick?: (row: T) => void;
  dataCellOnClick?: (cell: Cell<T, unknown>) => void;
  onTableReady?: (table: Table<T>) => void;
  enableExpandingColumn?: boolean;
  /** Expand toggle을 주입할 컬럼 ID */
  treeColumnId?: string;
  excludeRowClickColumns?: string[];
  className?: string;
  emptyCustomMessage?: string;
  rowCursor?: boolean;
  tableWidthMode?: 'independent' | 'spacer' | 'fill-last';
  tableBodyRef?: React.RefObject<HTMLTableSectionElement | null>;
  /** 무한 스크롤 설정 */
  infiniteScroll?: {
    isFetching: boolean;
    loadMoreRef: React.RefObject<HTMLDivElement | null>;
    hasNextPage: boolean;
  };
  /** 행 드래그앤드롭 재정렬 콜백. 제공 시 GripVertical 핸들 컬럼이 자동 추가됨 */
  onRowReorder?: (fromIndex: number, toIndex: number) => void;
}

export function DataGridTableView<T>({
  table,
  tableHeight,
  onRowClick,
  dataCellOnClick,
  enableExpandingColumn,
  treeColumnId,
  excludeRowClickColumns,
  emptyCustomMessage,
  rowCursor,
  tableWidthMode,
  onTableReady,
  tableBodyRef,
  className,
  infiniteScroll,
  onRowReorder,
}: DataGridTableViewProps<T>) {

  // Table ready callback
  useEffect(() => {
    if (onTableReady) {
      onTableReady(table);
    }
  }, [onTableReady, table]);


  const isFlex = tableHeight === 'flex';

  return (
    <div role="region" aria-label="Data Grid" className={cn(className, isFlex && 'flex-1 min-h-0')}>
      <DataTable
        table={table}
        tableHeight={tableHeight}
        dataRowOnClick={onRowClick ? (row) => onRowClick(row.original) : undefined}
        dataCellOnClick={dataCellOnClick}
        enableExpandingColumn={enableExpandingColumn}
        treeColumnId={treeColumnId}
        excludeRowClickColumns={excludeRowClickColumns}
        emptyCustomMessage={emptyCustomMessage}
        rowCursor={rowCursor}
        tableWidthMode={tableWidthMode}
        tableBodyRef={tableBodyRef}
        infiniteScroll={infiniteScroll}
        onRowReorder={onRowReorder}
      />
    </div>
  );
}
