/**
 * Query Parameter Constants
 * 
 * 백엔드 쿼리에 사용되는 공통 파라미터 상수 정의
 */

/** 시간 관련 쿼리 파라미터 (초 단위) */
export const QUERY_PARAM_START_TIME = '__start_time' as const;
export const QUERY_PARAM_END_TIME = '__end_time' as const;
export const QUERY_PARAM_STEP = '__step' as const;

/** 시간 관련 쿼리 파라미터 (밀리초 단위) */
export const QUERY_PARAM_START_TIME_MS = '__start_time_ms' as const;
export const QUERY_PARAM_END_TIME_MS = '__end_time_ms' as const;
export const QUERY_PARAM_STEP_MS = '__step_ms' as const;
export const QUERY_PARAM_INTERVAL = '__interval' as const;
export const QUERY_PARAM_INTERVAL_MS = '__interval_ms' as const;

/** 기타 시스템 쿼리 파라미터 */
export const QUERY_PARAM_REFRESH_COUNT = '__refreshCount' as const;

/** 템플릿 변수 형식 */
export const TEMPLATE_VAR_START_TIME = `{{${QUERY_PARAM_START_TIME}}}` as const;
export const TEMPLATE_VAR_END_TIME = `{{${QUERY_PARAM_END_TIME}}}` as const;
export const TEMPLATE_VAR_STEP = `{{${QUERY_PARAM_STEP}}}` as const;

/** 변수 참조 정보 (UI 표시용) */
export const QUERY_PARAM_DESCRIPTIONS = {
  [QUERY_PARAM_START_TIME]: {
    name: TEMPLATE_VAR_START_TIME,
    desc: '시작 시간 (Unix 초)',
    example: '1772686800',
    unit: 'seconds',
  },
  [QUERY_PARAM_END_TIME]: {
    name: TEMPLATE_VAR_END_TIME,
    desc: '종료 시간 (Unix 초)',
    example: '1772690400',
    unit: 'seconds',
  },
  [QUERY_PARAM_STEP]: {
    name: TEMPLATE_VAR_STEP,
    desc: '시간 간격 (초)',
    example: '300',
    unit: 'seconds',
  },
  [QUERY_PARAM_INTERVAL]: {
    name: '$__interval',
    desc: 'Grafana interval',
    example: '300s',
    unit: 'duration',
  },
  [QUERY_PARAM_INTERVAL_MS]: {
    name: '$__interval_ms',
    desc: 'Grafana interval (milliseconds)',
    example: '300000',
    unit: 'milliseconds',
  },
} as const;
