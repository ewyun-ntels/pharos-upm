import {Header} from '@tanstack/react-table';
import React from 'react';
import {cn} from '../../lib';

interface DataTableColumnResizerProps<TData, TValue> extends React.HTMLAttributes<HTMLDivElement> {
  header: Header<TData, TValue>;
}

export function ColumnResizer<TData, TValue>({header}: DataTableColumnResizerProps<TData, TValue>) {
  if (!header.column.getCanResize()) return <></>;

  return (
    <div
      className={cn(
        'absolute top-0 right-0 cursor-col-resize w-1 h-full bg-transparent hover:bg-primary/30 hover:w-1',
      )}
      {...{
        onMouseDown: header.getResizeHandler(),
        onTouchStart: header.getResizeHandler(),
        // className: `absolute top-0 right-0 cursor-col-resize w-px h-full bg-gray- hover:bg-gray-700 hover:w-2`,
        style: {
          userSelect: 'none',
          touchAction: 'none',
        },
      }}
    />
  );
}
