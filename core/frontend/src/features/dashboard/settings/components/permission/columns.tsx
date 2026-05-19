import { ColumnDef } from '@tanstack/react-table';
import { Badge, Button } from '@pharos/shared/components/ui';
import { SelectBox } from '@pharos/shared/components/ui-extension/select/box';
import { Trash2 } from '@pharos/shared/components';
import { ActionSchema } from '@pharos/shared/types/dashboard';

export interface PermissionData {
  subject: string;
  object?: string;
  action: string;
}

const PERMISSION_OPTIONS = [
  { value: ActionSchema.enum.viewer, label: 'Viewer' },
  { value: ActionSchema.enum.editor, label: 'Editor' },
  { value: ActionSchema.enum.owner, label: 'Owner' },
];

interface GetPermissionColumnsProps {
  dashboardId: string;
  isDeletePending: boolean;
  onActionChange: (subject: string, action: string) => void;
  onDelete: (subject: string) => void;
}

export function getPermissionColumns({
  isDeletePending,
  onActionChange,
  onDelete,
}: GetPermissionColumnsProps): ColumnDef<PermissionData>[] {
  return [
    {
      id: 'type',
      accessorKey: 'type',
      header: 'Type',
      size: 80,
      cell: () => (
        <div className="w-full flex justify-center items-center">
          <Badge variant="secondary">User</Badge>
        </div>
      ),
    },
    {
      id: 'subject',
      accessorKey: 'subject',
      header: 'Name',
      meta: { flex: 1 },
    },
    {
      id: 'action',
      accessorKey: 'action',
      header: 'Permission',
      size: 180,
      cell: ({ row }) => (
        <SelectBox
          size="full"
          className="h-8!"
          options={PERMISSION_OPTIONS}
          onChange={(e) => {
            if (row.original.action !== e) {
              onActionChange(row.original.subject, e);
            }
          }}
          value={row.original.action}
        />
      ),
    },
    {
      id: 'delete',
      header: '',
      size: 60,
      enableSorting: false,
      cell: ({ row }) => (
        <div className="flex justify-center">
          <Button
            variant="destructive"
            onClick={() => onDelete(row.original.subject)}
            disabled={isDeletePending}
            size={"icon-sm"}
            className="rounded-[var(--radius)]"
          >
            <Trash2/>
          </Button>
        </div>
      ),
    },
  ];
}
