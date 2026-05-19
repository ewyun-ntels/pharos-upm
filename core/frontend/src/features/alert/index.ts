/**
 * Alert Feature
 * 
 * Alert 기능의 모든 exports를 관리합니다.
 * 
 * ## 구조
 * - hooks/: Custom hooks (useAlertRuleList, useCreateAlertRule, etc.)
 * - components/: UI components (AlertTableTabs, etc.)
 * - types/: Type re-exports from shared
 * 
 * ## 사용법
 * 
 * ### Hooks
 * ```tsx
 * import { useAlertRuleList, useCreateAlertRule } from '@features/alert';
 * 
 * function MyComponent() {
 *   const { data, isLoading } = useAlertRuleList();
 *   const { mutate } = useCreateAlertRule();
 *   // ...
 * }
 * ```
 * 
 * ### Components
 * ```tsx
 * import { AlertTableTabs } from '@features/alert';
 * 
 * function AlertPage() {
 *   return <AlertTableTabs />;
 * }
 * ```
 */

// Hooks
export * from './hooks';

// Components
export * from './components';

// Types (re-export from shared)
export type {
  AlertRule,
  AlertValue,
  AlertType,
  Severity,
  Status,
  QueryAlertRule,
  EventHistoryAlertRule,
  EventStatusAlertRule,
  AlertStatusMaskRequest,
  AlertHistoryRequest,
} from '@pharos/shared/types/alert';

// Constants (re-export from provider)
export { ALERT_RESOURCES } from '@providers/alert-provider';
