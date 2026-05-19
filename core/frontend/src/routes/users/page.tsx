import React from 'react';
import {UsersTab, SessionsTab, HistoryTab} from '@features/user';
import {TableTabsTemplate, Tabs, Tab} from '@pharos/shared/components/template/table-tabs';
import {PageBreadcrumb} from '@components/breadcrumb';

export default function UsersPage() {
  return (
    <main className="flex flex-col w-full h-full">
      <TableTabsTemplate name="User Management" breadcrumb={<PageBreadcrumb />}>
        <Tabs>
          <Tab name="Users">
            <UsersTab />
          </Tab>
          <Tab name="Sessions">
            <SessionsTab />
          </Tab>
          <Tab name="History">
            <HistoryTab />
          </Tab>
        </Tabs>
      </TableTabsTemplate>
    </main>
  );
}
