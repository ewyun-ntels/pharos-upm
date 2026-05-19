import type {DataProvider} from '@/lib/data-provider';
import {axiosInstance} from '@lib/axios';
import qs from 'query-string';
import {
  UserSchema,
  CreateUserRequestSchema,
  SetAttributesRequestSchema,
  ChangePasswordRequestSchema,
  ChangeMyPasswordRequestSchema,
  UpdatePasswordExpirationRequestSchema,
  BlockUserRequestSchema,
} from '@pharos/shared/types/user';

const API_URL = '/';

// simpleRestProvider 제거 - 모든 메서드를 직접 구현으로 교체

/** User Provider 이름 상수 - /user, /me API 사용 */
export const USER_PROVIDER_NAME = 'userProvider' as const;

export const USER_RESOURCES = {
  USER: 'user',
  ME: 'me',
  ME_PASSWORD: 'me-password',
  DASHBOARD_PERMISSION: 'dashboard/{id}/permission',
  USER_METADATA_CONFIG: 'user-metadata-config',
  USER_METADATA_CONFIG_HISTORY: 'user-metadata-config-history',
  USER_METADATA_CONFIG_ROLLBACK: 'user-metadata-config-rollback',
  AUTH_SESSIONS: 'auth/sessions',
  AUTH_HISTORY: 'auth/history',
} as const;

