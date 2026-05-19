import {ArrowDownIcon, ArrowUpIcon, CaretSortIcon, EyeNoneIcon} from '@radix-ui/react-icons';
import {Column} from '@tanstack/react-table';

import {cn} from '../../lib';
import {Button} from '@pharos/shared/components/ui';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@pharos/shared/components/ui';
import React from 'react';

interface DataTableColumnHeaderProps<TData, TValue> extends React.HTMLAttributes<HTMLDivElement> {
  column: Column<TData, TValue>;
  title: string;
  hideOption?: boolean;
  sortOption?: boolean;
}

export function DataTableColumnHeader<TData, TValue>({
  column,
  title,
  hideOption = true,
  sortOption = true,
  className,
}: DataTableColumnHeaderProps<TData, TValue>) {
  if (!column.getCanSort()) {
    return <div className={cn(className)}>{title}</div>;
  }

  return (
    <div className={cn('w-full flex items-center', className)}>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            className={
              'flex-1 justify-between h-10 p-0 m-0 rounded-none text-[13px] hover:bg-transparent data-[state=open]:text-primary hover:text-primary active:text-primary'
            }
          >
            <div>{title}</div>
            {column.getIsSorted() === 'desc' ? (
              <ArrowDownIcon className="ml-2 h-4 w-4 shrink-0" />
            ) : column.getIsSorted() === 'asc' ? (
              <ArrowUpIcon className="ml-2 h-4 w-4 shrink-0" />
            ) : (
              <CaretSortIcon className="ml-2 h-4 w-4 shrink-0" />
            )}
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start">
          {sortOption && (
            <>
              <DropdownMenuItem
                onClick={() => {
                  column.getIsSorted() === false || column.getIsSorted() === 'desc'
                    ? column.toggleSorting(false)
                    : column.clearSorting();
                }}
              >
                <ArrowUpIcon className="mr-1 h-3 w-3 text-muted-foreground/70" />
                Asc
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => {
                  column.getIsSorted() === false || column.getIsSorted() === 'asc'
                    ? column.toggleSorting(true)
                    : column.clearSorting();
                }}
              >
                <ArrowDownIcon className="mr-1 h-3 w-3 text-muted-foreground/70" />
                Desc
              </DropdownMenuItem>
            </>
          )}
          {hideOption && (
            <>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => column.toggleVisibility(false)}>
                <EyeNoneIcon className="mr-1 h-3 w-3 text-muted-foreground/70" />
                Hide
              </DropdownMenuItem>{' '}
            </>
          )}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
