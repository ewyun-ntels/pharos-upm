import {useState} from 'react';
import {ColumnDef} from '@tanstack/react-table';
import {
  TableTabsTemplate,
  Tabs as TemplateTabs,
  Tab,
  TablePagination,
} from '@pharos/shared/components/template/table-tabs';
import {DataGrid, DataGridInfinity} from '@pharos/shared/components/ui-extension/data-grid';
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
  Badge,
  Button,
  TabsContent,
} from '@pharos/shared/components/ui';
import {
  SheetTemplate,
  SheetTemplateHeader,
  SheetTemplateTabs,
  SheetTemplateContent,
  SheetTemplateGroup,
  SheetTemplateGroupGrid,
  SheetTemplateGroupGridItem,
  SheetTemplateGroupTable,
  SheetTemplateGroupTab,
  SheetTemplateGroupChart,
  SheetTemplateFooter,
} from '@pharos/shared/components/template/sheet';
import {Copy, Ban, Pen, Trash2, X} from 'lucide-react';
import {useSheetStore} from '@components/sheet/sheetStore';
import {IconButton} from '@pharos/shared/components/ui-extension';

interface User {
  id: string;
  name: string;
  email: string;
  role: string;
  createdAt: string;
}

interface Product {
  id: string;
  name: string;
  price: number;
  category: string;
  stock: number;
}

interface Order {
  id: string;
  userId: string;
  productName: string;
  quantity: number;
  total: number;
  status: 'pending' | 'completed' | 'cancelled';
}

// Sample data
const generateUsers = (count: number): User[] => {
  return Array.from({length: count}, (_, i) => ({
    id: `user-${i + 1}`,
    name: `User ${i + 1}`,
    email: `user${i + 1}@example.com`,
    role: ['Admin', 'User', 'Moderator'][i % 3],
    createdAt: new Date(2024, 0, 15 + i).toLocaleDateString(),
  }));
};

const generateProducts = (count: number): Product[] => {
  const categories = ['Electronics', 'Books', 'Clothing', 'Home', 'Sports'];
  const baseNames = ['Laptop', 'Phone', 'Book', 'Shirt', 'Ball'];

  return Array.from({length: count}, (_, i) => ({
    id: `product-${i + 1}`,
    name: `${baseNames[i % baseNames.length]} ${i + 1}`,
    price: Math.floor(Math.random() * 1000) + 10,
    category: categories[i % categories.length],
    stock: Math.floor(Math.random() * 500) + 1,
  }));
};

