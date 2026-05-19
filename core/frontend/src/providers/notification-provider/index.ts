import type {DataProvider} from '@/lib/data-provider';
import {axiosInstance} from '@lib/axios';
import {
  NOTIFICATION_RESOURCES,
  CreateNotificationRuleRequest,
  UpdateNotificationRuleRequest,
} from './types';

const API_URL = '/notification';

/**
 * Notification Provider for refine.dev
 *
 * Provides CRUD operations for notification rules
 *
 * Resources:
 * - rule: Notification rules (SNMP, Slack, etc.)
 */
export const notificationProvider: DataProvider = {
  /**
   * Get list of resources
   * Used for: /notification/rule
   */
  getList: async ({resource, meta}) => {
    const {headers} = meta ?? {};

    switch (resource) {
      case NOTIFICATION_RESOURCES.RULE: {
        // GET /notification/rule
        const response = await axiosInstance.get(`${API_URL}/rule`, {
          headers,
        });

        const data = response.data ?? [];

        // Backend returns array of: { id, name, notification_type, rule, timestamp, updated_at }
        // We need to parse the rule JSON string to object for each item
        const parsedData = data.map((item: any) => {
          if (item && typeof item.rule === 'string') {
            try {
              return {
                ...item,
                rule: JSON.parse(item.rule),
              };
            } catch (e) {
              console.error('Failed to parse notification rule:', e);
              return item;
            }
          }
          return item;
        });

        return {
          data: parsedData as any,
          total: parsedData.length,
        };
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  /**
   * Get single resource by ID
   * Used for: /notification/rule/:id
   */
  getOne: async ({resource, id, meta}) => {
    const {headers} = meta ?? {};

    switch (resource) {
      case NOTIFICATION_RESOURCES.RULE: {
        // GET /notification/rule/:id
        const response = await axiosInstance.get(`${API_URL}/rule/${id}`, {headers});

        // Backend returns: { id, name, notification_type, rule, timestamp, updated_at }
        // We need to parse the rule JSON string to object
        const data = response.data;
        if (data && typeof data.rule === 'string') {
          try {
            data.rule = JSON.parse(data.rule);
          } catch (e) {
            console.error('Failed to parse notification rule:', e);
          }
        }

        return {
          data: data,
        };
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  /**
   * Create new resource
   * Used for: POST /notification/rule
   */
  create: async ({resource, variables, meta}) => {
    const {headers} = meta ?? {};

    switch (resource) {
      case NOTIFICATION_RESOURCES.RULE: {
        // POST /notification/rule
        const payload = variables as CreateNotificationRuleRequest;

        // Backend expects: { notification_type: string, rule: string (JSON) }
        const backendPayload = {
          notification_type: payload.notification_type,
          rule: JSON.stringify(payload.rule),
        };

        const response = await axiosInstance.post(`${API_URL}/rule`, backendPayload, {headers});

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
   * Used for: PUT /notification/rule/:id
   */
  update: async ({resource, id, variables, meta}) => {
    const {headers} = meta ?? {};

    switch (resource) {
      case NOTIFICATION_RESOURCES.RULE: {
        // PUT /notification/rule/:id
        const payload = variables as UpdateNotificationRuleRequest;

        // Backend expects: { notification_type: string, rule: string (JSON) }
        const backendPayload = {
          notification_type: payload.notification_type,
          rule: JSON.stringify(payload.rule),
        };

        const response = await axiosInstance.put(`${API_URL}/rule/${id}`, backendPayload, {headers});

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
   * Used for: DELETE /notification/rule/:id
   */
  deleteOne: async ({resource, id, meta}) => {
    const {headers} = meta ?? {};

    switch (resource) {
      case NOTIFICATION_RESOURCES.RULE: {
        // DELETE /notification/rule/:id
        const response = await axiosInstance.delete(`${API_URL}/rule/${id}`, {headers});

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

// ⚠️ Deprecated: Hooks moved to @features/notification
// Please use: import { useCreateNotificationRule } from '@features/notification';
// These exports are kept for backward compatibility
export * from './hooks';
