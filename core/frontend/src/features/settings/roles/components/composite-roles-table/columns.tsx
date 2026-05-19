import {ColumnDef} from '@tanstack/react-table';
import {Badge, Button} from '@pharos/shared/components/ui';
import {Pencil, Trash2} from 'lucide-react';
import type {RoleMetadata} from '../../types';

interface GetCompositeRoleColumnsProps {
  allAtomicRoles: RoleMetadata[];
  isSaving: boolean;
  onEdit: (role: RoleMetadata) => void;
  onDelete: (key: string) => void;
}

export function getCompositeRoleColumns({
  allAtomicRoles,
  isSaving,
  onEdit,
  onDelete,
}: GetCompositeRoleColumnsProps): ColumnDef<RoleMetadata>[] {
  return [
    {
      id: 'composite-role',
      accessorKey: 'displayName',
      meta: {flex: 1},
      enableResizing: false,
      header: 'Composite roles',
      cell: ({row}) => {
        const cr = row.original;
        const subRoleKeys = Object.keys(cr.roles ?? {});
        const isComposite = cr.roles && Object.keys(cr.roles).length > 0;
        return (
          <div className="flex flex-col gap-3 py-3 px-1 border-b last:border-b-0">
            {/* Header */}
            <div className="flex items-start justify-between gap-2">
              <div className="space-y-1.5 min-w-0 flex-1">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="text-base font-semibold">{cr.displayName}</span>
                  <Badge variant="outline" className="font-mono text-xs shrink-0">
                    {cr.key}
                  </Badge>
                  {isComposite && (
                    <Badge variant="default" className="text-xs shrink-0">
                      Composite
                    </Badge>
                  )}
                  {cr.group && (
                    <Badge variant="secondary" className="font-mono text-xs shrink-0">
                      {cr.group}
                    </Badge>
                  )}
                  {(cr.hide ?? false) && (
                    <Badge variant="secondary" className="text-xs shrink-0">
                      Hidden
                    </Badge>
                  )}
                </div>
                {cr.description && (
                  <p className="text-xs text-muted-foreground leading-relaxed">
                    {cr.description}
                  </p>
                )}
              </div>
              <div className="flex gap-1 shrink-0">
                <Button
                  size="sm"
                  variant="ghost"
                  className="h-8 w-8 p-0"
                  onClick={(e) => {
                    e.stopPropagation();
                    onEdit(cr);
                  }}
                  disabled={isSaving}
                >
                  <Pencil className="h-3.5 w-3.5" />
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  className="h-8 w-8 p-0 text-destructive hover:text-destructive"
                  onClick={(e) => {
                    e.stopPropagation();
                    onDelete(cr.key);
                  }}
                  disabled={isSaving}
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </Button>
              </div>
            </div>
            {/* Base Roles */}
            <div className="space-y-1.5">
              <span className="text-xs text-muted-foreground">
                Base Roles ({subRoleKeys.length})
              </span>
              <div className="flex flex-wrap gap-1.5">
                {subRoleKeys.map((subKey) => {
                  const meta = allAtomicRoles.find((r) => r.key === subKey);
                  return (
                    <Badge key={subKey} variant="secondary" className="text-xs gap-1.5">
                      <span>{meta?.displayName ?? subKey}</span>
                      <span className="font-mono text-muted-foreground opacity-70">
                        {subKey}
                      </span>
                    </Badge>
                  );
                })}
              </div>
            </div>
          </div>
        );
      },
    },
  ];
}
