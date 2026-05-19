import { DateTimeRangeValue } from '@pharos/shared/components/ui-extension';

/**
 * 필터 값 타입
 * 
 * - string: select (single), input, search 등 단일 값
 * - string[]: select (multi) 등 다중 값
 * 
 * 참고: 쿼리 생성 시 모두 string으로 변환됨
 * - string: "'value'"
 * - string[]: "'val1','val2','val3'"
 */
export type FilterValue = string | string[];

/**
 * DateTime 타입
 */
export type DateTime = {
  startTime: DateTimeRangeValue;
  endTime: DateTimeRangeValue;
};

/**
 * Step 타입
 *
 * StepType: UI에서 기본 제공하는 predefined 옵션 목록 (STEP 상수와 동일)
 * Step.step: 실제 선택된 값. rangeStepOptions로 임의 duration 문자열('2h' 등)이
 *            올 수 있으므로 string으로 선언한다.
 */
export type StepType = '15s' | '30s' | '1m' | '5m' | '15m' | '30m' | '1h';
export type Step = {
  step: string;
  stepValue: number;
};

