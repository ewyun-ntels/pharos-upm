/**
 * Timestamp Utility
 * 
 * 프론트엔드에서 타임스탬프를 일관되게 처리하기 위한 유틸리티
 * 
 * 설계 원칙:
 * - 내부 처리는 항상 밀리초 (JavaScript Date.getTime()과 동일)
 * - 백엔드 통신 시 초로 변환
 * - 함수명에 단위 명시
 * 
 * 주의: DateTimeRangeValue 관련 함수는 datetime-range.tsx에 있습니다.
 * - dateTimeToMs()
 * - getAbsoluteValueTimestamp()
 * - getAbsoluteValueTimestampMs()
 * 
 * @see /docs/TIMESTAMP_DESIGN.md
 */

/**
 * 타임스탬프가 초인지 밀리초인지 판단하는 기준값
 * 
 * 2001년 9월 9일 = 1,000,000,000초 = 1,000,000,000,000밀리초
 * 이 값보다 크면 밀리초로 간주
 */
const TIMESTAMP_THRESHOLD = 1e10;
const EXPLICIT_TIMEZONE_PATTERN = /(?:Z|[+-]\d{2}:?\d{2})$/i;

/**
 * 다양한 형식의 타임스탬프를 밀리초로 정규화
 * 
 * 지원 형식:
 * - ISO 8601 문자열: "2026-03-05T05:00:00Z"
 * - Unix seconds: 1772686800
 * - Unix milliseconds: 1772686800000
 * - Date 객체
 * 
 * @param value - 타임스탬프 (다양한 형식)
 * @returns 밀리초 타임스탬프
 * 
 * @example
 * parseTimestampToMs("2026-03-05T05:00:00Z")  // 1772686800000
 * parseTimestampToMs(1772686800)               // 1772686800000
 * parseTimestampToMs(1772686800000)            // 1772686800000
 */
export function parseTimestampToMs(value: unknown): number {
  // Date 객체
  if (value instanceof Date) {
    return value.getTime();
  }

  // ISO 8601 문자열
  if (typeof value === 'string') {
    const date = new Date(value);
    if (!isNaN(date.getTime())) {
      // UTC 시간이면 그대로, 로컬 시간이면 타임존 오프셋 적용
      const hasExplicitTimezone = value.includes('T') && EXPLICIT_TIMEZONE_PATTERN.test(value);
      return hasExplicitTimezone
        ? date.getTime()
        : date.getTime() - new Date().getTimezoneOffset() * 60 * 1000;
    }
    // 숫자 문자열일 수도 있음
    const num = Number(value);
    if (!isNaN(num)) {
      return num > TIMESTAMP_THRESHOLD ? num : num * 1000;
    }
    return 0;
  }

  // 숫자 (초 또는 밀리초)
  if (typeof value === 'number') {
    // 1e10보다 크면 밀리초, 작으면 초로 간주
    return value > TIMESTAMP_THRESHOLD ? value : value * 1000;
  }

  return 0;
}

/**
 * 밀리초를 초로 변환 (백엔드 API 전송용)
 * 
 * @param timestampMs - 밀리초 타임스탬프
 * @returns 초 타임스탬프
 * 
 * @example
 * toUnixSeconds(1772686800000)  // 1772686800
 */
export function toUnixSeconds(timestampMs: number): number {
  return Math.floor(timestampMs / 1000);
}

/**
 * 초를 밀리초로 변환
 * 
 * @param timestampSec - 초 타임스탬프
 * @returns 밀리초 타임스탬프
 * 
 * @example
 * toUnixMilliseconds(1772686800)  // 1772686800000
 */
export function toUnixMilliseconds(timestampSec: number): number {
  return timestampSec * 1000;
}

/**
 * 타임스탬프가 초인지 밀리초인지 자동 판단하여 밀리초로 정규화
 * 
 * @deprecated 새 코드에서는 사용하지 마세요. parseTimestampToMs를 사용하세요.
 * 레거시 코드 지원 목적으로만 제공됩니다.
 * 
 * @param value - 숫자 타임스탬프
 * @returns 밀리초 타임스탬프
 */
export function normalizeTimestamp(value: number): number {
  return value > TIMESTAMP_THRESHOLD ? value : value * 1000;
}

/**
 * 밀리초 타임스탬프를 로컬 시간 문자열로 변환
 * 
 * @param timestampMs - 밀리초 타임스탬프
 * @param locale - 로케일 (기본값: 브라우저 로케일)
 * @returns 로컬 시간 문자열
 * 
 * @example
 * formatTimestamp(1772686800000)  // "2026-03-05 14:00:00"
 */
export function formatTimestamp(
  timestampMs: number,
  locale: string = navigator.language
): string {
  const date = new Date(timestampMs);
  return date.toLocaleString(locale, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

/**
 * 두 타임스탬프 사이의 시간 차이를 초로 반환
 * 
 * @param startMs - 시작 시간 (밀리초)
 * @param endMs - 종료 시간 (밀리초)
 * @returns 시간 차이 (초)
 * 
 * @example
 * getTimeDiffSeconds(1772686800000, 1772690400000)  // 3600 (1시간)
 */
export function getTimeDiffSeconds(startMs: number, endMs: number): number {
  return Math.floor((endMs - startMs) / 1000);
}
