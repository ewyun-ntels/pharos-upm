import {ArrowDownIcon, ArrowUpIcon, CaretSortIcon} from '@radix-ui/react-icons';
import {Column} from '@tanstack/react-table';

import {cn} from '../../lib';
import React from 'react';

interface DataTableColumnSortHeaderProps<TData, TValue>
  extends React.HTMLAttributes<HTMLDivElement> {
  column: Column<TData, TValue>;
  title: string;
}

interface DataTableColumnDefaultHeader extends React.HTMLAttributes<HTMLDivElement> {
  title: string;
}

export function DataTableColumnSortHeader<TData, TValue>({
  column,
  title,
  className,
}: DataTableColumnSortHeaderProps<TData, TValue>) {
  let sortIcon = <CaretSortIcon className="h-4 w-4 shrink-0" />;
  if (column.getIsSorted() === false) {
    sortIcon = <CaretSortIcon className="h-4 w-4 shrink-0" />;
  } else if (column.getIsSorted() === 'asc') {
    sortIcon = <ArrowUpIcon className="h-4 w-4 shrink-0" />;
  } else if (column.getIsSorted() === 'desc') {
    sortIcon = <ArrowDownIcon className="h-4 w-4 shrink-0" />;
  }

  return (
    <div
      className={cn(
        className,
        `flex flex-1 items-center gap-2 text-[13px] p-0 mb-[-1px] rounded-none border-primary hover:text-primary active:text-primary dark:border-primary dark:hover:text-primary dark:active:text-primary cursor-pointer py-2`,
      )}
      onClick={() => {
        if (column.getIsSorted() === false) {
          column.toggleSorting(true);
        } else if (column.getIsSorted() === 'desc') {
          column.toggleSorting(false);
        } else if (column.getIsSorted() === 'asc') {
          column.clearSorting();
        }
      }}
    >
      {title} {sortIcon}
    </div>
  );
}

export function DataTableColumnDefaultHeader({title, className}: DataTableColumnDefaultHeader) {
  return (
    <div
      className={cn(
        'flex flex-1 items-center gap-2 text-[13px] justify-between p-0 mb-[-1px] rounded-none leading-tight',
        className,
      )}
    >
      {title}
    </div>
  );
}
