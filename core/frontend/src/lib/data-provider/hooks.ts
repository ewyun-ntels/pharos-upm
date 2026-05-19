import {
  useQuery,
  useMutation,
  useQueryClient,
  useInfiniteQuery,
  type UseQueryOptions,
  type InfiniteData,
} from '@tanstack/react-query';
import { dataProviderRegistry } from './registry';
import { axiosInstance } from '../axios';
import type {
  BaseRecord,
  GetListParams,
  GetListResponse,
  GetOneParams,
  GetOneResponse,
  CreateResponse,
  UpdateResponse,
  DeleteOneResponse,
  HttpError,
  Identifier,
} from '@pharos/shared/lib/data-provider/types';

// ============================================================
// Helpers
// ============================================================

function resolveProviderName(
  callSite?: string,
  metaProvider?: string,
  hookLevel?: string,
): string | undefined {
  return callSite ?? metaProvider ?? hookLevel;
}

// ============================================================
// useList
// ============================================================

export interface UseListProps<TData extends BaseRecord = BaseRecord, TError = HttpError> extends GetListParams {
  dataProviderName?: string;
  queryOptions?: Omit<UseQueryOptions<GetListResponse<TData>, TError>, 'queryKey' | 'queryFn'> & {
    queryKey?: ReadonlyArray<unknown>;
  };
}

export function useList<TData extends BaseRecord = BaseRecord, TError = HttpError>(
  props: UseListProps<TData, TError>,
) {
  const { resource, dataProviderName, filters, pagination, sorters, meta, queryOptions } = props;
  const { queryKey: customQueryKey, ...restQueryOptions } = queryOptions ?? {};

  const defaultQueryKey = [
    dataProviderName ?? 'default',
    resource,
    'list',
    { filters, pagination, sorters, meta },
  ] as const;

  const query = useQuery<GetListResponse<TData>, TError>({
    queryKey: (customQueryKey ?? defaultQueryKey) as ReadonlyArray<unknown>,
    queryFn: () => {
      const provider = dataProviderRegistry.get(dataProviderName);
      return provider.getList<TData>({ resource, filters, pagination, sorters, meta });
    },
    ...restQueryOptions,
  });

  return { query };
}

// ============================================================
// useOne
// ============================================================

export interface UseOneProps<TData extends BaseRecord = BaseRecord, TError = HttpError> extends GetOneParams {
  dataProviderName?: string;
  queryOptions?: Omit<UseQueryOptions<GetOneResponse<TData>, TError>, 'queryKey' | 'queryFn'> & {
    queryKey?: ReadonlyArray<unknown>;
  };
}

export function useOne<TData extends BaseRecord = BaseRecord, TError = HttpError>(
  props: UseOneProps<TData, TError>,
) {
  const { resource, id, dataProviderName, meta, queryOptions } = props;
  const { queryKey: customQueryKey, ...restQueryOptions } = queryOptions ?? {};

  const defaultQueryKey = [dataProviderName ?? 'default', resource, 'one', id] as const;

  const query = useQuery<GetOneResponse<TData>, TError>({
    queryKey: (customQueryKey ?? defaultQueryKey) as ReadonlyArray<unknown>,
    queryFn: () => {
      const provider = dataProviderRegistry.get(dataProviderName);
      return provider.getOne<TData>({ resource, id, meta });
    },
    ...restQueryOptions,
  });

  return { query };
}

// ============================================================
// useShow  (route-params based useOne wrapper)
// ============================================================

export type UseShowProps<TData extends BaseRecord = BaseRecord, TError = HttpError> = Partial<GetOneParams> & {
  dataProviderName?: string;
  queryOptions?: UseOneProps<TData, TError>['queryOptions'];
};

export function useShow<TData extends BaseRecord = BaseRecord, TError = HttpError>(
  props: UseShowProps<TData, TError>,
) {
  const { resource = '', id = '', dataProviderName, meta, queryOptions } = props;
  return useOne<TData, TError>({ resource, id, dataProviderName, meta, queryOptions });
}

// ============================================================
// useCreate
// ============================================================

interface CreateMutationVariables<TVariables = {}> {
  resource?: string;
  values: TVariables;
  dataProviderName?: string;
  meta?: Record<string, any> & { dataProviderName?: string };
}

export interface UseCreateProps {
  resource?: string;
  dataProviderName?: string;
  meta?: Record<string, any>;
}

export function useCreate<
  TData extends BaseRecord = BaseRecord,
  TError = HttpError,
  TVariables = {},
