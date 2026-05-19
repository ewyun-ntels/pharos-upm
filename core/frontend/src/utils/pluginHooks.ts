/**
 * Plugin Context - 플러그인 타입 정의
 * 
 * 외부 플러그인이 core의 hooks, 컴포넌트, 유틸리티를 사용할 수 있도록 타입을 정의합니다.
 */

import { z } from 'zod';
import { useDashboardQuery, toChartQueryRequests } from '@features/dashboard/hooks/useDashboardQuery';
import {
  useList,
  useOne,
  useCreate,
  useUpdate,
  useDelete,
  useCustom,
  DataProvider,
  useTranslation,
} from '@/lib/data-provider';


// 플러그인 Context 타입
export interface PluginContext {
  // Hooks
  hooks: {
    useDashboardQuery: typeof useDashboardQuery;
    toChartQueryRequests: typeof toChartQueryRequests;
    // Refine Data Hooks - CRUD 작업을 위한 hooks
    useList: typeof useList;
    useOne: typeof useOne;
    useCreate: typeof useCreate;
    useUpdate: typeof useUpdate;
    useDelete: typeof useDelete;
    useCustom: typeof useCustom;
    // 다국어 지원
    useTranslation: typeof useTranslation;
  };

  // Data Providers - 특정 리소스에 대한 접근
  providers?: {
    // alert, dashboard, datasource 등의 provider에 접근 가능
    getProvider?: (name: string) => DataProvider | undefined;
  };

  // 유틸리티 함수들 - 필요시 추가
  utils?: {
    formatLegendLabel?: (value: any, label: string) => string;
    calculateMetrics?: (data: any[]) => Record<string, number>;
    formatDate?: (date: Date | string | number, format?: string) => string;
    // 더 많은 유틸리티 함수 추가 가능
  };

  // UI Components (선택적) - 기본 HTML 컴포넌트 사용 권장
  components?: {
    // 기본 HTML 엘리먼트를 사용하여 외부 의존성 최소화
    [componentName: string]: React.ComponentType<any>;
  };
}

/**
 * PluginContext Zod Schema
 * 
 * Zod로 pluginContext 구조를 검증합니다.
 * z.custom()을 사용하여 TypeScript 타입을 유지합니다.
 */
export const PluginContextSchema = z.custom<PluginContext>((data) => {
  return (
    data &&
    typeof data === 'object' &&
    'hooks' in data &&
    typeof (data as any).hooks === 'object' &&
    typeof (data as any).hooks.useDashboardQuery === 'function'
  );
}).optional();

/**
 * Plugin Context를 props로 가져오기
 *
 * 외부 플러그인이 core의 hooks와 컴포넌트를 사용할 수 있도록
 * 안전하고 제한된 컨텍스트를 제공합니다.
 */
export function getPluginContext(): PluginContext {
  return {
    hooks: {
      useDashboardQuery,
      toChartQueryRequests,
      useList,
      useOne,
      useCreate,
      useUpdate,
      useDelete,
      useCustom,
      useTranslation,
    },
    // utils는 필요시 추가 가능
    utils: {
      // 기본 유틸리티 함수들을 여기에 추가
    },
    // providers는 필요시 추가
    providers: {
      getProvider: (_name: string) => {
        // DataProvider 접근 로직을 여기에 구현
        return undefined;
      },
    },
  };
}
