/**
 * CATV Provider for refine.dev
 *
 * Provides CRUD operations for CATV STB control results and records
 *
 * Resources:
 * - schedule-results: 스케줄 실행 결과 목록 (id 단위 집계)
 * - schedule-result-records: 개별 STB 제어 결과 레코드 (id 내 k8s job 단위)
 */

import type {DataProvider} from '@lib/data-provider';
import {axiosInstance} from '@/lib/axios';
import qs from 'query-string';
import {
  CATV_RESOURCES,
  ScheduleResultParams,
  ScheduleResultResponse,
  ScheduleResultDetailParams,
  ScheduleResultDetailResponse,
} from './types';

const API_URL = '/catv/control';

export const CATV_PROVIDER_NAME = 'catvProvider' as const;

export {CATV_RESOURCES} from './types';
export type {HierarchyItem, ScheduleResultRecord, Schedule} from './types';

/**
 * CATV Data Provider
 *
 * Provides access to CATV STB control APIs:
 * - GET /catv/control/schedules/results
 * - GET /catv/control/schedules/results/{resultId}
 * - GET /catv/control/topology/sos
 * - GET /catv/control/topology/sos/{soId}/l3s
 * - GET /catv/control/topology/sos/{soId}/l3s/{l3Id}/cells
 * - GET /catv/control/topology/sos/{soId}/l3s/{l3Id}/cells/{cellId}/settopboxes
 * - GET /catv/control/schedules
 * - POST /catv/control/schedules
 * - PUT /catv/control/schedules
 * - DELETE /catv/control/schedules/{name}
 */
