import {axiosInstance} from '@lib/axios';
import type {
  CreateParams,
  DataProvider,
  DeleteOneParams,
  GetListParams,
  GetOneParams,
  UpdateParams,
} from '@/lib/data-provider';
import qs from 'query-string';

const API_URL = '/ui-config';

export const UI_CONFIG_PROVIDER_NAME = 'uiConfigProvider' as const;

export const UI_CONFIG_RESOURCES = {
  CONFIG: 'config',
  HOME: 'home',
  IMAGES: 'images',
  BADGES: 'badges',
} as const;

/**
 * UI Config Provider for UI configuration
 *
 * Provides operations for UI configuration resources including:
 * - ui-config/config: UI configuration settings
 * - ui-config/images: UI image assets (future use)
 *
 * Resources:
 * - ui-config/config: GET/POST/PUT for UI configuration
 * - ui-config/images: GET for UI image assets
 */
export const uiConfigProvider: DataProvider = {
  /**
   * Get list of resources
   * Used for: /ui-config/config
   */
  getList: async ({
    resource,
    meta
  }: GetListParams) => {
    const { headers } = meta ?? {};

    switch (resource) {
      case UI_CONFIG_RESOURCES.CONFIG: {
        const query = meta?.query ?? {};
        const response = await axiosInstance.get(`${API_URL}/${resource}?${qs.stringify(query)}`, { headers });
        const data = response.data ?? {};

        return {
          data,
          total: 1,
          page: 1,
          pageSize: 1,
          pageCount: 1,
        };
      }

      case UI_CONFIG_RESOURCES.HOME: {
        const response = await axiosInstance.get(`${API_URL}/${resource}`, { headers });
        const data = response.data ?? null;
        return {
          data,
          total: 1,
          page: 1,
          pageSize: 1,
          pageCount: 1,
        };
      }

      case UI_CONFIG_RESOURCES.IMAGES: {
        // Future implementation for image listing
        throw new Error(`getList operation not supported for resource: ${resource}`);
      }

      default:
        throw new Error(`Unsupported resource: ${resource}. UI Config provider only handles ui-config/* resources.`);
    }
  },

  /**
   * Get single resource by ID
   * Used for: /ui-config/config, /ui-config/images/:id
   */
  getOne: async ({ resource, id, meta }: GetOneParams) => {
    const { headers } = meta ?? {};

    switch (resource) {
      case UI_CONFIG_RESOURCES.CONFIG: {
        if (!id) {
          const { query = {} } = meta ?? {};
          return await axiosInstance.get(`${API_URL}/${resource}?${qs.stringify(query)}`, {headers});
        } else {
          return await axiosInstance.get(`${API_URL}/${resource}/${id}`, {headers});
        }
      }

      case UI_CONFIG_RESOURCES.IMAGES: {
        if (!id) {
          throw new Error('Image ID is required for fetching image resource');
        }
        return await axiosInstance.get(`${API_URL}/${resource}/${id}`, {headers});
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  /**
   * Create new resource
   * Used for: POST /ui-config/config
   */
  create: async <TVariables = {}>(
    {resource, variables}: CreateParams<TVariables>
  ) => {
    switch (resource) {
      case UI_CONFIG_RESOURCES.CONFIG: {
        const response = await axiosInstance.post(`${API_URL}/${resource}`, variables);
        if (response.status < 200 || response.status > 299) throw response;
        return { data: response.data };
      }

      case UI_CONFIG_RESOURCES.IMAGES: {
        // Future implementation for image upload
        throw new Error(`create operation not supported for resource: ${resource}`);
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  /**
   * Update existing resource
   * Used for: PUT /ui-config/config/:id
   */
  update: async <TVariables = {}>(
    {resource, id, variables}: UpdateParams<TVariables>
  ) => {
    switch (resource) {
      case UI_CONFIG_RESOURCES.CONFIG: {
        const response = await axiosInstance.put(`${API_URL}/${resource}/${id}`, variables);
        if (response.status < 200 || response.status > 299) throw response;
        return { data: response.data };
      }

      case UI_CONFIG_RESOURCES.HOME: {
        const response = await axiosInstance.put(`${API_URL}/${resource}`, variables);
        if (response.status < 200 || response.status > 299) throw response;
        return { data: response.data };
      }

      case UI_CONFIG_RESOURCES.IMAGES: {
        // Future implementation for image update
        throw new Error(`update operation not supported for resource: ${resource}`);
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  /**
   * Delete resource
   * Used for: DELETE /ui-config/config/:id
   */
  deleteOne: async <TVariables = {}>(
    {resource, id}: DeleteOneParams<TVariables>
  ) => {
    switch (resource) {
      case UI_CONFIG_RESOURCES.CONFIG: {
        const response = await axiosInstance.delete(`${API_URL}/${resource}/${id}`);
        if (response.status < 200 || response.status > 299) throw response;
        return { data: response.data };
      }

      case UI_CONFIG_RESOURCES.HOME: {
        const response = await axiosInstance.delete(`${API_URL}/${resource}`);
        if (response.status < 200 || response.status > 299) throw response;
        return { data: response.data };
      }

      case UI_CONFIG_RESOURCES.IMAGES: {
        // Future implementation for image deletion
        throw new Error(`deleteOne operation not supported for resource: ${resource}`);
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  getApiUrl: () => API_URL,
};
