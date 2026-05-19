'use client';

import {MixerHorizontalIcon} from '@radix-ui/react-icons';
import type {Table} from '@tanstack/react-table';

import {Button, buttonVariants} from '@pharos/shared/components/ui';
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@pharos/shared/components/ui';
import type {VariantProps} from 'class-variance-authority';
import {useTranslation} from 'react-i18next';

interface DataTableViewOptionsProps<TData> extends VariantProps<typeof buttonVariants> {
  table: Table<TData>;
  hideColumn?: string[];
}

function DataTableViewOptions<TData>({
  table,
  hideColumn,
  variant,
  size,
}: DataTableViewOptionsProps<TData>) {
  const {t} = useTranslation();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          aria-label="Toggle columns"
          variant={variant}
          size={size}
          className="hidden h-8 lg:flex"
        >
          <MixerHorizontalIcon className={size !== 'icon' ? '' : '' + 'size-4'} />
          {size !== 'icon' && t('label.common.view')}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-40">
        <DropdownMenuLabel>Toggle columns</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {table
          .getAllColumns()
          .filter(
            (column) =>
              typeof column.accessorFn !== 'undefined' &&
              !hideColumn?.includes(column.id) &&
              column.getCanHide(),
          )
          .map((column) => {
            return (
              <DropdownMenuCheckboxItem
                key={column.id}
                className="capitalize"
                checked={column.getIsVisible()}
                onCheckedChange={(value) => column.toggleVisibility(value)}
              >
                <span className="truncate">{column.id}</span>
              </DropdownMenuCheckboxItem>
            );
          })}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export type {DataTableViewOptionsProps};
export {DataTableViewOptions};
