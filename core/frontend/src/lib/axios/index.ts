import axios, {InternalAxiosRequestConfig} from 'axios';
import type {HttpError} from '@/lib/data-provider';

// OAuth2 client_id
const CLIENT_ID = '7a1e1782d4162e17dfc6f2bedd2c74ca';

/**
 * Axios 인스턴스
 *
 * - 브라우저: HttpOnly 쿠키 자동 전송 (withCredentials: true)
 * - 401 응답 시 토큰 갱신 후 재시도
 *
 * REST API 클라이언트는 Authorization 헤더를 직접 설정해서 사용.
 */
export const axiosInstance = axios.create({
  timeout: 1000 * 60 * 11,
  withCredentials: true,
  headers: {
    'X-Auth-Type': 'cookie',
  },
});

let refreshPromise: Promise<void> | null = null;

const isAuthEndpoint = (url?: string) => {
  if (!url) return false;
  return url.startsWith('/auth/');
};

// CSRF 토큰 헤더 추가
axiosInstance.interceptors.request.use(
  async (config: InternalAxiosRequestConfig) => {
    if (!isAuthEndpoint(config.url)) {
      config.headers['X-CSRF-Token'] = 'fetch';
    }
    return config;
  },
  (error) => Promise.reject(error),
);

// 401 처리 - 토큰 갱신 후 재시도
axiosInstance.interceptors.response.use(
  (response) => response,
  async (error) => {
    // 취소된 요청은 그대로 전파 (React Query가 cancellation으로 인식)
    if (axios.isCancel(error)) {
      return Promise.reject(error);
    }

    const originalRequest = error.config;
    const status = error.response?.status;

    // config가 없는 경우
    if (!originalRequest) {
      const customError: HttpError = {
        ...error,
        message: error?.message || 'Network error',
        statusCode: status,
      };
      return Promise.reject(customError);
    }

    // 인증 엔드포인트는 재시도 안 함
    if (isAuthEndpoint(originalRequest.url)) {
      const customError: HttpError = {
        ...error,
        message: error?.response?.data?.message || 'Auth request failed',
        statusCode: status,
      };
      return Promise.reject(customError);
    }

    // 401 처리
    if (status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      // signal이 이미 abort 상태면 재시도하지 않음 (페이지 이동 등으로 취소된 경우)
      if (originalRequest.signal?.aborted) {
        return Promise.reject(axios.isCancel(error) ? error : new axios.Cancel('Request cancelled'));
      }

      // refresh 실패 → 로그인 페이지
      try {
        // 이미 갱신 중이면 대기
        if (!refreshPromise) {
          refreshPromise = axiosInstance.post('/auth/token',
            new URLSearchParams({
              grant_type: 'refresh_token',
              client_id: CLIENT_ID,
            }),
            {
              headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
              },
            }
          ).then(() => {}).finally(() => {
            refreshPromise = null;
          });
        }

        await refreshPromise;
      } catch {
        // refresh_token 만료 또는 갱신 실패 → 로그인 페이지
        // logged_in 쿠키 삭제 (backend에서도 삭제하지만 즉시 반영을 위해)
        document.cookie = 'logged_in=; path=/; max-age=0';
        
        if (window.location.pathname !== '/ui/login') {
          window.location.href = '/ui/login';
        }
        const customError: HttpError = {
          ...error,
          message: 'Session expired',
          statusCode: 401,
        };
        return Promise.reject(customError);
      }

      // 원본 signal 유지: 재시도 중에도 취소 가능하도록
      return axiosInstance(originalRequest);
    }

    // 기타 에러
    const customError: HttpError = {
      ...error,
      message: error.response?.data?.message || error.message,
      statusCode: status,
    };
    return Promise.reject(customError);
  },
);
