/**
 * Notification Feature Hooks
 *
 * Notification 기능 관련 모든 hooks를 export합니다.
 *
 * ## Mutation Hooks (CUD operations)
 * - useCreateNotificationRule: Notification Rule 생성
 * - useUpdateNotificationRule: Notification Rule 수정
 * - useDeleteNotificationRule: Notification Rule 삭제
 *
 * ## Query Hooks (Read operations)
 * - useNotificationRuleList: Notification Rule 목록 조회
 * - useNotificationRule: Notification Rule 단일 조회
 */

// Mutation hooks
export {
  useCreateNotificationRule,
  useUpdateNotificationRule,
  useDeleteNotificationRule
} from './use-notification-rule-mutation';

// Query hooks
export {
  useNotificationRuleList,
  useNotificationRule
} from './use-notification-list';
