/**
 * Dashboard URL Parameter Constants
 *
 * URL ↔ Store 동기화에 사용되는 모든 파라미터명을 한 곳에서 관리한다.
 * 이전에 'step' vs 'var_step' 등 파라미터명 불일치 문제가 있었으므로,
 * 반드시 이 상수를 통해서만 파라미터명에 접근해야 한다.
 */

/** DateTime 관련 URL 파라미터 */
export const URL_PARAM_FROM = 'from' as const;
export const URL_PARAM_TO = 'to' as const;

/** Step URL 파라미터 (var_step으로 통일) */
export const URL_PARAM_STEP = 'var_step' as const;

/** Refresh interval URL 파라미터 */
export const URL_PARAM_REFRESH = 'refresh' as const;

/** Select 필터 URL 파라미터 prefix */
export const URL_PARAM_FILTER_PREFIX = 'var-' as const;

/** Select 필터 ID로 URL 파라미터명 생성 */
export const getFilterParamName = (filterId: string) =>
  `${URL_PARAM_FILTER_PREFIX}${filterId}` as const;
