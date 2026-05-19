/**
 * Alert Feature Hooks
 * 
 * Alert 기능 관련 모든 hooks를 export합니다.
 * 
 * ## Mutation Hooks (CUD operations)
 * - useCreateAlertRule: Alert Rule 생성
 * - useUpdateAlertRule: Alert Rule 수정
 * - useUpdateAlertStatus: Alert Status 수정
 * - useMaskAlertStatus: Alert 마스킹/언마스킹
 * 
 * ## Query Hooks (Read operations)
 * - useAlertRuleList: Alert Rule 목록 조회
 * - useAlertStatusList: Alert Status 목록 조회
 * - useAlertHistoryList: Alert History 조회 (페이징 미지원)
 * - useAlertHistoryInfinite: Alert History Infinite Scroll (백엔드 준비 필요)
 */

// Mutation hooks
export { useCreateAlertRule, useUpdateAlertRule } from './use-alert-rule-mutation';
export { useUpdateAlertStatus, useMaskAlertStatus } from './use-alert-status-mutation';

// Query hooks
export { 
  useAlertRuleList,
  useAlertStatusList,
  useAlertHistoryList,
  useAlertHistoryInfinite
} from './use-alert-list';

// Single resource hook with validation
export { useAlertRule } from './useAlertRule';
