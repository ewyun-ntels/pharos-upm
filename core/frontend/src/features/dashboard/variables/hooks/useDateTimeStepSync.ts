import {useCallback, useEffect} from 'react';
import {useDashboardStore} from '@features/dashboard/hooks/use-dashboard-store';
import type {DateTimeRangeValue} from '@pharos/shared/components/ui-extension';
import {getAbsoluteValueTimestampMs} from '@pharos/shared/components/ui-extension';
import {
  findBestRangeKey,
  getMinRangeKey,
} from '@features/dashboard/utils/datetime-step-utils';

/**
 * DateTime-Step 통합 관리 Hook
 *
 * DateTime 필터와 Step 필터의 연동 로직을 담당합니다.
 * - rangeStepOptions 기반으로 stepOptions 초기화
 * - datetime 변경 시 적절한 stepOptions 자동 선택
 *
 * 이 hook은 DateTimeFilter와 StepFilter에서 함께 사용됩니다.
 */
export const useDateTimeStepSync = () => {
  const datetime = useDashboardStore((state) => state.filterState?.datetime);
  const rangeStepOptions = useDashboardStore((state) => state.filterState?.rangeStepOptions);
  const setStepOptions = useDashboardStore((state) => state.setStepOptions);

  // Initialize step options from rangeStepOptions
  useEffect(() => {
    if (rangeStepOptions && Object.keys(rangeStepOptions).length !== 0) {
      const minRangeKey = getMinRangeKey(rangeStepOptions);

      if (minRangeKey) {
        const steps = rangeStepOptions[minRangeKey];
        setStepOptions(steps);
      } else {
        setStepOptions([]);
      }
    } else {
      setStepOptions([]);
    }
  }, [rangeStepOptions, setStepOptions]);

  // Step options를 datetime range에 따라 업데이트
  const updateStepOptionsByRange = useCallback(
    (startTime: DateTimeRangeValue, endTime: DateTimeRangeValue) => {
      if (!rangeStepOptions) {
        return;
      }
      const start = getAbsoluteValueTimestampMs(startTime);
      const end = getAbsoluteValueTimestampMs(endTime);

      const diffMs = Math.abs(end - start);
      const bestKey = findBestRangeKey(rangeStepOptions, diffMs);

      if (bestKey) {
        setStepOptions(rangeStepOptions[bestKey]);
        // step 자동 변경 제거: stepOptions만 업데이트하고 step 값은 유지
        // 기존 문제: datetime이 변경될 때마다 step이 자동으로 덮어써짐
        // 수정: stepOptions만 업데이트하여 사용자 선택을 존중
      } else {
        setStepOptions([]);
      }
    },
    [rangeStepOptions, setStepOptions],
  );

  useEffect(() => {
    if (!datetime) return;
    updateStepOptionsByRange(
      datetime.startTime as DateTimeRangeValue,
      datetime.endTime as DateTimeRangeValue,
    );
  }, [datetime, updateStepOptionsByRange]);
};
