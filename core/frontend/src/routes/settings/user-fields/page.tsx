import React from 'react';
import {SchemaEditorTab, UserFieldsHistoryTab} from '@features/settings';
import {TableTabsTemplate, Tabs, Tab} from '@pharos/shared/components/template/table-tabs';
import {PageBreadcrumb} from '@components/breadcrumb';

export default function UserFieldsSettingsPage() {
  return (
    <main className="flex flex-col w-full h-full">
      <TableTabsTemplate name="User Fields Configuration" breadcrumb={<PageBreadcrumb />}>
        <Tabs>
          <Tab name="Schema Editor">
            <SchemaEditorTab />
          </Tab>
          <Tab name="History">
            <UserFieldsHistoryTab />
          </Tab>
        </Tabs>
      </TableTabsTemplate>
    </main>
  );
}
