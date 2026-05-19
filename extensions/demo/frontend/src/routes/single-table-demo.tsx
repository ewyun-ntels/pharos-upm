import {useState} from 'react';
import {ColumnDef} from '@tanstack/react-table';
import {
  TableTabsTemplate,
  Tabs,
  Tab,
} from '@pharos/shared/components/template/table-tabs';
import { DataGrid } from '@pharos/shared/components/ui-extension/data-grid';

interface User {
  id: string;
  name: string;
  email: string;
  role: string;
}

// Sample data
const generateUsers = (count: number): User[] => {
  return Array.from({ length: count }, (_, i) => ({
    id: `user-${i + 1}`,
    name: `User ${i + 1}`,
    email: `user${i + 1}@example.com`,
    role: ['Admin', 'User', 'Moderator'][i % 3],
  }));
};

// Column definitions
const userColumns: ColumnDef<User>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
  },
  {
    accessorKey: 'email',
    header: 'Email',
  },
  {
    accessorKey: 'role',
    header: 'Role',
  },
];

export function SingleTableDemo() {
  const [users] = useState<User[]>(generateUsers(20));

  return (
    <main className="flex flex-col w-full h-full">
      <TableTabsTemplate
        name="User Management"
        description="Single table interface without tab navigation"
        showTabs={false}
      >
        <Tabs>
          <Tab name="Users">
            <DataGrid
              data={users}
              columns={userColumns}
              tableKey="single-users-table"
              searchableColumns={['name', 'email']}
              useTableSorting={true}
              rowCursor={true}
            />
          </Tab>
        </Tabs>
      </TableTabsTemplate>
    </main>
  );
}
