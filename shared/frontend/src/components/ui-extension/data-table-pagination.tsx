/**
 * 두 가지 스타일 지원:
 * - default: 첫/이전/다음/끝 버튼 스타일
 * - numbered: 페이지 번호 클릭 스타일 (shadcn 기반 커스텀)
 */

import * as React from 'react';
import {useState} from 'react';
import {DoubleArrowLeftIcon, DoubleArrowRightIcon} from '@radix-ui/react-icons';
import {ChevronLeftIcon, ChevronRightIcon, MoreHorizontalIcon} from 'lucide-react';
import {Table} from '@tanstack/react-table';
import {Button, buttonVariants} from '@pharos/shared/components/ui';
import {
  Select, 
  SelectContent, 
  SelectItem, 
  SelectTrigger, 
  SelectValue} from '@pharos/shared/components/ui';
import {cn} from '../../lib';


interface DataTablePaginationProps<TData> {
  table: Table<TData>;
  additionalPageSizeItem?: number[];
  totalCount?: number;
  isPageSizeView?: boolean;
  variant?: 'default' | 'numbered';
}

function generatePaginationNumbers(currentPage: number, totalPages: number): (number | 'ellipsis')[] {
  const delta = 2;

  // Guard for non-positive totalPages
  if (totalPages <= 0) {
    return [];
  }

  // If totalPages is small, just list all pages directly.
  // This loop is bounded by a small constant (here 7), so it is effectively O(1).
  if (totalPages <= 7) {
    const pages: number[] = [];
    for (let i = 1; i <= totalPages; i++) {
      pages.push(i);
    }
    return pages;
  }

  const result: (number | 'ellipsis')[] = [];

  // Always include the first page
  result.push(1);

  // Compute window around the current page, excluding first and last
  const windowStart = Math.max(2, currentPage - delta);
  const windowEnd = Math.min(totalPages - 1, currentPage + delta);

  // Add leading ellipsis if needed
  if (windowStart > 2) {
    result.push('ellipsis');
  }

  // Add the window pages
  for (let i = windowStart; i <= windowEnd; i++) {
    result.push(i);
  }

  // Add trailing ellipsis if needed
  if (windowEnd < totalPages - 1) {
    result.push('ellipsis');
  }

  // Always include the last page
  result.push(totalPages);

  return result;
}

// ============================================================================
// Shared Sub-components
// ============================================================================