>(props?: UseCreateProps) {
  const hookProviderName = props?.dataProviderName;
  const hookResource = props?.resource;
  const hookMeta = props?.meta;
  const queryClient = useQueryClient();

  const mutation = useMutation<
    CreateResponse<TData>,
    TError,
    CreateMutationVariables<TVariables>
  >({
    mutationFn: ({ resource, values, dataProviderName, meta }) => {
      const resolvedResource = resource ?? hookResource ?? '';
      const resolvedMeta = { ...hookMeta, ...meta };
      const providerName = resolveProviderName(dataProviderName, resolvedMeta?.dataProviderName, hookProviderName);
      const provider = dataProviderRegistry.get(providerName);
      return provider.create<TData, TVariables>({ resource: resolvedResource, variables: values, meta: resolvedMeta });
    },
    onSuccess: (_, variables) => {
      const { resource, dataProviderName, meta } = variables;
      const resolvedResource = resource ?? hookResource ?? '';
      const resolvedMeta = { ...hookMeta, ...meta };
      const providerName = resolveProviderName(dataProviderName, resolvedMeta?.dataProviderName, hookProviderName);
      queryClient.invalidateQueries({
        queryKey: [providerName ?? 'default', resolvedResource],
      });
    },
  });

  return {
    mutate: mutation.mutate,
    mutateAsync: mutation.mutateAsync,
    mutation,
    isPending: mutation.isPending,
    isLoading: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error,
    data: mutation.data,
  };
}

// ============================================================
// useUpdate
// ============================================================

interface UpdateMutationVariables<TVariables = {}> {
  resource?: string;
  id?: Identifier;
  values: TVariables;
  dataProviderName?: string;
  meta?: Record<string, any> & { dataProviderName?: string };
}

export interface UseUpdateProps {
  resource?: string;
  id?: Identifier;
  dataProviderName?: string;
  meta?: Record<string, any>;
}

export function useUpdate<
  TData extends BaseRecord = BaseRecord,
  TError = HttpError,
  TVariables = {},
>(props?: UseUpdateProps) {
  const hookProviderName = props?.dataProviderName;
  const hookResource = props?.resource;
  const hookId = props?.id;
  const hookMeta = props?.meta;
  const queryClient = useQueryClient();

  const mutation = useMutation<
    UpdateResponse<TData>,
    TError,
    UpdateMutationVariables<TVariables>
  >({
    mutationFn: ({ resource, id, values, dataProviderName, meta }) => {
      const resolvedResource = resource ?? hookResource ?? '';
      const resolvedId = id ?? hookId ?? '';
      const resolvedMeta = { ...hookMeta, ...meta };
      const providerName = resolveProviderName(dataProviderName, resolvedMeta?.dataProviderName, hookProviderName);
      const provider = dataProviderRegistry.get(providerName);
      return provider.update<TData, TVariables>({ resource: resolvedResource, id: resolvedId, variables: values, meta: resolvedMeta });
    },
    onSuccess: (_, variables) => {
      const { resource, id, dataProviderName, meta } = variables;
      const resolvedResource = resource ?? hookResource ?? '';
      const resolvedId = id ?? hookId ?? '';
      const resolvedMeta = { ...hookMeta, ...meta };
      const providerName = resolveProviderName(dataProviderName, resolvedMeta?.dataProviderName, hookProviderName);
      const key = providerName ?? 'default';
      queryClient.invalidateQueries({ queryKey: [key, resolvedResource, 'list'] });
      queryClient.invalidateQueries({ queryKey: [key, resolvedResource, 'one', resolvedId] });
    },
  });

  return {
    mutate: mutation.mutate,
    mutateAsync: mutation.mutateAsync,
    mutation,
    isPending: mutation.isPending,
    isLoading: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error,
    data: mutation.data,
  };
}

// ============================================================
// useDelete
// ============================================================

interface DeleteMutationVariables<TVariables = {}> {
  resource: string;
  id: Identifier;
  dataProviderName?: string;
  values?: TVariables;
  meta?: Record<string, any> & { dataProviderName?: string };
}

export interface UseDeleteProps {
  dataProviderName?: string;
}

export function useDelete<
  TData extends BaseRecord = BaseRecord,
  TError = HttpError,
  TVariables = {},
>(props?: UseDeleteProps) {
  const hookProviderName = props?.dataProviderName;
  const queryClient = useQueryClient();

  const mutation = useMutation<
    DeleteOneResponse<TData>,
    TError,
    DeleteMutationVariables<TVariables>
  >({
    mutationFn: ({ resource, id, dataProviderName, values, meta }) => {
      const providerName = resolveProviderName(dataProviderName, meta?.dataProviderName, hookProviderName);
      const provider = dataProviderRegistry.get(providerName);
      return provider.deleteOne<TData, TVariables>({ resource, id, variables: values, meta });
    },
    onSuccess: (_, variables) => {
      const { resource, dataProviderName, meta } = variables;
      const providerName = resolveProviderName(dataProviderName, meta?.dataProviderName, hookProviderName);
      queryClient.invalidateQueries({
        queryKey: [providerName ?? 'default', resource, 'list'],
      });
    },
  });

  return {
    mutate: mutation.mutate,
    mutateAsync: mutation.mutateAsync,
    mutation,
    isPending: mutation.isPending,
    isLoading: mutation.isPending,
    isSuccess: mutation.isSuccess,
    isError: mutation.isError,
    error: mutation.error,
    data: mutation.data,
  };
}

// ============================================================
// useInvalidate
// ============================================================

