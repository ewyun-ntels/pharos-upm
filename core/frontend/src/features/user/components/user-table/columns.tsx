import {ColumnDef, createColumnHelper} from '@tanstack/react-table';
import {Badge} from '@pharos/shared/components/ui';
import {formatRelativeTime} from '@lib/format-date';
import type {User} from '@pharos/shared/types/user';
import {sortRoles, getRoleBadgeDisplay} from '@features/user/utils/role-badge';

export function getUserColumns(): ColumnDef<User>[] {
  const columnHelper = createColumnHelper<User>();

  return [
    columnHelper.accessor('name', {
      id: 'name',
      header: 'Username',
      enableSorting: true,
      cell: (info) => <span className="font-medium">{info.getValue()}</span>,
    }),
    columnHelper.accessor('roles', {
      id: 'roles',
      header: 'Assignments',
      meta: {flex: 1},
      enableSorting: false,
      cell: (info) => {
        const roles = info.getValue();
        if (!roles || roles.length === 0) return <span className="text-muted-foreground">-</span>;
        const sorted = sortRoles(roles);
        return (
          <div className="flex gap-1 flex-wrap items-center">
            {sorted.map((role) => {
              const {isAttr, prefix, displayName} = getRoleBadgeDisplay(role);
              return (
                <Badge key={role.role} variant="secondary" className="text-xs">
                  {isAttr ? (
                    <span className="text-muted-foreground">Permission:</span>
                  ) : prefix ? (
                    <span className="text-muted-foreground">{prefix}:</span>
                  ) : null}
                  <span>{displayName}</span>
                </Badge>
              );
            })}
          </div>
        );
      },
      filterFn: (row, _, filterValue: string[]) => {
        const roles = row.original.roles;
        if (!roles || roles.length === 0) return true;
        return roles.some((r) => filterValue.includes(r.display_name));
      },
      getUniqueValues: (row) => row.roles?.map((r) => r.display_name) ?? [],
    }),
    columnHelper.display({
      id: 'status',
      header: 'Status',
      size: 110,
      cell: (info) => {
        const user = info.row.original;
        if (user.blocked) {
          return (
            <Badge variant="outline" className="text-red-500 border-red-500 text-xs">
              Blocked
            </Badge>
          );
        }
        if (user.temporary_blocked_expires_at) {
          return (
            <Badge variant="outline" className="text-orange-500 border-orange-500 text-xs">
              Temp Locked
            </Badge>
          );
        }
        return <Badge variant="outline" className="text-green-600 border-green-600 text-xs">Active</Badge>;
      },
    }),
    columnHelper.accessor('password_expired_at', {
      id: 'password_expired_at',
      header: 'Password expiration',
      size: 180,
      enableSorting: true,
      cell: (info) => {
        const value = info.getValue();
        if (!value) return <span className="text-muted-foreground">-</span>;
        const now = new Date();
        const expiry = new Date(value);
        const fiveMonthsLater = new Date(now);
        fiveMonthsLater.setMonth(fiveMonthsLater.getMonth() + 5);
        const isExpired = expiry < now;
        const isExpiringSoon = !isExpired && expiry <= fiveMonthsLater;
        const colorClass = isExpired
          ? 'text-red-500 border-red-500'
          : isExpiringSoon
            ? 'text-orange-500 border-orange-500'
            : 'text-muted-foreground border-border';
        return (
          <Badge variant="outline" className={`text-xs w-fit ${colorClass}`}>
            {formatRelativeTime(value)}
          </Badge>
        );
      },
    }),
  ] as ColumnDef<User>[];
}