export const userProvider: DataProvider = {
  getList: async ({resource, meta, pagination}) => {
    const {headers} = meta ?? {};
    const query = meta?.query ?? meta?.queryObject?.query ?? {};

    // resource에 따른 URL 매핑
    let url: string;
    if (resource === 'user') {
      url = '/user';
    } else if (resource === USER_RESOURCES.USER_METADATA_CONFIG_HISTORY) {
      url = '/users/metadata/config/history';
    } else {
      url = `/${resource}`;
    }

    // auth/history: 서버사이드 페이지네이션
    if (resource === USER_RESOURCES.AUTH_HISTORY && pagination) {
      const pageSize = pagination.pageSize ?? 20;
      const currentPage = pagination.currentPage ?? 1;
      const params = {
        ...query,
        limit: pageSize,
        offset: (currentPage - 1) * pageSize,
      };
      const response = await axiosInstance.get(`${url}?${qs.stringify(params)}`, {headers});
      const body = response.data ?? {};
      const data = Array.isArray(body.data) ? body.data : [];
      return {
        data,
        total: body.total ?? data.length,
      };
    }

    try {
      const response = await axiosInstance.get(`${url}?${qs.stringify(query)}`, {
        headers,
      });

      const data = response.data ?? [];

      // user-metadata-config-history: { history: [...] } 구조
      if (resource === USER_RESOURCES.USER_METADATA_CONFIG_HISTORY) {
        const historyData = Array.isArray(data.history) ? data.history : [];
        return {
          data: historyData,
          total: historyData.length,
        };
      }

      // 응답 데이터가 배열인지 확인
      let processedData = [];

      if (Array.isArray(data)) {
        processedData = data;
      } else if (data && typeof data === 'object') {
        // 객체 안에 배열이 있는 경우
        const arrayValue = Object.values(data).find((v) => Array.isArray(v));
        if (arrayValue) {
          processedData = arrayValue;
        } else {
          console.warn('Expected array but received:', data);
          processedData = [];
        }
      }

      return {
        data: processedData,
        total: processedData.length,
        page: 1,
        pageSize: processedData.length,
        pageCount: 1,
      };
    } catch (error) {
      throw error;
    }
  },

  getOne: async ({resource, id, meta}) => {
    const {headers} = meta ?? {};
    let url: string;

    // resource에 따른 URL 매핑
    switch (resource) {
      case 'user':
        url = `/user/${id}`;
        break;
      case 'me':
        url = '/me';
        break;
      case 'user-metadata-config':
        url = '/users/metadata/config';
        break;
      default:
        url = `/${resource}/${id}`;
    }

    const response = await axiosInstance.get(url, {headers});

    // /me 엔드포인트의 경우 Zod로 검증
    if (resource === 'me') {
      const validationResult = UserSchema.safeParse(response.data);

      if (!validationResult.success) {
        console.error('User data validation failed:', validationResult.error);
        // Promise.reject를 사용하여 호출자에게 에러 전파
        return Promise.reject(new Error('Invalid user data format from server'));
      }

      return {
        data: validationResult.data,
      };
    }

    return {
      data: response.data,
    };
  },

  create: async ({resource, variables}) => {
    // Validate create user request
    if (resource === 'user') {
      const validationResult = CreateUserRequestSchema.safeParse(variables);

      if (!validationResult.success) {
        console.error('Create user request validation failed:', validationResult.error);
        return Promise.reject(new Error('Invalid user data format'));
      }
    }

    // user-metadata-config-rollback: POST /users/metadata/config/rollback
    if (resource === USER_RESOURCES.USER_METADATA_CONFIG_ROLLBACK) {
      const response = await axiosInstance.post('/users/metadata/config/rollback', variables);
      if (response.status < 200 || response.status > 299) throw response;
      return {data: response.data};
    }

    const response = await axiosInstance.post(`/${resource}`, variables);

    if (response.status < 200 || response.status > 299) throw response;

    return {
      data: response.data,
    };
  },

  update: async ({resource, id, variables}) => {
    let url: string;

    // Validate request based on resource type
    switch (resource) {
      case 'user-password':
        url = `/user/${id}/password`;
        const pwValidation = ChangePasswordRequestSchema.safeParse(variables);
        if (!pwValidation.success) {
          console.error('Change password request validation failed:', pwValidation.error);
          return Promise.reject(new Error('Invalid password change format'));
        }
        break;
      case 'user-block':
        url = `/user/${id}/block?block=true`;
        const blockValidation = BlockUserRequestSchema.safeParse(variables);
        if (!blockValidation.success) {
          console.error('Block user request validation failed:', blockValidation.error);
          return Promise.reject(new Error('Invalid block request format'));
        }
        break;
      case 'user-unblock':
        url = `/user/${id}/block`;
        // unblock uses same schema as block
        const unblockValidation = BlockUserRequestSchema.safeParse(variables);
        if (!unblockValidation.success) {
          console.error('Unblock user request validation failed:', unblockValidation.error);
          return Promise.reject(new Error('Invalid unblock request format'));
        }
        break;
      case 'user-password-expire':
        url = `/user/${id}/password-expire-date`;
        const expireValidation = UpdatePasswordExpirationRequestSchema.safeParse(variables);
        if (!expireValidation.success) {
          console.error(
            'Update password expiration request validation failed:',
            expireValidation.error,
          );
          return Promise.reject(new Error('Invalid password expiration format'));
        }
        break;
      case 'user-attributes':
        url = `/user/${id}/attributes`;
        const attrValidation = SetAttributesRequestSchema.safeParse(variables);
        if (!attrValidation.success) {
          console.error('Set attributes request validation failed:', attrValidation.error);
          return Promise.reject(new Error('Invalid attributes format'));
        }
        break;
      case 'me-password':
        url = '/me/password';
        const myPwValidation = ChangeMyPasswordRequestSchema.safeParse(variables);
        if (!myPwValidation.success) {
          console.error('Change my password request validation failed:', myPwValidation.error);
          return Promise.reject(new Error('Invalid password change format'));
        }
        break;
      case 'user-metadata-config':
        url = '/users/metadata/config';
        break;
      default:
        url = `/${resource}/${id}`;
    }

    const response = await axiosInstance.put(url, variables);

    if (response.status < 200 || response.status > 299) throw response;

    return {
      data: response.data,
    };
  },

  deleteOne: async ({resource, id}) => {
    let url: string;

    // resource와 id에 따른 URL 매핑
    if (resource === 'user') {
      url = `/user/${id}`;
    } else {
      url = `/${resource}/${id}`;
    }

    const response = await axiosInstance.delete(url);

    if (response.status < 200 || response.status > 299) throw response;

    return {
      data: response.data,
    };
  },

  getApiUrl: () => API_URL,
};
