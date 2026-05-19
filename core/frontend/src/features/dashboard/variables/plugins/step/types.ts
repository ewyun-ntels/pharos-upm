import type { StepType } from '@features/dashboard/types';

/**
 * Step Filter Options
 */
export interface StepFilterOptions {
  initValue?: string;
  rangeStepOptions?: Record<string, string[]>;
  savedValue?: string;  // e.g., "5m", "1h", "auto", or custom rangeStepOptions value
}

export const STEP: StepType[] = ['15s', '30s', '1m', '5m', '15m', '30m', '1h'];
export const STEP_DEFAULT = '5m';
