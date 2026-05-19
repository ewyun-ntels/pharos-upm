import type {
  CreateParams,
  DataProvider,
  DeleteOneParams,
  GetListParams,
  GetOneParams,
  UpdateParams,
} from '@/lib/data-provider';
import {axiosInstance} from '@lib/axios';

const API_URL = '/plugins';

export const PLUGIN_PROVIDER_NAME = 'pluginProvider' as const;

export const PLUGIN_RESOURCES = {
  DATASOURCES: 'datasources',
  // 향후 다른 플러그인 리소스 추가 가능
  // EXTENSIONS: 'extensions',
  // CONFIGS: 'configs',
} as const;

/**
 * Plugin Provider for plugin and extension management
 *
 * Provides operations for plugin resources including:
 * - datasources: Available datasource plugins
 * - Future: extensions, configs, etc.
 *
 * Resources:
 * - datasources: GET /plugins/datasources for datasource list
 */
export const pluginProvider: DataProvider = {
  /**
   * Get list of resources
   * Used for: /plugins/datasources
   */
  getList: async ({
    resource,
    meta
  }: GetListParams) => {
    const { headers } = meta ?? {};

    switch (resource) {
      case PLUGIN_RESOURCES.DATASOURCES: {
        try {
          const response = await axiosInstance.get(`${API_URL}/${resource}`, { headers });
          const data = Array.isArray(response.data) ? response.data : [];

          return {
            data,
            total: data.length,
            page: 1,
            pageSize: data.length,
            pageCount: 1,
          };
        } catch (error) {
          console.error(`Failed to fetch ${resource}:`, error);
          throw error;
        }
      }

      default:
        throw new Error(`Unsupported resource: ${resource}. Plugin provider only handles plugin/* resources.`);
    }
  },

  /**
   * Get single resource by ID
   * Used for: /plugins/{resource}/{id}
   */
  getOne: async ({ resource, id, meta }: GetOneParams) => {
    const { headers } = meta ?? {};

    switch (resource) {
      case PLUGIN_RESOURCES.DATASOURCES: {
        if (!id) {
          throw new Error('Datasource ID is required for fetching single datasource');
        }
        try {
          const response = await axiosInstance.get(`${API_URL}/${resource}/${id}`, { headers });
          return { data: response.data };
        } catch (error) {
          console.error(`Failed to fetch ${resource}/${id}:`, error);
          throw error;
        }
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  /**
   * Create new resource
   * Used for: POST /plugins/{resource}
   */
  create: async <TVariables = {}>(
    {resource}: CreateParams<TVariables>
  ) => {
    switch (resource) {
      case PLUGIN_RESOURCES.DATASOURCES: {
        // Future implementation for datasource creation
        throw new Error(`create operation not yet implemented for resource: ${resource}`);
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  /**
   * Update existing resource
   * Used for: PUT /plugins/{resource}/{id}
   */
  update: async <TVariables = {}>(
    {resource}: UpdateParams<TVariables>
  ) => {
    switch (resource) {
      case PLUGIN_RESOURCES.DATASOURCES: {
        // Future implementation for datasource update
        throw new Error(`update operation not yet implemented for resource: ${resource}`);
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  /**
   * Delete resource
   * Used for: DELETE /plugins/{resource}/{id}
   */
  deleteOne: async <TVariables = {}>(
    {resource}: DeleteOneParams<TVariables>
  ) => {
    switch (resource) {
      case PLUGIN_RESOURCES.DATASOURCES: {
        // Future implementation for datasource deletion
        throw new Error(`deleteOne operation not yet implemented for resource: ${resource}`);
      }

      default:
        throw new Error(`Unsupported resource: ${resource}`);
    }
  },

  getApiUrl: () => API_URL,
};
