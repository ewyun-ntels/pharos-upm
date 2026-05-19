import { DataProvider } from '@/lib/data-provider';
import { axiosInstance } from '@lib/axios';
import {
  RoleMetadataConfigResponseSchema,
  RoleMetadataConfigHistoryResponseSchema,
} from '@pharos/shared/types/role';

const API_URL = '/';

/** Role Provider 이름 상수 - /api/roles/metadata API 사용 */
export const ROLE_PROVIDER_NAME = 'roleProvider' as const;

export const ROLE_RESOURCES = {
  ALL_ROLES: 'all-roles',
  BASE_ROLES: 'base-roles',
  CONFIG: 'role-metadata-config',
  CONFIG_HISTORY: 'role-metadata-config-history',
  CONFIG_ROLLBACK: 'role-metadata-config-rollback',
} as const;

export const roleProvider: DataProvider = {
  getList: async ({ resource, meta }) => {
    if (resource === ROLE_RESOURCES.ALL_ROLES) {
      const params = meta?.includeHidden ? { includeHidden: 'true' } : undefined;
      const response = await axiosInstance.get('/api/roles', { params });
      const data = response.data ?? [];
      return {
        data: data as any,
        total: (data as any[]).length,
      };
    }
    if (resource === ROLE_RESOURCES.BASE_ROLES) {
      const response = await axiosInstance.get('/api/roles/base');
      const data = response.data ?? [];
      return {
        data: data as any,
        total: (data as any[]).length,
      };
    }
    throw new Error('getList is not supported for role provider');
  },

  getOne: async ({ resource }) => {
    let url: string;

    // resource에 따른 URL 매핑
    switch (resource) {
      case ROLE_RESOURCES.CONFIG:
        url = '/api/roles/metadata/config';
        break;
      case ROLE_RESOURCES.CONFIG_HISTORY:
        url = '/api/roles/metadata/config/history';
        break;
      default:
        url = `/${resource}`;
    }

    const response = await axiosInstance.get(url);

    // resource에 따른 validation
    if (resource === ROLE_RESOURCES.CONFIG) {
      const validationResult = RoleMetadataConfigResponseSchema.safeParse(response.data);
      if (!validationResult.success) {
        console.error('Role metadata config validation failed:', validationResult.error);
        return Promise.reject(new Error('Invalid role metadata config format from server'));
      }
      return {
        data: validationResult.data.config as any,
      };
    } else if (resource === ROLE_RESOURCES.CONFIG_HISTORY) {
      const validationResult = RoleMetadataConfigHistoryResponseSchema.safeParse(response.data);
      if (!validationResult.success) {
        console.error('Role metadata config history validation failed:', validationResult.error);
        return Promise.reject(new Error('Invalid role metadata config history format from server'));
      }
      return {
        data: validationResult.data.history as any,
      };
    }

    return {
      data: response.data as any,
    };
  },

  create: async ({ resource, variables }) => {
    if (resource === ROLE_RESOURCES.CONFIG) {
      const response = await axiosInstance.put('/api/roles/metadata/config', variables);
      return {
        data: response.data as any,
      };
    } else if (resource === ROLE_RESOURCES.CONFIG_ROLLBACK) {
      const response = await axiosInstance.post('/api/roles/metadata/config/rollback', variables);
      return {
        data: response.data as any,
      };
    }
    throw new Error(`create is not supported for resource: ${resource}`);
  },

  update: async ({ resource, variables }) => {
    if (resource === ROLE_RESOURCES.CONFIG) {
      const response = await axiosInstance.put('/api/roles/metadata/config', variables);
      return {
        data: response.data as any,
      };
    }
    throw new Error(`update is not supported for resource: ${resource}`);
  },

  deleteOne: async ({ resource }) => {
    if (resource === ROLE_RESOURCES.CONFIG) {
      const response = await axiosInstance.delete('/api/roles/metadata/config');
      return {
        data: response.data as any,
      };
    }
    throw new Error(`deleteOne is not supported for resource: ${resource}`);
  },

  getApiUrl: () => API_URL,
};
