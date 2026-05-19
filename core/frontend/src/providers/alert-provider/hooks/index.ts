/**
 * Alert Provider Custom Hooks
 * 
 * These hooks wrap Refine's data hooks with Zod validation.
 * Use these instead of calling the provider directly for better type safety.
 * 
 * @example
 * import { useCreateAlertRule, useMaskAlertStatus } from '@/providers/alert-provider/hooks';
 */

export { useCreateAlertRule, useUpdateAlertRule } from './useAlertRuleMutation';
export { useUpdateAlertStatus, useMaskAlertStatus } from './useAlertStatusMutation';
