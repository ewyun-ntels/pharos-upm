import {axiosInstance} from '@lib/axios';
import {getNumber, responseConvert} from '@lib/convert';
import {IndexSchema} from '@pharos/shared/types/datasource';
import {
  QUERY_PARAM_START_TIME,
  QUERY_PARAM_END_TIME,
  QUERY_PARAM_STEP,
} from '@lib/query-params';
import {
  BaseRecord,
  CreateParams,
  CreateResponse,
  DataProvider,
  DeleteOneParams,
  DeleteOneResponse,
  GetListParams,
  GetListResponse,
  GetOneParams,
  UpdateParams,
  UpdateResponse,
} from '@/lib/data-provider';
import {ChartMetricData, ChartQueryRequest} from '../../types';
import qs from 'query-string';
import {DashboardData, DashboardFolder} from '@pharos/shared/types/dashboard';

const API_URL = '/dashboard';

/**
 * args Map의 함수값을 실행하여 string | number Map으로 변환
 * - 이유: QueryPipeline이 시간 파라미터를 () => number 형태로 저장하므로,
 *   Object.fromEntries 전에 resolve하지 않으면 JSON 직렬화 시 함수가 누락됨
 */
function resolveArgs(
  args: Map<string, string | number | (() => number)> | undefined,
): Map<string, string | number> {
  if (!args) return new Map();
  return new Map(
    [...args.entries()].map(([k, v]) => [k, typeof v === 'function' ? v() : v]),
  );
}

export const DASHBOARD_PROVIDER_NAME = 'dashboardProvider' as const;

export const DASHBOARD_RESOURCES = {
  DASHBOARD: 'dashboard',
  TIMESERIES: 'timeseries',
  RAW: 'raw',
  FAVORITES: 'dashboard/favorites',
  HISTORY: 'dashboard/history',
  ANNOTATIONS: 'annotations',
} as const;

const convertToChartData = async (promQueries: ChartQueryRequest[], signal?: AbortSignal): Promise<ChartMetricData> => {
  const queryPromises = promQueries.map((promQuery) => {
    const { dashboardId, datasourceName, query, templateStyle, args, id, kind } = promQuery;
    const url = `${API_URL}/${dashboardId}/query`;
    const resolved = resolveArgs(args);
    const startTime = getNumber(resolved.get(QUERY_PARAM_START_TIME));
    const endTime = getNumber(resolved.get(QUERY_PARAM_END_TIME));
    const step = getNumber(resolved.get(QUERY_PARAM_STEP));

    const variables = {
      ...Object.fromEntries(resolved),
      [QUERY_PARAM_START_TIME]: startTime,
      [QUERY_PARAM_END_TIME]: endTime,
      [QUERY_PARAM_STEP]: step,
    };

    // Panel 모드: id+kind이 있고 query가 없으면 저장된 패널 실행 (viewer용)
    // Run 모드: query+datasourceName이 있으면 직접 실행 (non-viewer용)
    const body = id && kind && !query
      ? { panel: { kind, id }, variables }
      : { run: { datasourceName: datasourceName || '', query: query || '', templateStyle: templateStyle || 'pongo2' }, variables };

    return {
      queryName: promQuery.queryName,
      query: promQuery.query,
      label: promQuery.label,
      startTime,
      endTime,
      step,
      promise: axiosInstance.post(url, body, { signal }),
    };
  });

  const responses = await Promise.all(queryPromises.map((qp) => qp.promise));

  let outputResult: ChartMetricData = {
    chartMetric: [],
    chartType: 'timeseries',
    endTime: 0,
    startTime: 0,
    step: 0,
    uniqueKeys: [],
  };

  responses.forEach((resp, index) => {
    const query = queryPromises[index].query ?? '';
    const name = queryPromises[index].queryName ?? '';
    let responseData = resp.data;

    if (typeof responseData === 'string') {
      const trimmed = responseData.trim();
      if (!trimmed) {
        responseData = { data: [], meta: [], rows: 0 };
      } else if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
        // JSON 형태이지만 파싱 불가 (NaN/Inf 리터럴 등)
        throw new Error('Unexpected response format');
      } else {
        // ClickHouse 예외 텍스트 등 의미 있는 메시지
        throw new Error(trimmed);
      }
    } else if (responseData && responseData.data == null) {
      responseData = { ...responseData, data: [], meta: [] };
    }

    const result = IndexSchema.parse(responseData);
    const label = queryPromises[index].label ?? '';
    const startTime = queryPromises[index].startTime ?? 0;
    const endTime = queryPromises[index].endTime ?? 0;
    const step = queryPromises[index].step ?? 0;

    outputResult = responseConvert(outputResult, result, query, name, label, startTime, endTime, step);
  });

  return outputResult;
};

