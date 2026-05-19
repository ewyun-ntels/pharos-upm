import {ColumnDef} from '@tanstack/react-table';
import {Link} from 'react-router-dom';
import {Trash2} from 'lucide-react';
import {IconButton} from '@pharos/shared/components/ui-extension';
import type {ToggleTableRow} from './VariablesTable';

interface VariablesColumnsProps {
  dashboardId?: string;
  onDeleteRow: (rowId: string) => void;
}

export function getVariablesColumns({
  dashboardId,
  onDeleteRow,
}: VariablesColumnsProps): ColumnDef<ToggleTableRow>[] {
  return [
    {
      id: 'id',
      accessorKey: 'id',
      header: 'ID',
      size: 180,
      cell: ({row}) => (
        <Link
          to={`/dashboards/${dashboardId}/variables/edit?id=${row.original.id}`}
          className="hover:text-primary cursor-pointer underline text-sm"
        >
          {row.original.id}
        </Link>
      ),
    },
    {
      id: 'type',
      accessorKey: 'type',
      header: 'Variables type',
      size: 180,
      cell: ({getValue}) => {
        const value = getValue<string>();
        return <span className="text-sm">{value.charAt(0).toUpperCase() + value.slice(1)}</span>;
      },
    },
    {
      id: 'description',
      accessorKey: 'description',
      header: 'Description',
      meta: {flex:1},
      cell: ({getValue}) => <span className="text-sm">{getValue<string>()}</span>,
    },
    {
      id: 'delete',
      header: '',
      size: 60,
      cell: ({row}) => (
        <div className="flex justify-center">
          <IconButton
            variant="destructive"
            size="icon-xs"
            onClick={(e) => {
              e.stopPropagation();
              onDeleteRow(row.original.id);
            }}
            icon={<Trash2 />}
          >
            Delete
          </IconButton>
        </div>
      ),
    },
  ];
}
