import React from 'react';
import {TableTabsTemplate, Tab, Tabs} from '@pharos/shared/components/template/table-tabs';
import {PageBreadcrumb} from '@components/breadcrumb';
import {DataGrid} from '@pharos/shared/components/ui-extension/data-grid/DataGrid';
import {useNotificationRuleList} from '../../hooks';
import {ColumnDef} from '@tanstack/react-table';
import type {NotificationRule} from '@pharos/shared/types/notification';
import {Badge, Button} from '@pharos/shared/components/ui';
import {Plus, Edit, Trash2} from 'lucide-react';
import {useNavigate, NavigateFunction} from 'react-router-dom';
import {IconButton} from '@shared/frontend/components/ui-extension/icon-button';
import {formatDateTimeShort} from '@lib/format-date';

/**
 * Notification Rule Table Columns
 */
const getNotificationRuleColumns = (navigate: NavigateFunction): ColumnDef<NotificationRule>[] => [
  {
    accessorKey: 'rule.name',
    accessorFn: (row) => {
      const rule = row.rule as any;
      return rule?.name || '-';
    },
    header: 'Name',
    cell: ({getValue}) => <div className="font-medium">{getValue() as string}</div>,
  },
  {
    accessorKey: 'notification_type',
    header: 'Type',
    cell: ({getValue}) => {
      const type = getValue() as string;
      const typeLabel =
        type === 'snmp'
          ? 'SNMP'
          : type === 'slack'
            ? 'Slack'
            : type === 'email'
              ? 'Email'
              : type === 'webhook'
                ? 'Webhook'
                : type === 'teams'
                  ? 'Teams'
                  : type;

      return <Badge variant="outline">{typeLabel}</Badge>;
    },
  },
  {
    accessorKey: 'rule.target',
    accessorFn: (row) => {
      const rule = row.rule as any;
      return rule?.target || '-';
    },
    header: 'Target',
  },
  {
    accessorKey: 'rule.description',
    accessorFn: (row) => {
      const rule = row.rule as any;
      return rule?.description || '-';
    },
    header: 'Description',
  },
  {
    accessorKey: 'updated_at',
    header: 'Last Updated',
    cell: ({getValue}) => {
      const date = getValue() as Date | undefined;
      if (!date) return '-';
      return formatDateTimeShort(date);
    },
  },
  {
    id: 'actions',
    header: 'Actions',
    cell: ({row}) => {
      const notificationRule = row.original;
      const rule = notificationRule.rule as any;
      const ruleId = rule?.id;

      return (
        <div className="flex items-center justify-center gap-2">
          <IconButton
            variant="ghost"
            size="icon-xs"
            onClick={() => navigate(`/alert/edit?id=${ruleId}`)}
            icon={<Edit />}
            disabled={!ruleId}
          >
            Edit
          </IconButton>
          <IconButton
            variant="ghost"
            size="icon-xs"
            onClick={() => {
              // TODO: Delete confirmation dialog
            }}
            icon={<Trash2 />}
            disabled={!ruleId}
          >
            Delete
          </IconButton>
        </div>
      );
    },
  },
];

/**
 * Notification Rules Tab Component
 */
function NotificationRulesTab() {
  const navigate = useNavigate();
  const {
    query: {data: rulesData, isLoading: rulesLoading, refetch: refetchRules},
  } = useNotificationRuleList({
    pagination: {currentPage: 1, pageSize: 50},
  });

  const handleCreateNew = () => {
    navigate('/notification/edit');
  };

  return (
    <DataGrid
      data={rulesData?.data || []}
      columns={getNotificationRuleColumns(navigate)}
      tableKey="notification-rules-table"
      searchableColumns={['rule.name', 'rule.target']}
      useTableSorting={true}
      rowCursor={true}
      onRefresh={refetchRules}
      isRefreshing={rulesLoading}
      rightFilters={() => [
        <Button onClick={handleCreateNew} key="create">
          <Plus className="h-4 w-4" />
          Create Notification Rule
        </Button>,
      ]}
      emptyCustomMessage={
        rulesLoading ? 'Loading notification rules...' : 'No notification rules found'
      }
    />
  );
}

/**
 * Notification Table with Tabs (Rules)
 */
export function NotificationTableTabs() {
  return (
    <TableTabsTemplate
      name="Notification Management"
      defaultTab="Rules"
      breadcrumb={<PageBreadcrumb />}
    >
      <Tabs>
        <Tab name="Rules">
          <div className={'p-4'}>
            <NotificationRulesTab />
          </div>
        </Tab>
      </Tabs>
    </TableTabsTemplate>
  );
}