interface InvalidateParams {
  resource?: string;
  dataProviderName?: string;
  invalidates?: ('list' | 'one' | 'detail' | 'all')[];
  id?: Identifier;
  /** TanStack Query v5: 비활성(unmounted) 쿼리도 즉시 refetch할지 여부. 기본값 'active'. */
  refetchType?: 'active' | 'inactive' | 'all' | 'none';
}

export function useInvalidate() {
  const queryClient = useQueryClient();

  return async (params: InvalidateParams) => {
    const { resource, dataProviderName, invalidates = ['all'], id, refetchType } = params;
    const providerName = dataProviderName ?? 'default';

    for (const target of invalidates) {
      if (target === 'all') {
        await queryClient.invalidateQueries({
          queryKey: resource ? [providerName, resource] : [providerName],
          ...(refetchType && { refetchType }),
        });
      } else if (target === 'list') {
        await queryClient.invalidateQueries({
          queryKey: [providerName, resource, 'list'],
          ...(refetchType && { refetchType }),
        });
      } else if (target === 'one' || target === 'detail') {
        // 'detail' is an alias for 'one'
        const queryKey = id
          ? [providerName, resource, 'one', id]
          : [providerName, resource, 'one'];
        await queryClient.invalidateQueries({ queryKey, ...(refetchType && { refetchType }) });
      }
    }
  };
}

// ============================================================
// useInfiniteList
// ============================================================

export interface InfiniteListQueryOptions {
  enabled?: boolean;
  retry?: boolean | number;
  staleTime?: number;
  gcTime?: number;
  refetchOnMount?: boolean | 'always';
  refetchOnWindowFocus?: boolean | 'always';
}

export interface UseInfiniteListProps<TData extends BaseRecord = BaseRecord, TError = HttpError> extends GetListParams {
  dataProviderName?: string;
  queryOptions?: InfiniteListQueryOptions;
}

export function useInfiniteList<TData extends BaseRecord = BaseRecord, TError = HttpError>(
  props: UseInfiniteListProps<TData, TError>,
) {
  const { resource, dataProviderName, filters, pagination, sorters, meta, queryOptions } = props;
  const pageSize = pagination?.pageSize ?? 10;

  const query = useInfiniteQuery<GetListResponse<TData>, TError, InfiniteData<GetListResponse<TData>>>({
    queryKey: [dataProviderName ?? 'default', resource, 'infinite', { filters, sorters, meta, pageSize }],
    queryFn: ({ pageParam = 1 }) => {
      const provider = dataProviderRegistry.get(dataProviderName);
      return provider.getList<TData>({
        resource,
        filters,
        sorters,
        meta,
        pagination: { ...pagination, currentPage: pageParam as number, pageSize },
      });
    },
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) => {
      const loaded = allPages.flatMap((p) => p.data).length;
      return loaded < lastPage.total ? allPages.length + 1 : undefined;
    },
    ...queryOptions,
  });

  return { query };
}

// ============================================================
// useCustom
// ============================================================

export interface UseCustomProps<TData = unknown, TError = HttpError> {
  url: string;
  method?: 'get' | 'post' | 'put' | 'patch' | 'delete';
  config?: {
    headers?: Record<string, string>;
    query?: Record<string, string | number | boolean>;
    payload?: unknown;
  };
  dataProviderName?: string;
  queryOptions?: Omit<UseQueryOptions<{ data: TData }, TError>, 'queryKey' | 'queryFn'>;
}

export function useCustom<TData = unknown, TError = HttpError>(props: UseCustomProps<TData, TError>) {
  const { url, method = 'get', config, queryOptions } = props;

  const query = useQuery<{ data: TData }, TError>({
    queryKey: ['custom', url, method, config],
    queryFn: async () => {
      const response = await axiosInstance.request<TData>({
        
        url,
        method,
        params: config?.query,
        data: config?.payload,
        headers: config?.headers,
      });
      return { data: response.data };
    },
    enabled: method === 'get',
    ...queryOptions,
  });

  return { query };
}

// ============================================================
// useForm
// ============================================================

export interface UseFormProps {
  action: 'create' | 'edit';
  resource: string;
  dataProviderName?: string;
  redirect?: boolean | string;
  id?: Identifier;
  meta?: Record<string, any>;
}

export function useForm<TData extends BaseRecord = BaseRecord, TVariables = {}>(
  props: UseFormProps,
) {
  const { action, resource, dataProviderName, id, meta } = props;
  const queryClient = useQueryClient();

  const onFinish = async (variables: TVariables): Promise<TData | null> => {
    const provider = dataProviderRegistry.get(dataProviderName);

    let result: { data: TData };
    if (action === 'create') {
      result = await provider.create<TData, TVariables>({ resource, variables, meta });
    } else {
      if (!id) throw new Error('useForm with action="edit" requires an id');
      result = await provider.update<TData, TVariables>({ resource, id, variables, meta });
    }

    queryClient.invalidateQueries({
      queryKey: [dataProviderName ?? 'default', resource],
    });

    return result?.data ?? null;
  };

  return { onFinish };
}
