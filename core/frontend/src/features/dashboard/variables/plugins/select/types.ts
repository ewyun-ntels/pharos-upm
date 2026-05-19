import type { SelectOption } from '@pharos/shared/components';

/**
 * Select Filter Options
 * 
 * FilterConfig의 options 필드에 저장되는 플러그인별 옵션
 * UnifiedDashboardSelect에 전달되는 props와는 별개입니다.
 */
export interface SelectFilterOptions {
  // SelectBox로 전달되는 옵션
  options?: SelectOption[] | string[];
  labelKey?: string;
  valueKey?: string;

  // Select filter 전용 옵션
  initValue?: string | string[]; // Support both single and multi
  showAllOption?: boolean;
  defaultAllSelected?: boolean; // When showAllOption=true, default to "All" selected (only on initial render)
  customAllValue?: string; // Custom value to use when "All" is selected (e.g., "*", ".*"), instead of all individual values
  isSearchable?: boolean;
  isHidden?: boolean;
  autoSelectFirstOption?: boolean;
  
  // Multi-select support
  isMulti?: boolean; // ✅ Enable multi-selection
  showCheckbox?: boolean; // ✅ Show checkbox in multi-select mode
  displayMode?: 'count' | 'value-single' | 'value-all' | 'value-threshold'; // ✅ How to display selected values
  // - 'count': Always show count (e.g., "3")
  // - 'value-single': Show value if 1 selected, count if 2+ (e.g., "server-1" or "2")
  // - 'value-all': Show all values comma-separated (e.g., "server-1, server-2")
  // - 'value-threshold': Show values up to threshold, then count (default, threshold=1)
  displayThreshold?: number; // Threshold for 'value-threshold' mode (default: 1)
}
