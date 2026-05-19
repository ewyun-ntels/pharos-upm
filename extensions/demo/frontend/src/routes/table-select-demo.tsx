import {useState} from 'react';
import {ColumnDef} from '@tanstack/react-table';
import {TableSelectTemplate, Options, Option} from '@pharos/shared/components/template/table-select';
import {TablePagination} from '@pharos/shared/components/template/table-tabs';
import { DataGrid } from '@pharos/shared/components/ui-extension/data-grid';
import {Button} from '@pharos/shared/components/ui';
import {Input} from '@pharos/shared/components/ui';
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '@pharos/shared/components/ui';
import {Plus, Download, Bell, Mail, Webhook} from 'lucide-react';

// Sample data types
interface AlertRule {
  id: string;
  name: string;
  type: 'threshold' | 'anomaly' | 'pattern';
  datasource: string;
  status: 'active' | 'inactive' | 'error';
  severity: 'critical' | 'warning' | 'info';
  lastTriggered: string;
}

interface Notification {
  id: string;
  name: string;
  type: 'email' | 'slack' | 'webhook';
  destination: string;
  enabled: boolean;
  alertsLinked: number;
  createdAt: string;
}

interface DashboardItem {
  id: string;
  name: string;
  category: string;
  owner: string;
  lastModified: string;
  views: number;
  starred: boolean;
}

// Sample data generators
const generateAlertRules = (count: number): AlertRule[] => {
  const types: AlertRule['type'][] = ['threshold', 'anomaly', 'pattern'];
  const statuses: AlertRule['status'][] = ['active', 'inactive', 'error'];
  const severities: AlertRule['severity'][] = ['critical', 'warning', 'info'];
  const datasources = ['Prometheus', 'ClickHouse', 'PostgreSQL', 'MySQL', 'InfluxDB'];

  return Array.from({length: count}, (_, i) => ({
    id: `rule-${i + 1}`,
    name: `Alert Rule ${i + 1}`,
    type: types[i % types.length],
    datasource: datasources[i % datasources.length],
    status: statuses[i % statuses.length],
    severity: severities[i % severities.length],
    lastTriggered: new Date(2025, 9, 14 - i).toLocaleString(),
  }));
};

const generateNotifications = (count: number): Notification[] => {
  const types: Notification['type'][] = ['email', 'slack', 'webhook'];
  const destinations = [
    'team@example.com',
    '#alerts-channel',
    'https://webhook.site/abc123',
    'admin@example.com',
    '#monitoring',
    'https://api.service.com/webhook',
  ];

  return Array.from({length: count}, (_, i) => ({
    id: `notif-${i + 1}`,
    name: `Notification ${i + 1}`,
    type: types[i % types.length],
    destination: destinations[i % destinations.length],
    enabled: i % 3 !== 0,
    alertsLinked: Math.floor(Math.random() * 10) + 1,
    createdAt: new Date(2025, 9, 14 - i).toLocaleDateString(),
  }));
};

const generateDashboards = (category: string, count: number): DashboardItem[] => {
  const owners = ['Alice', 'Bob', 'Charlie', 'Diana', 'Eve'];

  return Array.from({length: count}, (_, i) => ({
    id: `dash-${category}-${i + 1}`,
    name: `${category} Dashboard ${i + 1}`,
    category,
    owner: owners[i % owners.length],
    lastModified: new Date(2025, 9, 14 - i).toLocaleDateString(),
    views: Math.floor(Math.random() * 1000) + 50,
    starred: i % 4 === 0,
  }));
};

