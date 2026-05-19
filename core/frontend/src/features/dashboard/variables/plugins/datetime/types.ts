import {DateTimeRangeValue} from '@pharos/shared/components/ui-extension';

/**
 * DateTime Filter Options
 */
export interface DateTimeFilterOptions {
  defaultStartTime?: DateTimeRangeValue;
  defaultEndTime?: DateTimeRangeValue;
  // ✅ 저장된 time range (URL string format)
  savedStartTime?: string;  // e.g., "now-1h", "now-24h"
  savedEndTime?: string;    // e.g., "now"
}
