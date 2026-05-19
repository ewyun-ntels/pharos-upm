import React from 'react';
import {ColumnDef} from '@tanstack/react-table';
import {Badge, Button} from '@pharos/shared/components/ui';
import {formatDateTime} from '@lib/format-date';

export type DashboardHistoryRecord = {
  id: string;
  dashboardId: string;
  dashboardTitle: string;
  action: 'created' | 'updated' | 'deleted';
  changedBy: string;
  changedAt: string;
  configJson?: string;
};

function ActionBadge({action}: {action: string}) {
  if (action === 'created') {
    return (
      <Badge variant="outline" className="text-green-600 border-green-600">
        Created
      </Badge>
    );
  }
  if (action === 'updated') {
    return (
      <Badge variant="outline" className="text-blue-500 border-blue-500">
        Updated
      </Badge>
    );
  }
  if (action === 'deleted') {
    return (
      <Badge variant="outline" className="text-red-500 border-red-500">
        Deleted
      </Badge>
    );
  }
  return <Badge variant="outline">{action}</Badge>;
}

export function getDashboardHistoryColumns({
  onRollback,
}: {
  onRollback: (record: DashboardHistoryRecord) => void;
}): ColumnDef<DashboardHistoryRecord>[] {
  return [
    {
      id: 'changedAt',
      accessorKey: 'changedAt',
      header: 'Changed at',
      size: 200,
      cell: ({getValue}) => {
        const v = getValue() as string;
        return formatDateTime(v);
      },
    },
    {
      id: 'action',
      accessorKey: 'action',
      header: 'Action',
      size: 100,
      cell: ({getValue}) => <ActionBadge action={getValue() as string} />,
    },
    {
      id: 'dashboardTitle',
      accessorKey: 'dashboardTitle',
      header: 'Dashboard',
      meta: {flex: 1},
      minSize: 150,
    },
    {
      id: 'dashboardId',
      accessorKey: 'dashboardId',
      header: 'UUID',
      size: 290,
      cell: ({getValue}) => (
        <span className="font-mono text-xs text-muted-foreground">{getValue() as string}</span>
      ),
    },
    {
      id: 'changedBy',
      accessorKey: 'changedBy',
      header: 'Changed by',
      size: 160,
    },
    {
      id: 'rollback',
      header: '',
      size: 110,
      cell: ({row}) => {
        const record = row.original;
        return (
          <Button
            size="xs"
            variant="outline"
            onClick={() => onRollback(record)}
          >
            Rollback
          </Button>
        );
      },
    },
  ];
}

/** @deprecated Use getDashboardHistoryColumns({onRollback}) instead */
export const dashboardHistoryColumns = getDashboardHistoryColumns({onRollback: () => {}});
