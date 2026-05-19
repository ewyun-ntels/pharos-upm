/**
 * Demo Data Provider
 *
 * Mock data provider for demonstration purposes
 */

import type {
  BaseRecord,
  DataProvider,
  GetListParams,
  GetListResponse,
  GetOneParams,
  GetOneResponse,
  CreateParams,
  CreateResponse,
  UpdateParams,
  UpdateResponse,
  DeleteOneParams,
  DeleteOneResponse
} from '@pharos/shared/lib/data-provider/types';

// Mock data types
interface DemoUser extends BaseRecord {
  id: string;
  name: string;
  email: string;
  role: string;
  status: 'active' | 'inactive';
  createdAt: string;
  department: string;
  avatar?: string;
}

interface DemoProduct extends BaseRecord {
  id: string;
  name: string;
  category: string;
  price: number;
  stock: number;
  status: 'available' | 'out_of_stock' | 'discontinued';
  createdAt: string;
  description: string;
}

// Mock data
const mockUsers: DemoUser[] = Array.from({ length: 50 }, (_, i) => ({
  id: `user-${i + 1}`,
  name: `User ${i + 1}`,
  email: `user${i + 1}@demo.com`,
  role: ['Admin', 'Manager', 'Developer', 'Designer', 'Analyst'][i % 5],
  status: Math.random() > 0.3 ? 'active' : 'inactive',
  department: ['Engineering', 'Design', 'Marketing', 'Sales', 'Operations'][i % 5],
  createdAt: new Date(Date.now() - Math.random() * 365 * 24 * 60 * 60 * 1000).toISOString(),
  avatar: `https://api.dicebear.com/7.x/avataaars/svg?seed=user${i + 1}`,
}));

const mockProducts: DemoProduct[] = Array.from({ length: 30 }, (_, i) => ({
  id: `product-${i + 1}`,
  name: `Product ${i + 1}`,
  category: ['Electronics', 'Clothing', 'Books', 'Home & Garden', 'Sports'][i % 5],
  price: Math.floor(Math.random() * 1000) + 10,
  stock: Math.floor(Math.random() * 100),
  status: ['available', 'out_of_stock', 'discontinued'][Math.floor(Math.random() * 3)] as any,
  description: `Description for Product ${i + 1}`,
  createdAt: new Date(Date.now() - Math.random() * 200 * 24 * 60 * 60 * 1000).toISOString(),
}));

// Helper functions
const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms));

const filterData = <T extends BaseRecord>(data: T[], params: GetListParams): { data: T[]; total: number } => {
  let filteredData = [...data];

  // Apply filters
  if (params.filters) {
    params.filters.forEach(filter => {
      if (filter.operator === 'contains' && filter.value) {
        filteredData = filteredData.filter(item =>
          String(item[filter.field]).toLowerCase().includes(String(filter.value).toLowerCase())
        );
      }
      if (filter.operator === 'eq' && filter.value) {
        filteredData = filteredData.filter(item => item[filter.field] === filter.value);
      }
    });
  }

  // Apply sorting
  if (params.sorters && params.sorters.length > 0) {
    const sorter = params.sorters[0];
    filteredData.sort((a, b) => {
      const aVal = a[sorter.field];
      const bVal = b[sorter.field];
      if (aVal < bVal) return sorter.order === 'asc' ? -1 : 1;
      if (aVal > bVal) return sorter.order === 'asc' ? 1 : -1;
      return 0;
    });
  }

  const total = filteredData.length;

  // Apply pagination
  const start = ((params.pagination?.currentPage || 1) - 1) * (params.pagination?.pageSize || 10);
  const end = start + (params.pagination?.pageSize || 10);

  return {
    data: filteredData.slice(start, end),
    total
  };
};

export const DEMO_PROVIDER_NAME = 'demoProvider' as const;

export const DEMO_RESOURCES = {
  USERS: 'demo-users',
  PRODUCTS: 'demo-products',
} as const;

