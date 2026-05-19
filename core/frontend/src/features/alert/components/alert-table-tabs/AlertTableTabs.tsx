import React from 'react';
import {TableTabsTemplate, Tab, Tabs} from '@pharos/shared/components/template/table-tabs';
import {PageBreadcrumb} from '@components/breadcrumb';
import {DataGrid} from '@pharos/shared/components/ui-extension/data-grid/DataGrid';
import {useAlertRuleList, useAlertHistoryList} from '../../hooks';
import {useDelete} from '@/lib/data-provider';
import {ALERT_RESOURCES, ALERT_PROVIDER_NAME} from '@providers/alert-provider/types';
import {ColumnDef} from '@tanstack/react-table';
import type {AlertRule, AlertValue} from '@pharos/shared/types/alert';
import {Badge, Button} from '@pharos/shared/components/ui';
import {AlertCircle, Clock, CheckCircle2, Plus, Edit, Trash2} from 'lucide-react';
import {useNavigate, NavigateFunction} from 'react-router-dom';
import {IconButton} from '@shared/frontend/components/ui-extension/icon-button';
import {formatDateTime, formatDateTimeShort} from '@lib/format-date';

/**
 * Alert Rule Table Columns
 */
const getAlertRuleColumns = (navigate: NavigateFunction, onDelete: (id: string) => void): ColumnDef<AlertRule>[] => [
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
    accessorKey: 'alert_type',
    header: 'Type',
    cell: ({getValue}) => {
      const type = getValue() as string;
      return (
        <Badge variant="outline">
          {type === 'query' ? 'Query' : type === 'event-status' ? 'Event Status' : 'Event History'}
        </Badge>
      );
    },
  },
  {
    accessorKey: 'rule.datasource',
    accessorFn: (row) => {
      const rule = row.rule as any;
      return rule?.datasource || '-';
    },
    header: 'Datasource',
  },
  {
    accessorKey: 'status',
    header: 'Current Status',
    cell: ({getValue}) => {
      const statuses = getValue() as AlertValue[] | undefined;
      if (!statuses || statuses.length === 0) {
        return (
          <div className="flex items-center gap-1 text-muted-foreground">
            <CheckCircle2 className="h-4 w-4" />
            <span>Normal</span>
          </div>
        );
      }

      const alertingCount = statuses.filter((s) => s.status === 'alerting').length;
      if (alertingCount > 0) {
        return (
          <div className="flex items-center gap-1 text-destructive">
            <AlertCircle className="h-4 w-4" />
            <span>{alertingCount} alerting</span>
          </div>
        );
      }

      return (
        <div className="flex items-center gap-1 text-muted-foreground">
          <CheckCircle2 className="h-4 w-4" />
          <span>Normal</span>
        </div>
      );
    },
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
      const alertRule = row.original;
      const rule = alertRule.rule as any;
      const ruleId = rule?.id || rule?.rule_id;

      return (
        <div className="flex justify-center items-center gap-2">
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
              if (window.confirm(`Delete rule "${rule?.name}"?`)) {
                onDelete(ruleId);
              }
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
 * Alert History Table Columns
 */
const alertHistoryColumns: ColumnDef<AlertValue>[] = [
  {
    accessorKey: 'name',
    header: 'Alert Name',
    cell: ({getValue}) => <div className="font-medium">{getValue() as string}</div>,
  },
  {
    accessorKey: 'alert_type',
    header: 'Type',
    cell: ({getValue}) => {
      const type = getValue() as string;
      return (
        <Badge variant="outline">
          {type === 'query' ? 'Query' : type === 'event-status' ? 'Event Status' : 'Event History'}
        </Badge>
      );
    },
  },
  {
    accessorKey: 'severity',
    header: 'Severity',
    cell: ({getValue}) => {
      const severity = getValue() as string;
      const variant =
        severity === 'Critical'
          ? 'destructive'
          : severity === 'Major'
            ? 'default'
            : severity === 'Minor'
              ? 'secondary'
              : 'outline';
      return <Badge variant={variant}>{severity}</Badge>;
    },
  },
  {
    accessorKey: 'status',
    header: 'Status',
    cell: ({getValue}) => {
      const status = getValue() as string;
      return <Badge variant={status === 'alerting' ? 'destructive' : 'secondary'}>{status}</Badge>;
    },
  },
  {
    accessorKey: 'value',
    header: 'Value',
    cell: ({getValue}) => {
      const value = getValue() as number;
      return <span className="font-mono">{value.toFixed(2)}</span>;
    },
  },
  {
    accessorKey: 'timestamp',
    header: 'Timestamp',
    cell: ({getValue}) => {
      const date = getValue() as Date;
      return (
        <div className="flex items-center gap-1 text-sm">
          <Clock className="h-3 w-3" />
          {formatDateTime(date)}
        </div>
      );
    },
  },
];

/**
 * Alert Rules Tab Component
 */
function AlertRulesTab() {
  const navigate = useNavigate();
  const {
    query: {data: rulesData, isLoading: rulesLoading, refetch: refetchRules},
  } = useAlertRuleList({
    pagination: {currentPage: 1, pageSize: 50},
  });

  const {mutate: deleteRule} = useDelete();

  const handleCreateNew = () => {
    navigate('/alert/edit');
  };

  const handleDelete = (ruleId: string) => {
    deleteRule(
      {resource: ALERT_RESOURCES.RULE, id: ruleId, meta: {dataProviderName: ALERT_PROVIDER_NAME}},
      {onSuccess: () => refetchRules()},
    );
  };

  return (
    <DataGrid
      data={rulesData?.data || []}
      columns={getAlertRuleColumns(navigate, handleDelete)}
      tableKey="alert-rules-table"
      searchableColumns={['rule.name', 'rule.datasource']}
      useTableSorting={true}
      rowCursor={true}
      onRefresh={refetchRules}
      isRefreshing={rulesLoading}
      rightFilters={() => [
        <Button onClick={handleCreateNew} key="create">
          <Plus className="h-4 w-4" />
          Create Alert Rule
        </Button>,
      ]}
      emptyCustomMessage={rulesLoading ? 'Loading alert rules...' : 'No alert rules found'}
    />
  );
}

/**
 * Alert History Tab Component
 */
function AlertHistoryTab() {
  const {
    query: {data: historyData, isLoading: historyLoading, refetch: refetchHistory},
  } = useAlertHistoryList({
    count: 100,
  });

  return (
    <DataGrid
      data={historyData?.data || []}
      columns={alertHistoryColumns}
      tableKey="alert-history-table"
      searchableColumns={['name', 'alert_type']}
      useTableSorting={true}
      rowCursor={false}
      onRefresh={refetchHistory}
      isRefreshing={historyLoading}
      emptyCustomMessage={
        historyLoading
          ? 'Loading alert history...'
          : 'No alert history found. Backend pagination support required.'
      }
    />
  );
}

/**
 * Alert Table with Tabs (Rules & History)
 *
 * ⚠️ History 탭은 백엔드 페이징 지원 대기 중입니다.
 */
export function AlertTableTabs() {
  return (
    <TableTabsTemplate
      name="Alert Management"
      defaultTab="Rules"
      breadcrumb={<PageBreadcrumb />}
    >
      <Tabs>
        <Tab name="Rules">
          <div className={'p-4'}>
            <AlertRulesTab />
          </div>
        </Tab>

        <Tab name="History">
          <div className={'p-4'}>
            <AlertHistoryTab />
          </div>
        </Tab>
      </Tabs>
    </TableTabsTemplate>
  );
}