function PageSizeSelect<TData>({pageSizeItem, table}: {pageSizeItem: number[]; table: Table<TData>}) {
  return (
    <Select
      value={`${table.getState().pagination.pageSize}`}
      onValueChange={(value) => table.setPageSize(Number(value))}
    >
      <SelectTrigger className="h-8 w-17.5 bg-background shadow-none cursor-pointer">
        <SelectValue placeholder={table.getState().pagination.pageSize} />
      </SelectTrigger>
      <SelectContent side="top">
        {pageSizeItem.map((pageSize) => (
          <SelectItem key={pageSize} value={`${pageSize}`}>
            {pageSize}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

function RowCount<TData>({table, totalCount, selectedRowCount}: {table: Table<TData>; totalCount?: number; selectedRowCount: number}) {
  return (
    <div className="text-xs text-muted-foreground">
      {(totalCount ?? table.getFilteredRowModel().rows.length).toLocaleString()} row(s)
      {selectedRowCount > 0 && (
        <span className="pl-2 text-foreground">{selectedRowCount} selected</span>
      )}
    </div>
  );
}

// ============================================================================
// Numbered Pagination UI Components (shadcn 기반 커스텀)
// ============================================================================

type PaginationLinkProps = {isActive?: boolean} & Pick<React.ComponentProps<typeof Button>, 'size'> & React.ComponentProps<'button'>;

function PaginationLink({className, isActive, size = 'icon', ...props}: PaginationLinkProps) {
  return (
    <button
      type="button"
      aria-current={isActive ? 'page' : undefined}
      className={cn(
        buttonVariants({variant: isActive ? 'outline' : 'ghost', size}), 'w-8 h-8', className)}
      {...props}
    />
  );
}

function PaginationPrevious({className, ...props}: React.ComponentProps<typeof PaginationLink>) {
  return (
    <PaginationLink 
        aria-label="Go to previous page" 
        size="icon" 
        className={cn('w-8 h-8', className)} {...props}>
      <ChevronLeftIcon className="h-4 w-4" />
      <span className="sr-only">Previous</span>
    </PaginationLink>
  );
}

function PaginationNext({className, ...props}: React.ComponentProps<typeof PaginationLink>) {
  return (
    <PaginationLink 
        aria-label="Go to next page" 
        size="icon" 
        className={cn('w-8 h-8', className)} {...props}>
      <ChevronRightIcon className="h-4 w-4" />
      <span className="sr-only">Next</span>
    </PaginationLink>
  );
}

// ============================================================================
// Main Component
// ============================================================================

export function DataTablePagination<TData>({
  table,
  additionalPageSizeItem = [],
  totalCount,
  isPageSizeView = true,
  variant = 'default',
}: DataTablePaginationProps<TData>) {
  const [pageSizeItem] = useState(() =>
    Array.from(new Set<number>([table.getState().pagination.pageSize, ...additionalPageSizeItem])).sort((a, b) => a - b),
  );

  const selectedRowCount = table.getSelectedRowModel().rows.length;
  const pageCount = table.getPageCount();
  const totalPageCount = Math.max(pageCount, 1);
  const currentPage = Math.min(table.getState().pagination.pageIndex + 1, totalPageCount);

  if (variant === 'numbered') {
    const pages = generatePaginationNumbers(currentPage, totalPageCount);
    const canPreviousPage = table.getCanPreviousPage();
    const canNextPage = table.getCanNextPage();
    return (
      <div className="pagination flex items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          {isPageSizeView && <PageSizeSelect pageSizeItem={pageSizeItem} table={table} />}
          <RowCount table={table} totalCount={totalCount} selectedRowCount={selectedRowCount} />
        </div>
        <nav role="navigation" aria-label="pagination">
          <ul className="flex flex-row items-center gap-1">
            <li>
              <PaginationPrevious
                onClick={canPreviousPage ? () => table.previousPage() : undefined}
                aria-disabled={!canPreviousPage}
                tabIndex={canPreviousPage ? 0 : -1}
                className={!canPreviousPage ? 'pointer-events-none opacity-50' : 'cursor-pointer'}
              />
            </li>
            {pages.map((page, idx) => (
              <li key={`${page}-${idx}`}>
                {page === 'ellipsis' ? (
                  <span className="flex w-8 h-8 items-center justify-center text-xs">
                    <MoreHorizontalIcon className="size-4" />
                  </span>
                ) : (
                  <PaginationLink
                    onClick={() => table.setPageIndex((page as number) - 1)}
                    isActive={currentPage === page}
                    className="cursor-pointer"
                  >
                    {page}
                  </PaginationLink>
                )}
              </li>
            ))}
            <li>
              <PaginationNext
                onClick={canNextPage ? () => table.nextPage() : undefined}
                aria-disabled={!canNextPage}
                tabIndex={canNextPage ? 0 : -1}
                className={!canNextPage ? 'pointer-events-none opacity-50' : 'cursor-pointer'}
              />
            </li>
          </ul>
        </nav>
      </div>
    );
  }

  return (
    <div className="pagination flex items-center justify-between">
      <div className="flex items-center gap-3">
        {isPageSizeView && <PageSizeSelect pageSizeItem={pageSizeItem} table={table} />}
        <RowCount table={table} totalCount={totalCount} selectedRowCount={selectedRowCount} />
      </div>
      <div className="flex items-center gap-4">
        <span className="text-xs">{`${currentPage.toLocaleString()} / ${totalPageCount.toLocaleString()}`}</span>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            className="hidden h-7 w-7 p-0 lg:flex shadow-none cursor-pointer"
            onClick={() => table.setPageIndex(0)}
            disabled={!table.getCanPreviousPage()}
          >
            <span className="sr-only">Go to first page</span>
            <DoubleArrowLeftIcon className="h-3.5 w-3.5" />
          </Button>
          <Button
            variant="ghost"
            className="h-7 w-7 p-0 shadow-none cursor-pointer"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            <span className="sr-only">Go to previous page</span>
            <ChevronLeftIcon className="h-3.5 w-3.5" />
          </Button>
          <Button
            variant="ghost"
            className="h-7 w-7 p-0 shadow-none cursor-pointer"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            <span className="sr-only">Go to next page</span>
            <ChevronRightIcon className="h-3.5 w-3.5" />
          </Button>
          <Button
            variant="ghost"
            className="hidden h-7 w-7 p-0 lg:flex shadow-none cursor-pointer"
            onClick={() => table.setPageIndex(pageCount - 1)}
            disabled={!table.getCanNextPage()}
          >
            <span className="sr-only">Go to last page</span>
            <DoubleArrowRightIcon className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>
    </div>
  );
}
