import {ColumnDef} from '@tanstack/react-table';
import {Button} from '@pharos/shared/components/ui';
import {formatDateTime} from '@lib/format-date';

export type UserMetadataConfigHistoryEntry = {
  ID: string;
  ConfigJSON: string;
  CreatedAt: string;
  UpdatedBy?: string | null;
  Description?: string | null;
};

interface GetColumnsProps {
  onRollback: (id: string) => void;
}

export function getUserFieldsHistoryColumns({
  onRollback,
}: GetColumnsProps): ColumnDef<UserMetadataConfigHistoryEntry>[] {
  return [
    {
      id: 'saved-at',
      accessorKey: 'CreatedAt',
      header: 'Saved at',
      size: 200,
      cell: ({getValue}) => {
        const v = getValue() as string;
        return formatDateTime(v);
      },
    },
    {
      id: 'changed-by',
      accessorKey: 'UpdatedBy',
      header: 'Changed by',
      meta: {flex: 1},
      cell: ({getValue}) => {
        const v = getValue() as string | undefined | null;
        return v ? (
          <span className="font-medium">{v}</span>
        ) : (
          <span className="text-muted-foreground">-</span>
        );
      },
    },
    {
      id: 'description',
      accessorKey: 'Description',
      header: 'Description',
      meta: {flex: 2},
      cell: ({getValue}) => {
        const v = getValue() as string | undefined | null;
        return v ? (
          <span>{v}</span>
        ) : (
          <span className="text-muted-foreground">-</span>
        );
      },
    },
    {
      id: 'rollback',
      header: '',
      size: 100,
      cell: ({row}) => (
        <Button size="sm" variant="outline" onClick={() => onRollback(row.original.ID)}>
          Rollback
        </Button>
      ),
    },
  ];
}