/**
 * Dashboard Provider for refine.dev
 *
 * Resources:
 * - dashboard: Dashboard CRUD operations (GET, POST, PUT, DELETE /dashboard)
 * - timeseries: Time-series chart data conversion (TimeSeries, Pie, BarGauge)
 * - raw: Raw data query execution (Table, Variables, Filters)
 *
 * All queries use "run" mode: POST /dashboard/:id/query with { run: { datasourceName, query } }
 * All data fetching now uses getOne for consistency (getList delegates to getOne)
 */
export const dashboardProvider: DataProvider = {
  /**
   * Unified data fetching (used by useShow, useList)
   * - TIMESERIES: Chart data processing with time-series conversion
   * - RAW: Direct query execution for tables and variables
   * - DASHBOARD: Dashboard metadata
   */
  getOne: async ({ resource, id, meta }: GetOneParams): Promise<any> => {
    const { headers, signal } = meta ?? {};

    switch (resource) {
      case DASHBOARD_RESOURCES.TIMESERIES: {
        // Time-series chart data processing with query execution
        if (meta?.variables?.value) {
          const promQueries = meta.variables.value as ChartQueryRequest[];
          const chartData = await convertToChartData(promQueries, signal);
          return { data: chartData };
        }
        throw new Error('Time-series chart data conversion requires meta.variables.value');
      }

      case DASHBOARD_RESOURCES.RAW: {
        // RAW 데이터 쿼리 (테이블, 필터 등)
        if (!meta?.variables?.value) {
          throw new Error(
            'RAW resource requires meta.variables.value (ChartQueryRequest[]). ' +
            'Received meta: ' + JSON.stringify(meta, null, 2)
          );
        }

        const queryObject = (meta.variables.value as ChartQueryRequest[])[0];
        
        if (!queryObject) {
          throw new Error(
            'RAW resource: meta.variables.value array is empty. ' +
            'Expected at least one ChartQueryRequest.'
          );
        }
        
        const { dashboardId, query, datasourceName, templateStyle, args, id, kind } = queryObject;
        let url = `${API_URL}/${dashboardId}/query`;

        // 🔍 Inspector 모드: URL 쿼리 파라미터 추가
        if (meta?.inspect === true) {
          url += '?inspect=true';
        }

        const resolved = resolveArgs(args);
        const startTime = getNumber(resolved.get(QUERY_PARAM_START_TIME));
        const endTime = getNumber(resolved.get(QUERY_PARAM_END_TIME));
        const step = getNumber(resolved.get(QUERY_PARAM_STEP));

        const variables = {
          ...Object.fromEntries(resolved),
          [QUERY_PARAM_START_TIME]: startTime,
          [QUERY_PARAM_END_TIME]: endTime,
          [QUERY_PARAM_STEP]: step,
        };

        let body: unknown;

        // 🔐 권한 기반 실행 모드 선택
        if (kind && id) {
          // Panel 모드: 저장된 패널 실행 (read 권한만 필요)
          body = { panel: { kind, id }, variables };
        } else {
          // Run 모드: query 직접 실행 (write 권한 필요)
          body = { run: { datasourceName: datasourceName || '', query: query || '', templateStyle: templateStyle || 'pongo2' }, variables };
        }

        const response = await axiosInstance.post(url, body, { headers, signal });
        
        // 🔍 Inspector 모드: 전체 응답 반환 (inspect, statistics 포함)
        if (meta?.inspect === true) {
          return { data: response.data };
        }
        
        return { data: response.data?.data ?? [] };
      }

      case DASHBOARD_RESOURCES.DASHBOARD: {
        // GET /dashboard/:id (single dashboard)
        if (!id) {
          throw new Error('Dashboard getOne requires an id. Use getList for listing dashboards.');
        }
        return await axiosInstance.get(`${API_URL}/${id}`, { headers, signal });
      }

      default:
        throw new Error(`Unsupported resource: ${resource}. Dashboard provider only handles 'dashboard', 'timeseries', and 'raw'.`);
    }
  },

  /**
   * Get list of resources
   * - DASHBOARD: Get all dashboards (GET /dashboard)
   * - TIMESERIES, RAW: Not supported (use getOne instead)
   */
  getList: async <TData extends BaseRecord = BaseRecord>({ resource, meta, pagination }: GetListParams): Promise<GetListResponse<TData>> => {
    const { headers } = meta ?? {};

    switch (resource) {
      case DASHBOARD_RESOURCES.DASHBOARD: {
        // GET /dashboard + GET /folders 병렬 fetch
        const { query = {} } = meta ?? {};
        const [dashboardsRes, foldersRes] = await Promise.all([
          axiosInstance.get<Record<string, Omit<DashboardData, 'id'>>>(
            `${API_URL}?${qs.stringify(query)}`,
            { headers },
          ),
          axiosInstance.get<DashboardFolder[]>('/folders', { headers }),
        ]);

        const dashboardMap = dashboardsRes.data ?? {};
        const folders: DashboardFolder[] = foldersRes.data ?? [];

        // { "id1": {...}, "id2": {...} } → [{ id: "id1", ... }, ...]
        const data = Object.entries(dashboardMap).map(([id, value]) => ({
          ...value,
          id,
        }));

        return {
          data: data as any,
          total: data.length,
          meta: {folders},
        };
      }

      case DASHBOARD_RESOURCES.FAVORITES: {
        // GET /dashboard/favorites → [{ id, displayName }]
        const response = await axiosInstance.get(`${API_URL}/favorites`, { headers });
        const data = response.data ?? [];
        return { data, total: data.length };
      }

      case DASHBOARD_RESOURCES.HISTORY: {
        const pageSize = pagination?.pageSize ?? 20;
        const currentPage = pagination?.currentPage ?? 1;
        const limit = pageSize;
        const offset = (currentPage - 1) * pageSize;
        const search = meta?.query?.search ?? '';
        const params = new URLSearchParams({limit: String(limit), offset: String(offset)});
        if (search) params.set('search', search);
        const response = await axiosInstance.get(
          `${API_URL}/history?${params.toString()}`,
          { headers },
        );
        const body = response.data ?? {};
        return { data: body.data ?? [], total: body.total ?? 0 };
      }

      case DASHBOARD_RESOURCES.TIMESERIES:
      case DASHBOARD_RESOURCES.RAW:
        throw new Error(`getList is not supported for ${resource}. Use getOne instead.`);

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  /**
   * Create new resource
   * Used for: POST /dashboard
   */
  create: async <TData extends BaseRecord = BaseRecord, TVariables = {}>({resource, variables}: CreateParams<TVariables>): Promise<CreateResponse<TData>> => {
    switch (resource) {
      case DASHBOARD_RESOURCES.DASHBOARD: {
        // POST /dashboard
        const response = await axiosInstance.post(API_URL, variables);
        if (response.status < 200 || response.status > 299) throw response;
        return { data: response.data };
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  /**
   * Update existing resource
   * Used for: PUT /dashboard/:id
   */
  update: async <TData extends BaseRecord = BaseRecord, TVariables = {}>({resource, id, variables}: UpdateParams<TVariables>): Promise<UpdateResponse<TData>> => {
    switch (resource) {
      case DASHBOARD_RESOURCES.DASHBOARD: {
        // PUT /dashboard/:id
        const response = await axiosInstance.put(`${API_URL}/${id}`, variables);
        if (response.status < 200 || response.status > 299) throw response;
        return { data: response.data };
      }

      case DASHBOARD_RESOURCES.FAVORITES: {
        // PATCH /dashboard/:id/favorite (toggles favorite)
        const response = await axiosInstance.patch(`${API_URL}/${id}/favorite`);
        if (response.status < 200 || response.status > 299) throw response;
        return { data: response.data };
      }

      case DASHBOARD_RESOURCES.ANNOTATIONS: {
        // PUT /dashboard/:id/annotations
        const response = await axiosInstance.put(`${API_URL}/${id}/annotations`, variables);
        if (response.status < 200 || response.status > 299) throw response;
        return { data: response.data };
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  /**
   * Delete resource
   * Used for: DELETE /dashboard/:id
   */
  deleteOne: async <TData extends BaseRecord = BaseRecord, TVariables = {}>({resource, id}: DeleteOneParams<TVariables>): Promise<DeleteOneResponse<TData>> => {
    switch (resource) {
      case DASHBOARD_RESOURCES.DASHBOARD: {
        // DELETE /dashboard/:id
        const response = await axiosInstance.delete(`${API_URL}/${id}`);
        if (response.status < 200 || response.status > 299) throw response;
        return { data: response.data };
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  getApiUrl: () => '/',
};
