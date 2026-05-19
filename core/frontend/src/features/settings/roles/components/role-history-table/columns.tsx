import {ColumnDef} from '@tanstack/react-table';
import {Badge, Button} from '@pharos/shared/components/ui';
import type {RoleMetadataConfigEntity} from '@pharos/shared/types/role';
import {formatDateTime} from '@lib/format-date';

interface GetRoleHistoryColumnsProps {
  onRollback: (id: string) => void;
}

export function getRoleHistoryColumns({
  onRollback,
}: GetRoleHistoryColumnsProps): ColumnDef<RoleMetadataConfigEntity>[] {
  return [
    {
      id: 'saved-at',
      accessorKey: 'created_at',
      header: 'Saved at',
      size: 200,
      cell: ({getValue}) => {
        const v = getValue() as string;
        return formatDateTime(v);
      },
    },
    {
      id: 'changed-by',
      accessorKey: 'updated_by',
      header: 'Changed by',
      meta: {flex: 1},
      cell: ({getValue}) => {
        const v = getValue() as string;
        return v ? (
          <span className="font-medium">{v}</span>
        ) : (
          <span className="text-muted-foreground">-</span>
        );
      },
    },
    {
      id: 'overrides',
      accessorKey: 'config',
      header: 'Overrides',
      size: 110,
      cell: ({getValue}) => {
        const v = getValue() as RoleMetadataConfigEntity['config'];
        const count = Array.isArray(v) ? v.length : 0;
        return <Badge variant="secondary">{count}</Badge>;
      },
    },
    {
      id: 'rollback',
      header: '',
      size: 100,
      cell: ({row}) => (
        <Button size="sm" variant="outline" onClick={() => onRollback(row.original.id)}>
          Rollback
        </Button>
      ),
    },
  ];
}
