// ============================================================
// Shared Data Provider Types
// Extensions import from '@pharos/shared/lib/data-provider/types'
// ============================================================

// BaseRecord: Record<string, any>를 유지해야 함.
// Record<string, unknown>으로 바꾸면 인덱스 시그니처 없는 타입(RoleGroup, DemoUser 등)이
// TData extends BaseRecord 제약을 만족하지 못하는 TypeScript 한계가 있음.
export type BaseRecord = Record<string, any> & { id?: string | number };

export type Identifier = string | number;

// [key: string]: any를 유지해야 함.
// unknown으로 바꾸면 error.response?.data?.message 등 axios 에러 구조 접근 코드에서 타입에러 발생.
export interface HttpError {
  message: string;
  statusCode: number;
  [key: string]: any;
}

export type SortOrder = 'asc' | 'desc';

export interface CrudSorting {
  field: string;
  order: SortOrder;
}

// value: any를 유지해야 함.
// unknown으로 바꾸면 provider 내 filter.value 직접 사용 코드가 전부 타입에러 발생.
export type CrudFilter = {
  field: string;
  operator: string;
  value: any;
};

export interface Pagination {
  currentPage?: number;
  pageSize?: number;
  mode?: 'client' | 'server' | 'off';
}

export interface GetListParams {
  resource: string;
  pagination?: Pagination;
  filters?: CrudFilter[];
  sorters?: CrudSorting[];
  meta?: Record<string, any>;
}

export interface GetListResponse<TData extends BaseRecord = BaseRecord> {
  data: TData[];
  total: number;
  [key: string]: unknown;
}

export interface GetOneParams {
  resource: string;
  id: Identifier;
  meta?: Record<string, any>;
}

export interface GetOneResponse<TData extends BaseRecord = BaseRecord> {
  data: TData;
}

export interface CreateParams<TVariables = {}> {
  resource: string;
  variables: TVariables;
  meta?: Record<string, any>;
}

export interface CreateResponse<TData extends BaseRecord = BaseRecord> {
  data: TData;
}

export interface UpdateParams<TVariables = {}> {
  resource: string;
  id: Identifier;
  variables: TVariables;
  meta?: Record<string, any>;
}

export interface UpdateResponse<TData extends BaseRecord = BaseRecord> {
  data: TData;
}

export interface DeleteOneParams<TVariables = {}> {
  resource: string;
  id: Identifier;
  variables?: TVariables;
  meta?: Record<string, any>;
}

export interface DeleteOneResponse<TData extends BaseRecord = BaseRecord> {
  data: TData;
}

/**
 * ResourceProvider interface – replaces @refinedev/core DataProvider.
 * All extension data providers should implement this interface.
 */
export interface ResourceProvider {
  getList: <TData extends BaseRecord = BaseRecord>(
    params: GetListParams,
  ) => Promise<GetListResponse<TData>>;
  getOne: <TData extends BaseRecord = BaseRecord>(
    params: GetOneParams,
  ) => Promise<GetOneResponse<TData>>;
  create: <TData extends BaseRecord = BaseRecord, TVariables = {}>(
    params: CreateParams<TVariables>,
  ) => Promise<CreateResponse<TData>>;
  update: <TData extends BaseRecord = BaseRecord, TVariables = {}>(
    params: UpdateParams<TVariables>,
  ) => Promise<UpdateResponse<TData>>;
  deleteOne: <TData extends BaseRecord = BaseRecord, TVariables = {}>(
    params: DeleteOneParams<TVariables>,
  ) => Promise<DeleteOneResponse<TData>>;
  getApiUrl?: () => string;
}

// Alias for backward compatibility
export type DataProvider = ResourceProvider;
