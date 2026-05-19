import type { DataProvider } from '@/lib/data-provider';
import { axiosInstance } from '@lib/axios';
import qs from 'query-string';
import {
  ALERT_RESOURCES,
  // AlertRule,
  // AlertValue,
  AlertStatusRequest,
  // AlertHistoryRequest,
  // AlertHistoryResponse,
  AlertQueryRequest,
  // AlertQueryResponse,
  AlertStatusMaskRequest,
  CreateAlertRuleRequest,
  UpdateAlertRuleRequest,
} from './types';

const API_URL = '/alert';

/**
 * Alert Provider for refine.dev
 *
 * Provides CRUD operations for alert rules and status management
 *
 * Resources:
 * - rule: Alert rules (query, event-status, event-history)
 * - status: Current alert status instances
 * - hist: Historical alert records
 * - query: Test alert queries
 */
export const alertProvider: DataProvider = {
  /**
   * Get list of resources
   * Used for: /alert/rule, /alert/status, /alert/hist
   */
  getList: async ({ resource, filters, pagination, meta }) => {
    const { headers } = meta ?? {};

    switch (resource) {
      case ALERT_RESOURCES.RULE: {
        // GET /alert/rule?detail=true
        const params: any = {};
        if (meta?.detail) {
          params.detail = 'true';
        }

        const response = await axiosInstance.get(`${API_URL}/rule?${qs.stringify(params)}`, {
          headers,
        });

        const data = response.data ?? [];

        return {
          data: data as any,
          total: data.length,
        };
      }

      case ALERT_RESOURCES.STATUS: {
        // GET /alert/status?name=...&status=...&severity=...
        const params: AlertStatusRequest = {};

        if (filters) {
          filters.forEach((filter: any) => {
            if (filter.field === 'name') params.name = filter.value;
            if (filter.field === 'status') params.status = filter.value;
            if (filter.field === 'severity') params.severity = filter.value;
            if (filter.field === 'alertType') params.alertType = filter.value;
          });
        }

        if (pagination) {
          params.limit = pagination.pageSize;
          params.offset = ((pagination.currentPage ?? 1) - 1) * (pagination.pageSize ?? 10);
        }

        const response = await axiosInstance.get(`${API_URL}/status?${qs.stringify(params)}`, {
          headers,
        });

        const data = response.data ?? [];

        return {
          data: data as any,
          total: data.length,
        };
      }

      case ALERT_RESOURCES.HIST: {
        // GET /alert/hist?name=...&start-time=...&end-time=...&count=...
        const params: any = {};

        if (filters) {
          filters.forEach((filter: any) => {
            if (filter.field === 'name') params.name = filter.value;
            if (filter.field === 'startTime') params['start-time'] = filter.value;
            if (filter.field === 'endTime') params['end-time'] = filter.value;
            if (filter.field === 'count') params.count = filter.value;
          });
        }

        const response = await axiosInstance.get(`${API_URL}/hist?${qs.stringify(params)}`, {
          headers,
        });

        const data = response.data ?? [];

        return {
          data: data as any,
          total: data.length,
        };
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  /**
   * Get single resource by ID
   * Used for: /alert/rule/:id
   */
  getOne: async ({ resource, id, meta }) => {
    const { headers } = meta ?? {};

    switch (resource) {
      case ALERT_RESOURCES.RULE: {
        // GET /alert/rule/:id
        const response = await axiosInstance.get(`${API_URL}/rule/${id}`, { headers });

        return {
          data: response.data,
        };
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  /**
   * Create new resource
   * Used for: POST /alert/rule, POST /alert/query
   */
  create: async ({ resource, variables, meta }) => {
    const { headers } = meta ?? {};

    switch (resource) {
      case ALERT_RESOURCES.RULE: {
        // POST /alert/rule
        const payload = variables as CreateAlertRuleRequest;
        const response = await axiosInstance.post(`${API_URL}/rule`, payload, { headers });

        return {
          data: response.data,
        };
      }

      case ALERT_RESOURCES.QUERY: {
        // POST /alert/query - Test alert query
        const payload = variables as AlertQueryRequest;
        const response = await axiosInstance.post(`${API_URL}/query`, payload, { headers });

        return {
          data: response.data,
        };
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  /**
   * Update existing resource
   * Used for: PUT /alert/rule/:id, PUT /alert/status/:name/:id/mask
   */
  update: async ({ resource, id, variables, meta }) => {
    const { headers } = meta ?? {};

    switch (resource) {
      case ALERT_RESOURCES.RULE: {
        // PUT /alert/rule/:id
        const payload = variables as UpdateAlertRuleRequest;
        const response = await axiosInstance.put(`${API_URL}/rule/${id}`, payload, { headers });

        return {
          data: response.data,
        };
      }

      case ALERT_RESOURCES.STATUS_MASK: {
        // PUT /alert/status/:name/:id/mask
        // id should be in format "name/alertId"
        const [name, alertId] = String(id).split('/');
        const payload = variables as AlertStatusMaskRequest;
        const response = await axiosInstance.put(
          `${API_URL}/status/${name}/${alertId}/mask`,
          payload,
          { headers },
        );

        return {
          data: response.data,
        };
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  /**
   * Delete resource
   * Used for: DELETE /alert/rule/:id, DELETE /alert/status/:name/:id
   */
  deleteOne: async ({ resource, id, meta }) => {
    const { headers } = meta ?? {};

    switch (resource) {
      case ALERT_RESOURCES.RULE: {
        // DELETE /alert/rule/:id
        const response = await axiosInstance.delete(`${API_URL}/rule/${id}`, { headers });

        return {
          data: response.data,
        };
      }

      case ALERT_RESOURCES.STATUS: {
        // DELETE /alert/status/:name/:id
        // id should be in format "name/alertId"
        const [name, alertId] = String(id).split('/');
        const response = await axiosInstance.delete(`${API_URL}/status/${name}/${alertId}`, {
          headers,
        });

        return {
          data: response.data,
        };
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  /**
   * Get API URL
   */
  getApiUrl: () => API_URL,
};

// Export types and constants
export * from './types';

// ⚠️ Deprecated: Hooks moved to @features/alert
// Please use: import { useCreateAlertRule } from '@features/alert';
// These exports are kept for backward compatibility
export * from './hooks';
