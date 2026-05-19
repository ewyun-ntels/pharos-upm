/**
 * Data Provider Demo
 *
 * Demonstrates DataGrid usage with custom data provider from extension
 */

import { useState } from 'react';
import { ColumnDef } from '@tanstack/react-table';
import {DataGrid, SearchInput, SelectBox, IconButton} from '@pharos/shared/components/ui-extension';
import {Download} from 'lucide-react';
import { Badge } from '@pharos/shared/components/ui';
import { Button } from '@pharos/shared/components/ui';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@pharos/shared/components/ui';
import { useList } from '@pharos/core/lib/data-provider';
import { DEMO_PROVIDER_NAME, DEMO_RESOURCES } from '../providers/demoDataProvider';
import { Users, Package } from '@pharos/shared/components';

interface DemoUser {
  id: string;
  name: string;
  email: string;
  role: string;
  status: 'active' | 'inactive';
  createdAt: string;
  department: string;
  avatar?: string;
}

interface DemoProduct {
  id: string;
  name: string;
  category: string;
  price: number;
  stock: number;
  status: 'available' | 'out_of_stock' | 'discontinued';
  createdAt: string;
  description: string;
}

const userColumns: ColumnDef<DemoUser>[] = [
  {
    accessorKey: 'avatar',
    header: 'Avatar',
    cell: ({ row }) => (
      <img
        src={row.original.avatar}
        alt={row.original.name}
        className="w-8 h-8 rounded-full"
      />
    ),
  },
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
    cell: ({ getValue }) => {
      const role = getValue<string>();
      const roleColors = {
        Admin: 'destructive',
        Manager: 'default',
        Developer: 'secondary',
        Designer: 'outline',
        Analyst: 'default',
      } as const;

      return (
        <Badge variant={roleColors[role as keyof typeof roleColors] || 'default'}>
          {role}
        </Badge>
      );
    },
  },
  {
    accessorKey: 'department',
    header: 'Department',
  },
  {
    accessorKey: 'status',
    header: 'Status',
    cell: ({ getValue }) => {
      const status = getValue<string>();
      return (
        <Badge variant={status === 'active' ? 'default' : 'secondary'}>
          {status}
        </Badge>
      );
    },
  },
  {
    accessorKey: 'createdAt',
    header: 'Created At',
    cell: ({ getValue }) => {
      const date = getValue<string>();
      return new Date(date).toLocaleDateString();
    },
  },
];

const productColumns: ColumnDef<DemoProduct>[] = [
  {
    accessorKey: 'name',
    header: 'Product Name',
  },
  {
    accessorKey: 'category',
    header: 'Category',
    cell: ({ getValue }) => {
      const category = getValue<string>();
      return (
        <Badge variant="outline">
          {category}
        </Badge>
      );
    },
  },
  {
    accessorKey: 'price',
    header: 'Price',
    cell: ({ getValue }) => {
      const price = getValue<number>();
      return `$${price.toFixed(2)}`;
    },
  },
  {
    accessorKey: 'stock',
    header: 'Stock',
    cell: ({ getValue }) => {
      const stock = getValue<number>();
      const isLowStock = stock < 20;
      return (
        <Badge variant={isLowStock ? 'destructive' : 'default'}>
          {stock} units
        </Badge>
      );
    },
  },
  {
    accessorKey: 'status',
    header: 'Status',
    cell: ({ getValue }) => {
      const status = getValue<string>();
      const statusColors = {
        available: 'default',
        out_of_stock: 'destructive',
        discontinued: 'secondary',
      } as const;

      return (
        <Badge variant={statusColors[status as keyof typeof statusColors]}>
          {status.replace('_', ' ')}
        </Badge>
      );
    },
  },
  {
    accessorKey: 'createdAt',
    header: 'Created At',
    cell: ({ getValue }) => {
      const date = getValue<string>();
      return new Date(date).toLocaleDateString();
    },
  },
];

