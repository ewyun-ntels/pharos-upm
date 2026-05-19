/**
 * DataGridFilters
 * Dashboard Panel 모드에서 사용되는 필터 UI
 */

import React from 'react';
import { Table } from '@tanstack/react-table';
import { ReactNode } from 'react';

interface DataGridFiltersProps<T> {
  leftFilters?: (table: Table<T>) => ReactNode[];
  rightFilters?: (table: Table<T>) => ReactNode[];
  description?: string;
  table: Table<T>;
}

export function DataGridFilters<T>(props: DataGridFiltersProps<T>) {
  const { leftFilters, rightFilters, description, table } = props;

  const leftItems = leftFilters?.(table);
  const rightItems = rightFilters?.(table);

  if (!leftItems?.length && !rightItems?.length && !description) {
    return null;
  }

  return (
    <div className="flex flex-col gap-2 pt-3 pb-4">
      {/* Description - 별도 줄 */}
      {description && (
        <p className="text-sm text-muted-foreground">{description}</p>
      )}

      {/* Filters row */}
      {(leftItems?.length || rightItems?.length) && (
        <div className="flex items-center justify-between gap-2">
          {/* Left filters */}
          {leftItems && leftItems.length > 0 && (
            <div className="flex items-center gap-2">
              {leftItems.map((item, index) => (
                <React.Fragment key={`left-${index}`}>{item}</React.Fragment>
              ))}
            </div>
          )}

          {/* Right filters */}
          {rightItems && (
            <div className="flex items-center gap-2 ml-auto">
              {rightItems.map((item, index) => (
                <React.Fragment key={`right-${index}`}>{item}</React.Fragment>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
