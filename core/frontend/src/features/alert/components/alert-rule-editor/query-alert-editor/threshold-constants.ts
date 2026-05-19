/**
 * Threshold Editor Constants
 * 
 * Zod Schema에서 자동 생성된 상수들
 * Schema가 변경되면 자동으로 타입 체크됨
 */

import {
  type Severity,
  type Condition,
  SeveritySchema,
  ConditionSchema,
} from '@pharos/shared/types/alert';

/**
 * Severity Color Mapping
 */
const SEVERITY_COLOR_MAP: Record<Severity, string> = {
  Critical: 'text-red-600',
  Major: 'text-orange-600',
  Minor: 'text-yellow-600',
  Normal: 'text-green-600',
};

/**
 * Severity Display Configuration
 */
export const SEVERITY_OPTIONS: {
  value: Severity;
  label: string;
  color: string;
}[] = SeveritySchema.options.map((severity) => ({
  value: severity,
  label: severity,
  color: SEVERITY_COLOR_MAP[severity],
}));

/**
 * Severity Border Color Mapping (Hex)
 */
const SEVERITY_BORDER_COLOR_MAP: Record<Severity, string> = {
  Critical: '#dc2626',
  Major: '#ea580c',
  Minor: '#ca8a04',
  Normal: '#16a34a',
};

/**
 * Get Severity Border Color
 */
export function getSeverityColor(severity: Severity): string {
  return SEVERITY_BORDER_COLOR_MAP[severity];
}

/**
 * Condition Label Mapping
 */
const CONDITION_LABEL_MAP: Record<Condition, string> = {
  is_above: 'Is Above',
  is_below: 'Is Below',
  within_range: 'Within Range',
  outside_range: 'Outside Range',
};

/**
 * Condition Description Mapping
 */
const CONDITION_DESCRIPTION_MAP: Record<Condition, string> = {
  is_above: 'Trigger when value > threshold',
  is_below: 'Trigger when value < threshold',
  within_range: 'Trigger when value is between start and end',
  outside_range: 'Trigger when value is outside start and end',
};

/**
 * Condition Display Configuration
 */
export const CONDITION_OPTIONS: {
  value: Condition;
  label: string;
  description: string;
}[] = ConditionSchema.options.map((condition) => ({
  value: condition,
  label: CONDITION_LABEL_MAP[condition],
  description: CONDITION_DESCRIPTION_MAP[condition],
}));