export function DataProviderDemo() {
  const [activeTab, setActiveTab] = useState<'users' | 'products'>('users');

  // Use demo data provider for users
  const usersQuery = useList<DemoUser>({
    resource: DEMO_RESOURCES.USERS,
    dataProviderName: DEMO_PROVIDER_NAME,
    pagination: {
      currentPage: 1,
      pageSize: 20,
      mode: 'server' as const,
    },
  });

  // Use demo data provider for products
  const productsQuery = useList<DemoProduct>({
    resource: DEMO_RESOURCES.PRODUCTS,
    dataProviderName: DEMO_PROVIDER_NAME,
    pagination: {
      currentPage: 1,
      pageSize: 15,
      mode: 'server' as const,
    },
  });

  const users = usersQuery.query.data;
  const usersLoading = usersQuery.query.isLoading;
  const products = productsQuery.query.data;
  const productsLoading = productsQuery.query.isLoading;

  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Data Provider Demo</h1>
        <p className="text-muted-foreground mt-2">
          Demonstration of DataGrid with custom data provider from extension.
          This shows how extensions can provide their own data sources without modifying core code.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
        <Card className="cursor-pointer transition-all hover:shadow-md" onClick={() => setActiveTab('users')}>
          <CardHeader className="flex flex-row items-center space-y-0 pb-2">
            <div className="flex items-center space-x-2">
              <Users className="h-6 w-6 text-blue-600" />
              <CardTitle className="text-lg">Demo Users</CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            <CardDescription>
              50 mock users with avatars, roles, departments, and status
            </CardDescription>
            <div className="mt-2">
              <Badge variant={activeTab === 'users' ? 'default' : 'outline'}>
                {activeTab === 'users' ? 'Active' : 'Click to view'}
              </Badge>
            </div>
          </CardContent>
        </Card>

        <Card className="cursor-pointer transition-all hover:shadow-md" onClick={() => setActiveTab('products')}>
          <CardHeader className="flex flex-row items-center space-y-0 pb-2">
            <div className="flex items-center space-x-2">
              <Package className="h-6 w-6 text-green-600" />
              <CardTitle className="text-lg">Demo Products</CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            <CardDescription>
              30 mock products with categories, pricing, and inventory
            </CardDescription>
            <div className="mt-2">
              <Badge variant={activeTab === 'products' ? 'default' : 'outline'}>
                {activeTab === 'products' ? 'Active' : 'Click to view'}
              </Badge>
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="flex space-x-2 mb-4">
        <Button
          variant={activeTab === 'users' ? 'default' : 'outline'}
          onClick={() => setActiveTab('users')}
          className="flex items-center space-x-2"
        >
          <Users className="h-4 w-4" />
          <span>Users</span>
        </Button>
        <Button
          variant={activeTab === 'products' ? 'default' : 'outline'}
          onClick={() => setActiveTab('products')}
          className="flex items-center space-x-2"
        >
          <Package className="h-4 w-4" />
          <span>Products</span>
        </Button>
      </div>

      {activeTab === 'users' && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Users className="h-5 w-5" />
              <span>Demo Users Table</span>
            </CardTitle>
            <CardDescription>
              Users data from custom extension data provider
            </CardDescription>
          </CardHeader>
          <CardContent>
            {usersLoading ? (
              <div className="flex items-center justify-center h-64">Loading users...</div>
            ) : (
              <DataGrid<DemoUser>
                data={users?.data || []}
                columns={userColumns}
                tableHeight="500px"
                searchableColumns={['name', 'email', 'role', 'department']}
                enablePagination
                onRowClick={(user) => {
                  console.log('User clicked:', user);
                  alert(`Selected user: ${user.name} (${user.email})`);
                }}
                exportConfig={{
                  fileName: 'demo-users',
                  mapData: (user) => ({
                    Name: user.name,
                    Email: user.email,
                    Role: user.role,
                    Department: user.department,
                    Status: user.status,
                    'Created At': new Date(user.createdAt).toLocaleDateString(),
                  }),
                }}
              />
            )}
          </CardContent>
        </Card>
      )}

      {activeTab === 'products' && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Package className="h-5 w-5" />
              <span>Demo Products Table</span>
            </CardTitle>
            <CardDescription>
              Products data from custom extension data provider with advanced filtering
            </CardDescription>
          </CardHeader>
          <CardContent>
            {productsLoading ? (
              <div className="flex items-center justify-center h-64">Loading products...</div>
            ) : (
              <DataGrid<DemoProduct>
                data={products?.data || []}
                columns={productColumns}
                tableHeight="500px"
                searchableColumns={['name', 'category', 'description']}
                enablePagination
                onRowClick={(product) => {
                  console.log('Product clicked:', product);
                  alert(`Selected product: ${product.name} - $${product.price}`);
                }}
                leftFilters={(table) => [
                  <SearchInput key="search" table={table} placeholder="Search products..." size="hsmall" />,
                  <SelectBox
                    key="category"
                    label="Category"
                    size="hsmall"
                    options={['All', 'Electronics', 'Clothing', 'Books', 'Home & Garden', 'Sports'].map(o => ({label: o, value: o}))}
                    onChange={(val) => console.log('Category filter:', val)}
                  />,
                  <SelectBox
                    key="status"
                    label="Status"
                    size="hsmall"
                    options={['All', 'available', 'out_of_stock', 'discontinued'].map(o => ({label: o, value: o}))}
                    onChange={(val) => console.log('Status filter:', val)}
                  />,
                ]}
                rightFilters={() => [
                  <IconButton key="export" icon={<Download />} size="icon-xs" variant="outline" />
                ]}
                exportConfig={{
                  fileName: 'demo-products',
                  mapData: (product) => ({
                    'Product Name': product.name,
                    Category: product.category,
                    Price: `$${product.price.toFixed(2)}`,
                    Stock: `${product.stock} units`,
                    Status: product.status,
                    'Created At': new Date(product.createdAt).toLocaleDateString(),
                    Description: product.description,
                  }),
                }}
              />
            )}
          </CardContent>
        </Card>
      )}

      <Card className="mt-8">
        <CardHeader>
          <CardTitle>🔧 Technical Implementation</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4 text-sm">
            <div>
              <h4 className="font-semibold">Extension Data Provider Registration:</h4>
              <code className="block bg-muted p-2 mt-1 rounded text-xs">
                {`// In extension/demo/src/index.ts
registerExtension({
  dataProviders: {
    [DEMO_PROVIDER_NAME]: demoDataProvider,
  },
});`}
              </code>
            </div>
            <div>
              <h4 className="font-semibold">Automatic Core Integration:</h4>
              <code className="block bg-muted p-2 mt-1 rounded text-xs">
                {`// Core automatically merges extension providers
dataProvider={{
  default: defaultProvider,
  ...getExtensionDataProviders(), // Auto-injected
}}`}
              </code>
            </div>
            <div>
              <h4 className="font-semibold">Usage in Components:</h4>
              <code className="block bg-muted p-2 mt-1 rounded text-xs">
                {`const { data } = useList({
  resource: 'demo-users',
  dataProviderName: 'demoProvider',
});`}
              </code>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