const generateOrders = (count: number): Order[] => {
  const statuses: Order['status'][] = ['pending', 'completed', 'cancelled'];
  const productNames = ['Laptop Pro', 'Gaming Phone', 'Tech Book', 'Sports Shirt', 'Soccer Ball'];

  return Array.from({length: count}, (_, i) => ({
    id: `order-${i + 1}`,
    userId: `user-${(i % 10) + 1}`,
    productName: productNames[i % productNames.length],
    quantity: Math.floor(Math.random() * 5) + 1,
    total: Math.floor(Math.random() * 2000) + 50,
    status: statuses[i % statuses.length],
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
    cell: ({getValue}) => {
      const role = getValue<string>();
      const roleColors = {
        Admin: 'text-red-600 bg-red-50 px-2 py-1 rounded-md text-xs font-medium',
        Moderator: 'text-blue-600 bg-blue-50 px-2 py-1 rounded-md text-xs font-medium',
        User: 'text-green-600 bg-green-50 px-2 py-1 rounded-md text-xs font-medium',
      };
      return <span className={roleColors[role as keyof typeof roleColors] || ''}>{role}</span>;
    },
  },
  {
    accessorKey: 'createdAt',
    header: 'Created At',
  },
];

const productColumns: ColumnDef<Product>[] = [
  {
    accessorKey: 'name',
    header: 'Product Name',
  },
  {
    accessorKey: 'price',
    header: 'Price',
    cell: ({getValue}) => `$${getValue<number>().toLocaleString()}`,
  },
  {
    accessorKey: 'category',
    header: 'Category',
    cell: ({getValue}) => {
      const category = getValue<string>();
      return (
        <span className="bg-gray-100 text-gray-800 px-2 py-1 rounded-md text-xs font-medium">
          {category}
        </span>
      );
    },
  },
  {
    accessorKey: 'stock',
    header: 'Stock',
    cell: ({getValue}) => {
      const stock = getValue<number>();
      const stockColor =
        stock > 100 ? 'text-green-600' : stock > 50 ? 'text-yellow-600' : 'text-red-600';
      return <span className={stockColor}>{stock}</span>;
    },
  },
];

const orderColumns: ColumnDef<Order>[] = [
  {
    accessorKey: 'id',
    header: 'Order ID',
    cell: ({getValue}) => <span className="font-mono text-sm">{getValue<string>()}</span>,
  },
  {
    accessorKey: 'productName',
    header: 'Product',
  },
  {
    accessorKey: 'quantity',
    header: 'Quantity',
    cell: ({getValue}) => (
      <span className="bg-blue-50 text-blue-700 px-2 py-1 rounded-md text-sm font-medium">
        {getValue<number>()}
      </span>
    ),
  },
  {
    accessorKey: 'total',
    header: 'Total',
    cell: ({getValue}) => (
      <span className="font-semibold text-green-600">${getValue<number>().toLocaleString()}</span>
    ),
  },
  {
    accessorKey: 'status',
    header: 'Status',
    cell: ({getValue}) => {
      const status = getValue<string>();
      const statusStyles = {
        pending: 'bg-yellow-100 text-yellow-800',
        completed: 'bg-green-100 text-green-800',
        cancelled: 'bg-red-100 text-red-800',
      };
      return (
        <span
          className={`px-2 py-1 rounded-full text-xs font-medium ${statusStyles[status as keyof typeof statusStyles] || ''}`}
        >
          {status.charAt(0).toUpperCase() + status.slice(1)}
        </span>
      );
    },
  },
];

interface Activity {
  action: string;
  timestamp: string;
}

const activityColumns: ColumnDef<Activity>[] = [
  {
    accessorKey: 'action',
    header: '활동명',
  },
  {
    accessorKey: 'timestamp',
    header: '시간',
  },
];

const activityData: Activity[] = [
  { action: '권한 부여 (admin)', timestamp: '2026. 03. 10.' },
  { action: '비밀번호 변경', timestamp: '2026. 03. 01.' },
];

const chartDataDaily = [
  { name: 'Mon', value: 400 }, { name: 'Tue', value: 300 }, { name: 'Wed', value: 550 },
  { name: 'Thu', value: 450 }, { name: 'Fri', value: 700 }, { name: 'Sat', value: 600 },
  { name: 'Sun', value: 800 },
];

const chartDataWeekly = [
  { name: 'W1', value: 2400 }, { name: 'W2', value: 1398 },
  { name: 'W3', value: 9800 }, { name: 'W4', value: 3908 },
];

const chartDataMonthly = [
  { name: 'Jan', value: 4000 }, { name: 'Feb', value: 3000 },
  { name: 'Mar', value: 2000 }, { name: 'Apr', value: 2780 },
];

function DemoUserSheetTypeA({user}: {user: User}) {
  const {setClose} = useSheetStore();
  const [chartTab, setChartTab] = useState('daily');

  return (
    <SheetTemplate>
      <SheetTemplateHeader
        title={user.name}
        leftInfoGroup={<Badge variant="default">Active</Badge>}
        rightButtonGroup={
          <>
            <IconButton icon={<Pen className="h-4 w-4" />} variant="ghost" className="h-7 w-7">
              Edit
            </IconButton>
            <IconButton icon={<Copy className="h-4 w-4" />} variant="ghost" className="h-7 w-7">
              Copy
            </IconButton>
            <IconButton icon={<X className="h-4 w-4" />} variant="ghost" className="h-7 w-7 text-muted-foreground"  onClick={setClose}>
              Close
            </IconButton>
          </>
        }
      />

      <SheetTemplateContent>
        {/* 1단 grid */}
        <SheetTemplateGroup title="기본 정보">
          <SheetTemplateGroupGrid cols={1}>
            <SheetTemplateGroupGridItem cols={1} label="Name" value={user.name} />
            <SheetTemplateGroupGridItem cols={1} label="Email" value={user.email} />
            <SheetTemplateGroupGridItem cols={1} label="Role" value={user.role} />
            <SheetTemplateGroupGridItem cols={1} label="Created at" value={user.createdAt} />
          </SheetTemplateGroupGrid>
        </SheetTemplateGroup>

        {/* 2단 grid */}
        <SheetTemplateGroup title="상세 정보">
          <SheetTemplateGroupGrid cols={2}>
            <SheetTemplateGroupGridItem cols={2} label="Name" value={user.name} description="사용자 이름입니다." />
            <SheetTemplateGroupGridItem cols={2} label="Email" value={user.email} />
            <SheetTemplateGroupGridItem cols={2} label="Role" value={user.role} />
            <SheetTemplateGroupGridItem cols={2} label="Created at" value={user.createdAt} description="1개월 전" />
          </SheetTemplateGroupGrid>
        </SheetTemplateGroup>

        {/* table */}
        <SheetTemplateGroup title="활동 이력">
          <SheetTemplateGroupTable>
            <DataGrid
              enablePagination={false}
              tableHeight="auto"
              columns={activityColumns}
              data={activityData}
            />
          </SheetTemplateGroupTable>
        </SheetTemplateGroup>

        {/* chart only (단일 차트 케이스) */}
        <SheetTemplateGroup title="통계 요약 (Chart Only)">
          <SheetTemplateGroupChart>
            <div className="h-[200px] w-full">
              <div className="w-full h-full bg-slate-50 border border-slate-100 rounded-md flex items-end justify-around p-4">
                {chartDataDaily.map((d) => (
                  <div key={d.name} className="w-8 bg-slate-400 rounded-t-sm flex flex-col justify-end items-center text-xs text-white pb-1 transition-all" style={{ height: `${(d.value / 800) * 100}%` }}>
                    <span className="opacity-0 hover:opacity-100">{d.value}</span>
                  </div>
                ))}
              </div>
            </div>
          </SheetTemplateGroupChart>
        </SheetTemplateGroup>

        {/* chart + tabs (탭이 있는 차트 케이스) */}
        <SheetTemplateGroup title="상세 통계 (Tab + Chart)">
          <SheetTemplateGroupTab
            tabs={[
              {label: '일간', value: 'daily'},
              {label: '주간', value: 'weekly'},
              {label: '월간', value: 'monthly'},
            ]}
            activeTab={chartTab}
            onTabChange={setChartTab}
          >
            {/* 사용자는 TabsContent만 채우면 됩니다 */}
            <TabsContent value="daily">
              <div className="w-full h-full bg-blue-50/50 border border-blue-100 rounded-md flex items-end justify-around p-4">
                {chartDataDaily.map((d) => (
                  <div key={d.name} className="w-8 bg-blue-400 rounded-t-sm flex flex-col justify-end items-center text-xs text-white pb-1 transition-all" style={{ height: `${(d.value / 800) * 100}%` }}>
                    <span className="opacity-0 hover:opacity-100">{d.value}</span>
                  </div>
                ))}
              </div>
            </TabsContent>

            <TabsContent value="weekly" className="h-[250px] w-full">
              <div className="w-full h-full bg-green-50/50 border border-green-100 rounded-md flex items-end justify-around p-4">
                {chartDataWeekly.map((d) => (
                  <div key={d.name} className="w-16 bg-green-400 rounded-t-sm flex flex-col justify-end items-center text-xs text-white pb-1 transition-all" style={{ height: `${(d.value / 9800) * 100}%` }}>
                    <span className="opacity-0 hover:opacity-100">{d.value}</span>
                  </div>
                ))}
              </div>
            </TabsContent>

            <TabsContent value="monthly" className="h-[250px] w-full">
              <div className="w-full h-full bg-amber-50/50 border border-amber-100 rounded-md flex items-end justify-around p-4">
                {chartDataMonthly.map((d) => (
                  <div key={d.name} className="w-16 bg-amber-400 rounded-t-sm flex flex-col justify-end items-center text-xs text-white pb-1 transition-all" style={{ height: `${(d.value / 4000) * 100}%` }}>
                    <span className="opacity-0 hover:opacity-100">{d.value}</span>
                  </div>
                ))}
              </div>
            </TabsContent>
          </SheetTemplateGroupTab>
        </SheetTemplateGroup>
      </SheetTemplateContent>

      <SheetTemplateFooter>
        <Button
          variant="destructive"
          className="w-full flex items-center justify-center gap-2"
          onClick={() => {
            console.log('Delete clicked');
            setClose();
          }}
        >
          <Trash2 className="w-3.5 h-3.5" />
          Delete User
        </Button>
      </SheetTemplateFooter>
    </SheetTemplate>
  );
}

function DemoUserSheetTypeB({user}: {user: User}) {
  const {setClose} = useSheetStore();
  const [activeTab, setActiveTab] = useState('info');
  const [chartTab, setChartTab] = useState('daily');

  return (
    <SheetTemplate>
      <SheetTemplateHeader
        title={user.name}
        leftInfoGroup={<Badge variant="default">Active</Badge>}
        rightButtonGroup={
          <>
            <Button size="sm" variant="secondary" onClick={() => console.log('Blocked!')}>
              <Ban className="h-4 w-4" />
              Block
            </Button>
            <IconButton icon={<Pen className="h-4 w-4" />} variant="ghost" className="h-7 w-7">
              Edit
            </IconButton>
            <IconButton icon={<Copy className="h-4 w-4" />} variant="ghost" className="h-7 w-7">
              Copy
            </IconButton>
            <IconButton icon={<X className="h-4 w-4" />} variant="ghost" className="h-7 w-7 text-muted-foreground" onClick={setClose}>
              Close
            </IconButton>
          </>
        }
      />

      <SheetTemplateTabs
        tabs={[
          {label: '기본 정보', value: 'info'},
          {label: '상세 정보', value: 'detail'},
          {label: '활동 이력', value: 'activity'},
        ]}
        activeTab={activeTab}
        onTabChange={setActiveTab}
      />

      {activeTab === 'info' && (
        <SheetTemplateContent>
          <SheetTemplateGroup>
            <SheetTemplateGroupGrid cols={1}>
              <SheetTemplateGroupGridItem cols={1} label="Name" value={user.name} />
              <SheetTemplateGroupGridItem cols={1} label="Email" value={user.email} />
              <SheetTemplateGroupGridItem cols={1} label="Role" value={user.role} />
              <SheetTemplateGroupGridItem cols={1} label="Created at" value={user.createdAt} />
            </SheetTemplateGroupGrid>
          </SheetTemplateGroup>
        </SheetTemplateContent>
      )}

      {activeTab === 'detail' && (
        <SheetTemplateContent>
          <SheetTemplateGroup>
            <SheetTemplateGroupGrid cols={2}>
              <SheetTemplateGroupGridItem cols={2} label="Name" value={user.name} description="사용자 이름입니다." />
              <SheetTemplateGroupGridItem cols={2} label="Email" value={user.email} />
              <SheetTemplateGroupGridItem cols={2} label="Role" value={user.role} />
              <SheetTemplateGroupGridItem cols={2} label="Created at" value={user.createdAt} description="1개월 전" />
            </SheetTemplateGroupGrid>
          </SheetTemplateGroup>

        </SheetTemplateContent>
      )}

      {activeTab === 'activity' && (
        <SheetTemplateContent>
          <SheetTemplateGroup>
            <SheetTemplateGroupTab
              tabs={[
                {label: '일간', value: 'daily'},
                {label: '주간', value: 'weekly'},
                {label: '월간', value: 'monthly'},
              ]}
              activeTab={chartTab}
              onTabChange={setChartTab}
            >
              {/* 사용자는 TabsContent만 채우면 됩니다 */}
              <TabsContent value="daily">
                <div className="w-full h-full bg-blue-50/50 border border-blue-100 rounded-md flex items-end justify-around p-4">
                  {chartDataDaily.map((d) => (
                    <div key={d.name} className="w-8 bg-blue-400 rounded-t-sm flex flex-col justify-end items-center text-xs text-white pb-1 transition-all" style={{ height: `${(d.value / 800) * 100}%` }}>
                      <span className="opacity-0 hover:opacity-100">{d.value}</span>
                    </div>
                  ))}
                </div>
              </TabsContent>

              <TabsContent value="weekly" className="h-[250px] w-full">
                <div className="w-full h-full bg-green-50/50 border border-green-100 rounded-md flex items-end justify-around p-4">
                  {chartDataWeekly.map((d) => (
                    <div key={d.name} className="w-16 bg-green-400 rounded-t-sm flex flex-col justify-end items-center text-xs text-white pb-1 transition-all" style={{ height: `${(d.value / 9800) * 100}%` }}>
                      <span className="opacity-0 hover:opacity-100">{d.value}</span>
                    </div>
                  ))}
                </div>
              </TabsContent>

              <TabsContent value="monthly" className="h-[250px] w-full">
                <div className="w-full h-full bg-amber-50/50 border border-amber-100 rounded-md flex items-end justify-around p-4">
                  {chartDataMonthly.map((d) => (
                    <div key={d.name} className="w-16 bg-amber-400 rounded-t-sm flex flex-col justify-end items-center text-xs text-white pb-1 transition-all" style={{ height: `${(d.value / 4000) * 100}%` }}>
                      <span className="opacity-0 hover:opacity-100">{d.value}</span>
                    </div>
                  ))}
                </div>
              </TabsContent>
            </SheetTemplateGroupTab>
          </SheetTemplateGroup>
          <SheetTemplateGroup>
            <SheetTemplateGroupTable>
              <DataGrid
                enablePagination={false}
                tableHeight="auto"
                columns={activityColumns}
                data={activityData}
              />
            </SheetTemplateGroupTable>
          </SheetTemplateGroup>
        </SheetTemplateContent>
      )}

      <SheetTemplateFooter>
        <Button variant="outline" onClick={() => setClose()}>
          Cancel
        </Button>
        <Button variant="outline" onClick={() => console.log('Save clicked')}>
          Save
        </Button>
      </SheetTemplateFooter>
    </SheetTemplate>
  );
}

export function TableTabsDemo() {
  const [users] = useState<User[]>(generateUsers(15));
  const [products] = useState<Product[]>(generateProducts(25));
  const [infinityOrders, setInfinityOrders] = useState<Order[]>(generateOrders(8));
  const [allOrders] = useState<Order[]>(generateOrders(50));
  const [isLoading, setIsLoading] = useState(false);
  const [hasNextPage, setHasNextPage] = useState(true);

  const {setSheet} = useSheetStore();

  // Infinity scroll simulation
  const loadMoreOrders = async () => {
    if (isLoading) return;

    setIsLoading(true);

    // Simulate API call
    await new Promise((resolve) => setTimeout(resolve, 1500));

    const nextBatch = generateOrders(5);
    setInfinityOrders((prev) => [...prev, ...nextBatch]);

    // Simulate end of data after 3 loads
    if (infinityOrders.length > 20) {
      setHasNextPage(false);
    }

    setIsLoading(false);
  };

  const handlePageChange = (page: number) => {
    console.log(`Page changed to: ${page}`);
  };

  const handlePageSizeChange = (size: number) => {
    console.log(`Page size changed to: ${size}`);
  };

  // Tab-specific custom items
  const usersLeftItems = (
    <div className="flex items-center gap-2">
      <Select>
        <SelectTrigger className="w-35 h-8">
          <SelectValue placeholder="Filter Role" />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            <SelectLabel>User Roles</SelectLabel>
            <SelectItem value="admin">Admin</SelectItem>
            <SelectItem value="user">User</SelectItem>
            <SelectItem value="moderator">Moderator</SelectItem>
          </SelectGroup>
        </SelectContent>
      </Select>
    </div>
  );

  const usersRightItems = (
    <div className="flex items-center gap-2">
      <Button className="h-8 px-3 py-1 bg-blue-600 text-white rounded text-sm">Invite User</Button>
      <Button className="h-8 px-3 py-1 bg-green-600 text-white rounded text-sm">
        Export Users
      </Button>
    </div>
  );

  const productsLeftItems = (
    <div className="flex items-center gap-2">
      <Select>
        <SelectTrigger className="w-40 h-8">
          <SelectValue placeholder="Filter Category" />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            <SelectLabel>Categories</SelectLabel>
            <SelectItem value="electronics">Electronics</SelectItem>
            <SelectItem value="books">Books</SelectItem>
            <SelectItem value="clothing">Clothing</SelectItem>
            <SelectItem value="home">Home</SelectItem>
            <SelectItem value="sports">Sports</SelectItem>
          </SelectGroup>
        </SelectContent>
      </Select>
    </div>
  );

  const productsRightItems = (
    <div className="flex items-center gap-2">
      <Button className="h-8 px-3 py-1 bg-purple-600 text-white rounded text-sm">
        Add Product
      </Button>
      <Button className="h-8 px-3 py-1 bg-orange-600 text-white rounded text-sm">
        Bulk Import
      </Button>
    </div>
  );

  const ordersLeftItems = (
    <div className="flex items-center gap-2">
      <Select>
        <SelectTrigger className="w-35 h-8">
          <SelectValue placeholder="Filter Status" />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            <SelectLabel>Order Status</SelectLabel>
            <SelectItem value="pending">Pending</SelectItem>
            <SelectItem value="completed">Completed</SelectItem>
            <SelectItem value="cancelled">Cancelled</SelectItem>
          </SelectGroup>
        </SelectContent>
      </Select>
    </div>
  );

  const ordersRightItems = (
    <div className="flex items-center gap-2">
      <Button className="h-8 px-3 py-1 bg-red-600 text-white rounded text-sm">
        Process Orders
      </Button>
      <Button className="h-8 px-3 py-1 bg-indigo-600 text-white rounded text-sm">
        Export Orders
      </Button>
    </div>
  );

  const emptyLeftItems = (
    <div className="flex items-center gap-2">
      <Select>
        <SelectTrigger className="w-30 h-8">
          <SelectValue placeholder="No Filter" />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            <SelectLabel>No Options</SelectLabel>
            <SelectItem value="empty">Empty</SelectItem>
          </SelectGroup>
        </SelectContent>
      </Select>
    </div>
  );

  const emptyRightItems = (
    <div className="flex items-center gap-2">
      <Button className="h-8 px-3 py-1 bg-gray-600 text-white rounded text-sm">
        Create First Item
      </Button>
      <Button className="h-8 px-3 py-1 bg-blue-600 text-white rounded text-sm">Import Data</Button>
    </div>
  );

  return (
    <main className="flex flex-col w-full h-full">
      <TableTabsTemplate
        name="Admin Dashboard Demo"
        description="Perfect separation of concerns - each tab manages its own data, search, and custom items"
        defaultTab="Users"
        onTabChange={(tabName) => console.log(`Switched to tab: ${tabName}`)}
      >
        <TemplateTabs>
          <Tab name="Users">
            <div className={'p-4'}>
              <DataGrid
                data={users}
                columns={userColumns}
                tableKey="demo-users"
                leftFilters={() => [usersLeftItems]}
                rightFilters={() => [usersRightItems]}
                dataCellOnClick={(cell) => console.log('Users cell clicked:', cell.getValue())}
                onTableReady={(table) => console.log('Users table ready:', table)}
                onCurrentPageDataChange={(data) =>
                  console.log('Users current page:', data.length, 'rows')
                }
                emptyCustomMessage="No users found"
                searchableColumns={['name', 'email']}
                useTableSorting={true}
                rowCursor={true}
              />
            </div>
          </Tab>
          <Tab name="Products">
            <div className={'p-4'}>
              <TablePagination
                data={products}
                columns={productColumns}
                onPageChange={handlePageChange}
                onPageSizeChange={handlePageSizeChange}
                tableKey="demo-products"
                leftFilters={() => [productsLeftItems]}
                rightFilters={() => [productsRightItems]}
                dataCellOnClick={(cell) => console.log('Products cell clicked:', cell.getValue())}
                searchableColumns={['name', 'category']}
                useTableSorting={true}
                rowCursor={true}
              />
            </div>
          </Tab>
          <Tab name="Orders (Infinite)">
            <div className={'p-4'}>
              <DataGridInfinity
                data={infinityOrders}
                columns={orderColumns}
                hasNextPage={hasNextPage}
                isFetchingNextPage={isLoading}
                fetchNextPage={loadMoreOrders}
                tableKey="demo-infinity-orders"
                tableHeight={500}
                leftFilters={() => [ordersLeftItems]}
                rightFilters={() => [ordersRightItems]}
                dataCellOnClick={(cell) => console.log('Orders cell clicked:', cell.getValue())}
                searchableColumns={['productName', 'status']}
                useTableSorting={true}
                rowCursor={true}
              />
            </div>
          </Tab>
          <Tab name="All Orders">
            <div className={'p-4'}>
              <DataGrid
                data={allOrders}
                columns={orderColumns}
                tableKey="demo-all-orders"
                leftFilters={() => [ordersLeftItems]}
                rightFilters={() => [ordersRightItems]}
                dataCellOnClick={(cell) => console.log('All Orders cell clicked:', cell.getValue())}
                searchableColumns={['productName', 'status']}
                useTableSorting={true}
                rowCursor={true}
              />
            </div>
          </Tab>
          <Tab name="Empty Table">
            <div className={'p-4'}>
              <DataGrid
                data={[]}
                columns={userColumns}
                tableKey="demo-empty-table"
                leftFilters={() => [emptyLeftItems]}
                rightFilters={() => [emptyRightItems]}
                emptyCustomMessage="Sorry, nothing was found"
                searchableColumns={['name', 'email']}
                useTableSorting={true}
                rowCursor={false}
              />
            </div>
          </Tab>
          <Tab name="Disable Tab" disabled={true}>
            <div className={'p-4'}>
              <DataGrid
                data={users.slice(0, 8)}
                columns={userColumns}
                tableKey="demo-disable-tab"
                leftFilters={() => [usersLeftItems]}
                rightFilters={() => [usersRightItems]}
                emptyCustomMessage="This tab is disabled when you click the button above"
                searchableColumns={['name', 'email']}
                useTableSorting={true}
                rowCursor={true}
              />
            </div>
          </Tab>
          <Tab name="Sheet (default)">
            <div className={'p-4'}>
              <DataGrid
                data={users}
                columns={userColumns}
                tableKey="demo-sheet-type-a"
                leftFilters={() => [usersLeftItems]}
                rightFilters={() => [usersRightItems]}
                onRowClick={(row: User) => setSheet(<DemoUserSheetTypeA user={row} />)}
                searchableColumns={['name', 'email']}
                useTableSorting={true}
                rowCursor={true}
              />
            </div>
          </Tab>
          <Tab name="Sheet (tab)">
            <div className={'p-4'}>
              <DataGrid
                data={users}
                columns={userColumns}
                tableKey="demo-sheet-type-b"
                leftFilters={() => [usersLeftItems]}
                rightFilters={() => [usersRightItems]}
                onRowClick={(row: User) => setSheet(<DemoUserSheetTypeB user={row} />)}
                searchableColumns={['name', 'email']}
                useTableSorting={true}
                rowCursor={true}
              />
            </div>
          </Tab>

        </TemplateTabs>
      </TableTabsTemplate>
    </main>
  );
}
