import {produce} from 'immer';

/**
 * 패널 options 업데이트 헬퍼
 *
 * 사용:
 *   const handleChange = useOptionsChange(options, onOptionsChange);
 *   handleChange(draft => { draft.unit = value; });
 *   handleChange(draft => { draft.scales.yMin = value; });
 */
export function useOptionsChange<T extends object>(
  options: T,
  onOptionsChange: ((newOptions: T) => void) | undefined,
) {
  return (updater: (draft: T) => void) => {
    onOptionsChange?.(produce(options, updater));
  };
}
