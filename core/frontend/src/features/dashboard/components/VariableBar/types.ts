/**
 * Variable Bar Common Types
 * 
 * Architecture: Direct usage of schema-generated FilterConfig
 * - Schema: shared/schema/dashboard/FilterConfig.schema
 * - Generated: @pharos/shared/types/dashboard
 * - Pattern: FilterConfig (common) + options (type-specific)
 * 
 * Note: FilterConfig is directly imported from @pharos/shared/types/dashboard
 * No need for UI-specific wrapper types (following Chart pattern)
 * 
 * Plugin-specific types (Options, constants) are defined in each plugin's types.ts:
 * - variables/step/types.ts: StepVariableOptions, STEP, STEP_DEFAULT, StepType
 * - variables/refresh/types.ts: RefreshVariableOptions, REFRESH_INTERVAL, REFRESH_INTERVAL_DEFAULT
 * - variables/select/types.ts: SelectVariableOptions (supports both single and multi-select via isMulti option)
 * - variables/datetime/types.ts: DateTimeVariableOptions
 * - variables/custom/types.ts: CustomVariableOptions
 * - variables/tableSearch/types.ts: TableSearchVariableOptions
 */

import type { FilterConfig } from '@pharos/shared/types/dashboard';
import React from 'react';

// Re-export FilterConfig from schema for convenience
export type { FilterConfig };

/**
 * VariableBar Component Props
 */
export type VariableBarBodyProps = {
  children?: React.ReactNode;
};

export interface LayoutHeaderProps {
  dashboardId: string;
  leftItems: FilterConfig[];
  rightItems: FilterConfig[];
}