export const demoDataProvider: DataProvider = {
  getList: async <TData extends BaseRecord = BaseRecord>(
    params: GetListParams,
  ): Promise<GetListResponse<TData>> => {
    // Simulate network delay
    await delay(300 + Math.random() * 500);

    let result;

    switch (params.resource) {
      case DEMO_RESOURCES.USERS:
        result = filterData(mockUsers, params);
        break;
      case DEMO_RESOURCES.PRODUCTS:
        result = filterData(mockProducts, params);
        break;
      default:
        throw new Error(`Unknown resource: ${params.resource}`);
    }

    return {
      data: result.data as unknown as TData[],
      total: result.total,
    };
  },

  getOne: async <TData extends BaseRecord = BaseRecord>(
    params: GetOneParams,
  ): Promise<GetOneResponse<TData>> => {
    await delay(200);

    let item;

    switch (params.resource) {
      case DEMO_RESOURCES.USERS:
        item = mockUsers.find(u => u.id === params.id);
        break;
      case DEMO_RESOURCES.PRODUCTS:
        item = mockProducts.find(p => p.id === params.id);
        break;
      default:
        throw new Error(`Unknown resource: ${params.resource}`);
    }

    if (!item) {
      throw new Error(`Item with id ${params.id} not found`);
    }

    return { data: item as unknown as TData };
  },

  create: async <TData extends BaseRecord = BaseRecord, TVariables = {}>(
    params: CreateParams<TVariables>,
  ): Promise<CreateResponse<TData>> => {
    await delay(500);

    // For demo purposes, we'll create a mock item since we don't know TVariables structure
    let newItem: any;

    switch (params.resource) {
      case DEMO_RESOURCES.USERS:
        newItem = {
          id: `user-${Date.now()}`,
          name: `New User`,
          email: `newuser@demo.com`,
          role: 'Developer',
          status: 'active' as const,
          department: 'Engineering',
          createdAt: new Date().toISOString(),
          avatar: `https://api.dicebear.com/7.x/avataaars/svg?seed=user${Date.now()}`,
          ...params.variables,
        };
        mockUsers.unshift(newItem);
        break;
      case DEMO_RESOURCES.PRODUCTS:
        newItem = {
          id: `product-${Date.now()}`,
          name: `New Product`,
          category: 'Electronics',
          price: 99.99,
          stock: 50,
          status: 'available' as const,
          description: `New product description`,
          createdAt: new Date().toISOString(),
          ...params.variables,
        };
        mockProducts.unshift(newItem);
        break;
      default:
        throw new Error(`Unknown resource: ${params.resource}`);
    }

    return { data: newItem as TData };
  },

  update: async <TData extends BaseRecord = BaseRecord, TVariables = {}>(
    params: UpdateParams<TVariables>,
  ): Promise<UpdateResponse<TData>> => {
    await delay(400);

    let dataArray;
    switch (params.resource) {
      case DEMO_RESOURCES.USERS:
        dataArray = mockUsers;
        break;
      case DEMO_RESOURCES.PRODUCTS:
        dataArray = mockProducts;
        break;
      default:
        throw new Error(`Unknown resource: ${params.resource}`);
    }

    const index = dataArray.findIndex(item => item.id === params.id);
    if (index === -1) {
      throw new Error(`Item with id ${params.id} not found`);
    }

    const updatedItem = { ...dataArray[index], ...params.variables };
    dataArray[index] = updatedItem;

    return { data: updatedItem as unknown as TData };
  },

  deleteOne: async <TData extends BaseRecord = BaseRecord, TVariables = {}>(params: DeleteOneParams<TVariables>): Promise<DeleteOneResponse<TData>> => {
    await delay(300);

    let dataArray;
    switch (params.resource) {
      case DEMO_RESOURCES.USERS:
        dataArray = mockUsers;
        break;
      case DEMO_RESOURCES.PRODUCTS:
        dataArray = mockProducts;
        break;
      default:
        throw new Error(`Unknown resource: ${params.resource}`);
    }

    const index = dataArray.findIndex(item => item.id === params.id);
    if (index === -1) {
      throw new Error(`Item with id ${params.id} not found`);
    }

    dataArray.splice(index, 1);

    return { data: { id: params.id } as TData };
  },

  getApiUrl: () => '/demo-api',
};