export const catvProvider: DataProvider = {
  /**
   * Get list of resources
   * Used for: schedule-results, schedule-result-records (with resultId meta)
   */
  getList: async ({resource, filters, pagination, meta}) => {
    const {headers} = meta ?? {};

    switch (resource) {
      case CATV_RESOURCES.SCHEDULE_RESULTS: {
        const params: ScheduleResultParams = {
          limit: pagination?.pageSize || 100,
          offset: ((pagination?.currentPage ?? 1) - 1) * (pagination?.pageSize || 100),
        };

        // Apply filters
        if (filters) {
          filters.forEach((filter: any) => {
            if (filter.field === 'id' && filter.value) params.id = filter.value;
            if (filter.field === 'work_type' && filter.value) params.work_type = filter.value;
            if (filter.field === 'schedule_id' && filter.value) params.schedule_id = filter.value;
            if (filter.field === 'schedule_name' && filter.value)
              params.schedule_name = filter.value;
            if (filter.field === 'from' && filter.value) params.from = filter.value;
            if (filter.field === 'to' && filter.value) params.to = filter.value;
          });
        }

        const response = await axiosInstance.get<ScheduleResultResponse>(
          `${API_URL}/schedules/results?${qs.stringify(params)}`,
          {headers},
        );

        const data = response.data;

        return {
          data: (data.items || []) as any,
          total: data.total,
        };
      }

      case CATV_RESOURCES.SCHEDULE_RESULT_DETAILS: {
        const resultId = meta?.resultId;
        if (!resultId) {
          throw new Error('resultId is required in meta for schedule-result-details resource');
        }

        const params: ScheduleResultDetailParams = {
          offset: ((pagination?.currentPage ?? 1) - 1) * (pagination?.pageSize || 1000),
        };

        // Apply filters
        if (filters) {
          filters.forEach((filter: any) => {
            if (filter.field === 'k8s_job_name' && filter.value) params.k8s_job_name = filter.value;
            if (filter.field === 'cm_mac_addr' && filter.value) params.cm_mac_addr = filter.value;
            if (filter.field === 'stb_mac_addr' && filter.value) params.stb_mac_addr = filter.value;
            if (filter.field === 'progress_status' && filter.value)
              params.progress_status = filter.value;
            if (filter.field === 'result_code' && filter.value) params.result_code = filter.value;
            if (filter.field === 'search' && filter.value) params.search = filter.value;
          });
        }

        const response = await axiosInstance.get<ScheduleResultDetailResponse>(
          `${API_URL}/schedules/results/${encodeURIComponent(resultId)}?${qs.stringify(params)}`,
          {headers},
        );

        const data = response.data;

        return {
          data: (data.items || []) as any,
          total: data.total,
        };
      }

      case CATV_RESOURCES.SOS: {
        const response = await axiosInstance.get<{items: string[]; count: number}>(
          `${API_URL}/topology/sos`,
          {headers},
        );

        const data = response.data;

        return {
          data: (data.items || []).map((item) => ({id: item, name: item})) as any,
          total: data.count,
        };
      }

      case CATV_RESOURCES.L3S: {
        const soId = meta?.soId;
        if (!soId) {
          throw new Error('soId is required in meta for l3s resource');
        }

        const response = await axiosInstance.get<{items: string[]; count: number}>(
          `${API_URL}/topology/sos/${encodeURIComponent(soId)}/l3s`,
          {headers},
        );

        const data = response.data;

        return {
          data: (data.items || []).map((item) => ({id: item, name: item})) as any,
          total: data.count,
        };
      }

      case CATV_RESOURCES.CELLS: {
        const soId = meta?.soId;
        const l3Id = meta?.l3Id;
        if (!soId || !l3Id) {
          throw new Error('soId and l3Id are required in meta for cells resource');
        }

        const response = await axiosInstance.get<{items: string[]; count: number}>(
          `${API_URL}/topology/sos/${encodeURIComponent(soId)}/l3s/${encodeURIComponent(l3Id)}/cells`,
          {headers},
        );

        const data = response.data;

        return {
          data: (data.items || []).map((item) => ({id: item, name: item})) as any,
          total: data.count,
        };
      }

      case CATV_RESOURCES.SETTOPBOXES: {
        const soId = meta?.soId;
        const l3Id = meta?.l3Id;
        const cellId = meta?.cellId;
        if (!soId || !l3Id || !cellId) {
          throw new Error('soId, l3Id, and cellId are required in meta for settopboxes resource');
        }

        const response = await axiosInstance.get<{items: string[]; count: number}>(
          `${API_URL}/topology/sos/${encodeURIComponent(soId)}/l3s/${encodeURIComponent(l3Id)}/cells/${encodeURIComponent(cellId)}/settopboxes`,
          {headers},
        );

        const data = response.data;

        return {
          data: (data.items || []).map((item) => ({id: item, name: item})) as any,
          total: data.count,
        };
      }

      case CATV_RESOURCES.SCHEDULES: {
        const response = await axiosInstance.get(`${API_URL}/schedules`, {headers});

        const data = response.data;

        return {
          data: (Array.isArray(data) ? data : []) as any,
          total: Array.isArray(data) ? data.length : 0,
        };
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  /**
   * Get single resource by ID
   * Used for: 단일 결과 조회
   */
  getOne: async ({resource, id, meta}) => {
    const {headers} = meta ?? {};

    switch (resource) {
      case CATV_RESOURCES.SCHEDULE_RESULTS: {
        const response = await axiosInstance.get<ScheduleResultResponse>(
          `${API_URL}/schedules/results?${qs.stringify({id: id, limit: 1})}`,
          {headers},
        );

        const data = response.data;
        const result = data.items?.[0];

        if (!result) {
          throw new Error(`Result not found: ${id}`);
        }

        return {
          data: result as any,
        };
      }

      case CATV_RESOURCES.SCHEDULE_RESULT_DETAILS: {
        // 결과 상세 엔드포인트에서 조회 (resultId in meta 필요)
        const resultId = meta?.resultId;
        if (!resultId) {
          throw new Error('resultId is required in meta for control-records getOne');
        }

        const response = await axiosInstance.get<ScheduleResultDetailResponse>(
          `${API_URL}/schedules/results/${encodeURIComponent(resultId)}?${qs.stringify({limit: 1})}`,
          {headers},
        );

        const data = response.data;
        const record = data.items?.[0];

        if (!record) {
          throw new Error(`Record not found: ${id}`);
        }

        return {
          data: record as any,
        };
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  /**
   * Create new resource
   * Used for: submitting control requests and creating schedules
   */
  create: async ({resource, variables, meta}) => {
    const {headers} = meta ?? {};

    switch (resource) {
      case CATV_RESOURCES.SCHEDULES: {
        const requestBody = variables as any;

        const response = await axiosInstance.post(`${API_URL}/schedules`, requestBody, {
          headers,
        });

        return {
          data: {
            id: requestBody.name,
            name: requestBody.name,
          } as any,
        };
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  /**
   * Update resource
   * Used for: updating schedules
   */
  update: async ({resource, id, variables, meta}) => {
    const {headers} = meta ?? {};

    switch (resource) {
      case CATV_RESOURCES.SCHEDULES: {
        const requestBody = variables as any;

        const response = await axiosInstance.put(`${API_URL}/schedules`, requestBody, {
          headers,
        });

        return {
          data: {
            id: requestBody.name || id,
            name: requestBody.name || id,
          } as any,
        };
      }

      default:
        throw new Error(`Update operation is not supported for resource: ${resource}`);
    }
  },

  /**
   * Delete resource
   * Used for: deleting schedules
   */
  deleteOne: async ({resource, id, meta}) => {
    const {headers} = meta ?? {};

    switch (resource) {
      case CATV_RESOURCES.SCHEDULES: {
        const response = await axiosInstance.delete(
          `${API_URL}/schedules/${encodeURIComponent(id)}`,
          {headers},
        );

        return {
          data: {id} as any,
        };
      }

      default:
        throw new Error(`Delete operation is not supported for resource: ${resource}`);
    }
  },

  /**
   * Get API URL
   */
  getApiUrl: () => API_URL,
};
