import {ColumnDef} from '@tanstack/react-table';
import {Input, Switch, Badge} from '@pharos/shared/components/ui';
import type {RoleMetadata, AtomicEditState} from '../../types';

// 편집 상태를 row 데이터에 포함 → columns 재생성 없이 셀 값이 항상 최신
export type TableRow = RoleMetadata & {
  editState: AtomicEditState;
};

interface GetRoleConfigColumnsProps {
  handleAtomicEdit: (key: string, field: keyof AtomicEditState, value: string | boolean) => void;
}

export function getRoleConfigColumns({
  handleAtomicEdit,
}: GetRoleConfigColumnsProps): ColumnDef<TableRow>[] {
  return [
    {
      id: 'key',
      accessorKey: 'key',
      header: 'Key / Group',
      size: 250,
      cell: ({row}) => {
        const isComposite = row.original.roles && Object.keys(row.original.roles).length > 0;
        const isHidden = row.original.editState.hide;
        return (
          <div className="flex flex-col gap-1 py-0.5">
            <div className="flex items-center gap-1.5">
              <span className={`font-mono text-xs leading-tight${isHidden ? ' text-muted-foreground line-through' : ''}`}>{row.original.key}</span>
              {isComposite && (
                <Badge variant="default" className="text-xs shrink-0">
                  Composite
                </Badge>
              )}
              {isHidden && (
                <Badge variant="secondary" className="text-xs shrink-0">
                  Hidden
                </Badge>
              )}
            </div>
            <span className="font-mono text-xs text-muted-foreground">{row.original.group}</span>
          </div>
        );
      },
    },
    {
      id: 'displayName',
      accessorFn: (row) => row.editState.displayName,
      header: 'Display name',
      size: 170,
      cell: ({row}) => (
        <Input
          value={row.original.editState.displayName}
          onChange={(e) => handleAtomicEdit(row.original.key, 'displayName', e.target.value)}
          className="h-7 text-xs"
          onClick={(e) => e.stopPropagation()}
        />
      ),
    },
    {
      id: 'description',
      accessorFn: (row) => row.editState.description,
      meta: {flex: 1},
      header: 'Description',
      cell: ({row}) => (
        <Input
          value={row.original.editState.description}
          onChange={(e) => handleAtomicEdit(row.original.key, 'description', e.target.value)}
          className="h-7 text-xs"
          onClick={(e) => e.stopPropagation()}
        />
      ),
    },
    {
      id: 'visible',
      header: 'Visible',
      size: 72,
      cell: ({row}) => (
        <Switch
          checked={!row.original.editState.hide}
          onCheckedChange={(v) => handleAtomicEdit(row.original.key, 'hide', !v)}
          onClick={(e) => e.stopPropagation()}
        />
      ),
    },
  ];
}
