// ============================================================
// @/lib/data-provider – Drop-in replacement for @refinedev/core
//
// Import from here instead of '@refinedev/core'.
// ============================================================

// Types
export type {
  BaseRecord,
  Identifier,
  HttpError,
  SortOrder,
  CrudSorting,
  CrudFilter,
  Pagination,
  GetListParams,
  GetListResponse,
  GetOneParams,
  GetOneResponse,
  CreateParams,
  CreateResponse,
  UpdateParams,
  UpdateResponse,
  DeleteOneParams,
  DeleteOneResponse,
  ResourceProvider,
  DataProvider,
} from '@pharos/shared/lib/data-provider/types';
export type { AuthProvider } from './registry';

// Registry
export { dataProviderRegistry, setAuthProvider, getAuthProvider } from './registry';

// Data hooks
export {
  useList,
  useOne,
  useShow,
  useCreate,
  useUpdate,
  useDelete,
  useInvalidate,
  useInfiniteList,
  useCustom,
  useForm,
} from './hooks';

export type {
  UseListProps,
  UseOneProps,
  UseShowProps,
  UseCreateProps,
  UseUpdateProps,
  UseDeleteProps,
  UseInfiniteListProps,
  UseCustomProps,
  UseFormProps,
} from './hooks';

// Auth hooks & components
export { useGetIdentity, useLogout, Authenticated } from './auth';

// i18n hook
export { useTranslation } from './i18n';