// Column definitions
const alertRuleColumns: ColumnDef<AlertRule>[] = [
  {
    accessorKey: 'name',
    header: 'Rule Name',
    cell: ({getValue}) => <span className="font-medium">{getValue<string>()}</span>,
  },
  {
    accessorKey: 'type',
    header: 'Type',
    cell: ({getValue}) => {
      const type = getValue<string>();
      const typeColors = {
        threshold: 'bg-blue-100 text-blue-800',
        anomaly: 'bg-purple-100 text-purple-800',
        pattern: 'bg-green-100 text-green-800',
      };
      return (
        <span className={`px-2 py-1 rounded-md text-xs font-medium ${typeColors[type as keyof typeof typeColors]}`}>
          {type}
        </span>
      );
    },
  },
  {
    accessorKey: 'datasource',
    header: 'Data Source',
  },
  {
    accessorKey: 'severity',
    header: 'Severity',
    cell: ({getValue}) => {
      const severity = getValue<string>();
      const severityColors = {
        critical: 'bg-red-100 text-red-800',
        warning: 'bg-yellow-100 text-yellow-800',
        info: 'bg-gray-100 text-gray-800',
      };
      return (
        <span
          className={`px-2 py-1 rounded-full text-xs font-medium ${severityColors[severity as keyof typeof severityColors]}`}
        >
          {severity}
        </span>
      );
    },
  },
  {
    accessorKey: 'status',
    header: 'Status',
    cell: ({getValue}) => {
      const status = getValue<string>();
      const statusColors = {
        active: 'text-green-600',
        inactive: 'text-gray-600',
        error: 'text-red-600',
      };
      return <span className={`font-medium ${statusColors[status as keyof typeof statusColors]}`}>{status}</span>;
    },
  },
  {
    accessorKey: 'lastTriggered',
    header: 'Last Triggered',
    cell: ({getValue}) => <span className="text-sm text-muted-foreground">{getValue<string>()}</span>,
  },
];

const notificationColumns: ColumnDef<Notification>[] = [
  {
    accessorKey: 'name',
    header: 'Notification Name',
    cell: ({getValue}) => <span className="font-medium">{getValue<string>()}</span>,
  },
  {
    accessorKey: 'type',
    header: 'Type',
    cell: ({getValue}) => {
      const type = getValue<string>();
      const icons = {
        email: <Mail className="w-4 h-4" />,
        slack: <Bell className="w-4 h-4" />,
        webhook: <Webhook className="w-4 h-4" />,
      };
      const typeColors = {
        email: 'bg-blue-100 text-blue-800',
        slack: 'bg-purple-100 text-purple-800',
        webhook: 'bg-orange-100 text-orange-800',
      };
      return (
        <span
          className={`flex items-center gap-1 px-2 py-1 rounded-md text-xs font-medium ${typeColors[type as keyof typeof typeColors]}`}
        >
          {icons[type as keyof typeof icons]}
          {type}
        </span>
      );
    },
  },
  {
    accessorKey: 'destination',
    header: 'Destination',
    cell: ({getValue}) => <span className="font-mono text-sm">{getValue<string>()}</span>,
  },
  {
    accessorKey: 'alertsLinked',
    header: 'Alerts Linked',
    cell: ({getValue}) => (
      <span className="bg-blue-50 text-blue-700 px-2 py-1 rounded-md text-sm font-medium">
        {getValue<number>()}
      </span>
    ),
  },
  {
    accessorKey: 'enabled',
    header: 'Enabled',
    cell: ({getValue}) => {
      const enabled = getValue<boolean>();
      return (
        <span className={`font-medium ${enabled ? 'text-green-600' : 'text-gray-400'}`}>
          {enabled ? 'Yes' : 'No'}
        </span>
      );
    },
  },
  {
    accessorKey: 'createdAt',
    header: 'Created At',
    cell: ({getValue}) => <span className="text-sm text-muted-foreground">{getValue<string>()}</span>,
  },
];

const dashboardColumns: ColumnDef<DashboardItem>[] = [
  {
    accessorKey: 'name',
    header: 'Dashboard Name',
    cell: ({getValue, row}) => {
      const starred = row.original.starred;
      return (
        <div className="flex items-center gap-2">
          {starred && <span className="text-yellow-500">⭐</span>}
          <span className="font-medium">{getValue<string>()}</span>
        </div>
      );
    },
  },
  {
    accessorKey: 'owner',
    header: 'Owner',
  },
  {
    accessorKey: 'views',
    header: 'Views',
    cell: ({getValue}) => (
      <span className="bg-gray-100 text-gray-800 px-2 py-1 rounded-md text-sm font-medium">
        {getValue<number>().toLocaleString()}
      </span>
    ),
  },
  {
    accessorKey: 'lastModified',
    header: 'Last Modified',
    cell: ({getValue}) => <span className="text-sm text-muted-foreground">{getValue<string>()}</span>,
  },
];

