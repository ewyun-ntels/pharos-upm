import React, { useMemo, useCallback } from 'react';
import { useList } from '@/lib/data-provider';
import { DataGrid } from '@pharos/shared/components/ui-extension/data-grid/DataGrid';
import { VariableSelect, VariableInput } from '@pharos/shared/components/ui-extension';
import { USER_PROVIDER_NAME, USER_RESOURCES } from '@providers/user-provider';
import type { User } from '@pharos/shared/types/user';
import { useNavigate } from 'react-router-dom';
import { useSheetStore } from '@components/sheet/sheetStore';
import { getUserColumns } from './columns';
import { UsersSettings } from '@features/user';
import { UserDetailSheet } from '@features/user';

export function UsersTab() {
  const navigate = useNavigate();
  const { setSheet } = useSheetStore();

  const {
    query: { data: userList, refetch, isLoading },
  } = useList<User>({
    dataProviderName: USER_PROVIDER_NAME,
    resource: USER_RESOURCES.USER,
    queryOptions: { retry: false, refetchOnMount: 'always' },
  });

  const openUserSheet = useCallback(
    (user: User) => {
      setSheet(
        <UserDetailSheet
          user={user}
          refetch={refetch}
          onEdit={(u) => navigate(`/users/edit?id=${encodeURIComponent(u.name)}`)}
        />,
      );
    },
    [setSheet, refetch, navigate],
  );

  const handleCreate = useCallback((_type?: 'create' | 'edit', _rowData?: unknown) => {
    navigate('/users/edit');
  }, [navigate]);

  const columns = useMemo(() => getUserColumns(), []);

  const tableData = useMemo(() => userList?.data ?? [], [userList]);

  const settingsItem = useMemo(
    () => (
      <UsersSettings
        selectedUsers={[]}
        setShowAlert={() => { }}
        userListRefetch={refetch}
        handelEditDialogOpen={handleCreate}
      />
    ),
    [refetch, handleCreate],
  );

  return (
    <div className="px-5">
      <DataGrid<User>
        tableKey="users"
        data={tableData}
        columns={columns}
        useTableSorting={true}
        onRowClick={openUserSheet}
        rowCursor={true}
        leftFilters={(table) => {
          const roleOptions = Array.from(
            table.getColumn('roles')?.getFacetedUniqueValues()?.keys() ?? [],
          ).map((v: string) => ({ label: v, value: v }));

          return [
            <VariableInput
              key="username-filter"
              label="Username"
              placeholder="Search..."
              onChange={(v) => {
                table.getColumn('name')?.setFilterValue(v || undefined);
              }}
            />,
            <VariableSelect
              key="role-filter"
              label="Assignments"
              options={roleOptions}
              multiple={true}
              showAllOption={true}
              defaultAllSelected={true}
              onChange={(v) => {
                const values = v as string[];
                table
                  .getColumn('roles')
                  ?.setFilterValue(values.length ? values : undefined);
              }}
            />,
          ];
        }}
        rightFilters={() => [settingsItem]}
        isLoading={isLoading}
      />
    </div>
  );
}
