/**
 * Timestamp Utility Tests
 */

import {
  parseTimestampToMs,
  toUnixSeconds,
  toUnixMilliseconds,
  normalizeTimestamp,
  getTimeDiffSeconds,
} from './timestamp';
import { getAbsoluteValueTimestampMs } from '@pharos/shared/components/ui-extension';
import type { DateTimeRangeValue } from '@pharos/shared/components/ui-extension';

describe('timestamp utilities', () => {
  describe('parseTimestampToMs', () => {
    it('ISO 8601 UTC 문자열을 밀리초로 변환', () => {
      const result = parseTimestampToMs('2026-03-05T05:00:00Z');
      expect(result).toBe(new Date('2026-03-05T05:00:00Z').getTime());
    });

    it('ISO 8601 offset strings are not shifted again', () => {
      const result = parseTimestampToMs('2026-05-19T18:40:00+09:00');
      expect(result).toBe(new Date('2026-05-19T18:40:00+09:00').getTime());
    });

    it('ISO 8601 compact offset strings are not shifted again', () => {
      const result = parseTimestampToMs('2026-05-19T04:40:00-0500');
      expect(result).toBe(new Date('2026-05-19T04:40:00-0500').getTime());
    });

    it('Unix seconds를 밀리초로 변환', () => {
      const seconds = 1772686800;
      const result = parseTimestampToMs(seconds);
      expect(result).toBe(seconds * 1000);
    });

    it('Unix milliseconds를 그대로 반환', () => {
      const milliseconds = 1772686800000;
      const result = parseTimestampToMs(milliseconds);
      expect(result).toBe(milliseconds);
    });

    it('Date 객체를 밀리초로 변환', () => {
      const date = new Date('2026-03-05T05:00:00Z');
      const result = parseTimestampToMs(date);
      expect(result).toBe(date.getTime());
    });

    it('숫자 문자열을 변환 (초)', () => {
      const result = parseTimestampToMs('1772686800');
      expect(result).toBe(1772686800000);
    });

    it('숫자 문자열을 변환 (밀리초)', () => {
      const result = parseTimestampToMs('1772686800000');
      expect(result).toBe(1772686800000);
    });

    it('잘못된 입력에 대해 0 반환', () => {
      expect(parseTimestampToMs(null)).toBe(0);
      expect(parseTimestampToMs(undefined)).toBe(0);
      expect(parseTimestampToMs('invalid')).toBe(0);
    });
  });

  describe('toUnixSeconds', () => {
    it('밀리초를 초로 변환', () => {
      const result = toUnixSeconds(1772686800000);
      expect(result).toBe(1772686800);
    });

    it('소수점 버림', () => {
      const result = toUnixSeconds(1772686800999);
      expect(result).toBe(1772686800);
    });
  });

  describe('toUnixMilliseconds', () => {
    it('초를 밀리초로 변환', () => {
      const result = toUnixMilliseconds(1772686800);
      expect(result).toBe(1772686800000);
    });
  });

  describe('getAbsoluteValueTimestampMs', () => {
    it('absolute 타입 - Date 객체', () => {
      const date = new Date('2026-03-05T05:00:00Z');
      const value: DateTimeRangeValue = {
        type: 'absolute',
        absoluteValue: date,
      };
      const result = getAbsoluteValueTimestampMs(value);
      expect(result).toBe(date.getTime());
    });

    it('relative 타입 - now', () => {
      const value: DateTimeRangeValue = {
        type: 'relative',
        relativeNow: true,
      };
      const result = getAbsoluteValueTimestampMs(value);
      expect(result).toBeCloseTo(Date.now(), -2); // 100ms 오차 허용
    });

    it('relative 타입 - 1 hour ago', () => {
      const value: DateTimeRangeValue = {
        type: 'relative',
        relativeValue: 1,
        relativeFormat: 'Hours ago',
      };
      const result = getAbsoluteValueTimestampMs(value);
      const expected = Date.now() - 60 * 60 * 1000;
      expect(result).toBeCloseTo(expected, -4); // 10초 오차 허용
    });
  });

  describe('normalizeTimestamp', () => {
    it('초를 밀리초로 변환', () => {
      const result = normalizeTimestamp(1772686800);
      expect(result).toBe(1772686800000);
    });

    it('밀리초는 그대로 반환', () => {
      const result = normalizeTimestamp(1772686800000);
      expect(result).toBe(1772686800000);
    });
  });

  describe('getTimeDiffSeconds', () => {
    it('두 타임스탬프 사이의 초 계산', () => {
      const start = 1772686800000; // 2026-03-05 14:00:00
      const end = 1772690400000;   // 2026-03-05 15:00:00
      const result = getTimeDiffSeconds(start, end);
      expect(result).toBe(3600); // 1시간
    });

    it('음수 차이도 계산', () => {
      const start = 1772690400000;
      const end = 1772686800000;
      const result = getTimeDiffSeconds(start, end);
      expect(result).toBe(-3600);
    });
  });

  describe('boundary cases', () => {
    it('2001년 9월 9일 (임계값)', () => {
      const threshold = 1e10;
      
      // 초로 간주 (임계값보다 작음)
      const seconds = threshold - 1;
      expect(parseTimestampToMs(seconds)).toBe(seconds * 1000);
      
      // 밀리초로 간주 (임계값보다 큼)
      const milliseconds = threshold + 1;
      expect(parseTimestampToMs(milliseconds)).toBe(milliseconds);
    });

    it('0은 0으로 변환', () => {
      expect(parseTimestampToMs(0)).toBe(0);
    });
  });
});