export function TableSelectDemo() {
  // Data states
  const [alertRules] = useState<AlertRule[]>(generateAlertRules(12));
  const [notifications] = useState<Notification[]>(generateNotifications(8));
  const [personalDashboards] = useState<DashboardItem[]>(generateDashboards('Personal', 15));
  const [teamDashboards] = useState<DashboardItem[]>(generateDashboards('Team', 10));
  const [publicDashboards] = useState<DashboardItem[]>(generateDashboards('Public', 20));

  // Filter states
  const [searchTerm, setSearchTerm] = useState('');
  const [filterValue, setFilterValue] = useState('all');

  return (
    <main className="flex flex-col w-full h-full">
      <TableSelectTemplate
        title="Table Select Template Demo"
        description="Select different content types using the dropdown - demonstrates full-page template usage"
        selectLabel="Content Type"
        selectPlaceholder="Select content type"
        selectWidth="240px"
        defaultValue="Alert Rules"
        onSelectionChange={(value) => console.log('Selected:', value)}
        leftCustomItems={
          <>
            <Input
              placeholder="Search..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-48 h-8"
            />
            <Select value={filterValue} onValueChange={setFilterValue}>
              <SelectTrigger className="w-36 h-8">
                <SelectValue placeholder="Filter" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectLabel>Filter Options</SelectLabel>
                  <SelectItem value="all">All</SelectItem>
                  <SelectItem value="active">Active</SelectItem>
                  <SelectItem value="inactive">Inactive</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </>
        }
        rightCustomItems={
          <>
            <Button variant="outline" size="sm" className="h-8">
              <Download className="w-4 h-4 mr-2" />
              Export
            </Button>
            <Button variant="default" size="sm" className="h-8">
              <Plus className="w-4 h-4 mr-2" />
              Create New
            </Button>
          </>
        }
      >
        <Options>
          <Option value="alert-rules" label="Alert Rules">
            <DataGrid
              data={alertRules}
              columns={alertRuleColumns}
              tableKey="alert-rules-demo"
              searchableColumns={['name', 'datasource']}
              useTableSorting={true}
              rowCursor={true}
              dataCellOnClick={(cell) => console.log('Alert Rule clicked:', cell.getValue())}
            />
          </Option>
          
          <Option value="notifications" label="Notifications">
            <DataGrid
              data={notifications}
              columns={notificationColumns}
              tableKey="notifications-demo"
              searchableColumns={['name', 'destination']}
              useTableSorting={true}
              rowCursor={true}
              dataCellOnClick={(cell) => console.log('Notification clicked:', cell.getValue())}
            />
          </Option>
          
          <Option value="personal-dashboards" label="Personal Dashboards">
            <TablePagination
              data={personalDashboards}
              columns={dashboardColumns}
              tableKey="personal-dashboards-demo"
              onPageChange={(page) => console.log('Page:', page)}
              onPageSizeChange={(size) => console.log('PageSize:', size)}
              searchableColumns={['name', 'owner']}
              useTableSorting={true}
              rowCursor={true}
              dataCellOnClick={(cell) => console.log('Dashboard clicked:', cell.getValue())}
            />
          </Option>
          
          <Option value="team-dashboards" label="Team Dashboards">
            <DataGrid
              data={teamDashboards}
              columns={dashboardColumns}
              tableKey="team-dashboards-demo"
              searchableColumns={['name', 'owner']}
              useTableSorting={true}
              rowCursor={true}
            />
          </Option>
          
          <Option value="public-dashboards" label="Public Dashboards">
            <TablePagination
              data={publicDashboards}
              columns={dashboardColumns}
              tableKey="public-dashboards-demo"
              onPageChange={(page) => console.log('Page:', page)}
              onPageSizeChange={(size) => console.log('PageSize:', size)}
              searchableColumns={['name', 'owner']}
              useTableSorting={true}
              rowCursor={true}
            />
          </Option>
          
          <Option value="alert-history" label="Alert History">
            <div className="flex items-center justify-center h-full">
              <div className="text-center">
                <Bell className="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
                <h3 className="text-lg font-semibold mb-2">Alert History</h3>
                <p className="text-sm text-muted-foreground mb-4">
                  View and analyze historical alert triggers
                </p>
                <Button variant="default" size="sm">
                  <Download className="w-4 h-4 mr-2" />
                  Export History
                </Button>
              </div>
            </div>
          </Option>
        </Options>
      </TableSelectTemplate>
    </main>
  );
}
