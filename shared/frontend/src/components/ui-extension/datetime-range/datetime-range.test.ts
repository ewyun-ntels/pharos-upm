import { dateTimeValueToUrlString, urlStringToDateTimeValue } from './datetime-range';

describe('dateTimeValueToUrlString', () => {
  describe('relative time - relativeNow', () => {
    it('should return "now" when relativeNow is true', () => {
      const result = dateTimeValueToUrlString({
        type: 'relative',
        relativeNow: true,
      });
      expect(result).toBe('now');
    });
  });

  describe('relative time - with relativeFormat (needs conversion)', () => {
    it('should convert "1 week ago" to "now-1w"', () => {
      const result = dateTimeValueToUrlString({
        type: 'relative',
        relativeValue: '1',
        relativeFormat: 'Weeks ago',
        relativeNow: false,
      });
      expect(result).toBe('now-1w');
    });

    it('should convert "5 minutes ago" to "now-5m"', () => {
      const result = dateTimeValueToUrlString({
        type: 'relative',
        relativeValue: '5',
        relativeFormat: 'Minutes ago',
        relativeNow: false,
      });
      expect(result).toBe('now-5m');
    });

    it('should convert "2 hours ago" to "now-2h"', () => {
      const result = dateTimeValueToUrlString({
        type: 'relative',
        relativeValue: '2',
        relativeFormat: 'Hours ago',
        relativeNow: false,
      });
      expect(result).toBe('now-2h');
    });

    it('should convert "7 days ago" to "now-7d"', () => {
      const result = dateTimeValueToUrlString({
        type: 'relative',
        relativeValue: '7',
        relativeFormat: 'Days ago',
        relativeNow: false,
      });
      expect(result).toBe('now-7d');
    });

    it('should convert "3 months ago" to "now-3M"', () => {
      const result = dateTimeValueToUrlString({
        type: 'relative',
        relativeValue: '3',
        relativeFormat: 'Months ago',
        relativeNow: false,
      });
      expect(result).toBe('now-3M');
    });

    it('should convert "1 year ago" to "now-1y"', () => {
      const result = dateTimeValueToUrlString({
        type: 'relative',
        relativeValue: '1',
        relativeFormat: 'Years ago',
        relativeNow: false,
      });
      expect(result).toBe('now-1y');
    });

    it('should convert "30 seconds ago" to "now-30s"', () => {
      const result = dateTimeValueToUrlString({
        type: 'relative',
        relativeValue: '30',
        relativeFormat: 'Seconds ago',
        relativeNow: false,
      });
      expect(result).toBe('now-30s');
    });

    it('should handle numeric relativeValue', () => {
      const result = dateTimeValueToUrlString({
        type: 'relative',
        relativeValue: 15,
        relativeFormat: 'Minutes ago',
        relativeNow: false,
      });
      expect(result).toBe('now-15m');
    });

    it('should default to 1 when relativeValue is missing', () => {
      const result = dateTimeValueToUrlString({
        type: 'relative',
        relativeFormat: 'Hours ago',
        relativeNow: false,
      });
      expect(result).toBe('now-1h');
    });
  });

  describe('relative time - pre-formatted string (no relativeFormat)', () => {
    it('should handle pre-formatted offset "-5m"', () => {
      const result = dateTimeValueToUrlString({
        type: 'relative',
        relativeValue: '-5m',
        relativeNow: false,
      });
      expect(result).toBe('now-5m');
    });

    it('should handle pre-formatted offset starting with "now"', () => {
      const result = dateTimeValueToUrlString({
        type: 'relative',
        relativeValue: 'now-1h',
        relativeNow: false,
      });
      expect(result).toBe('now-1h');
    });

    it('should handle "now" string as relativeValue', () => {
      const result = dateTimeValueToUrlString({
        type: 'relative',
        relativeValue: 'now',
        relativeNow: false,
      });
      expect(result).toBe('now');
    });
  });

  describe('absolute time', () => {
    it('should convert Date object to Unix timestamp', () => {
      const testDate = new Date('2026-03-05T12:00:00.000Z');
      const result = dateTimeValueToUrlString({
        type: 'absolute',
        absoluteValue: testDate,
      });
      expect(result).toBe(testDate.getTime().toString());
    });

    it('should convert date string to Unix timestamp', () => {
      const dateString = '2026-03-05T12:00:00.000Z';
      const expectedTimestamp = new Date(dateString).getTime().toString();
      const result = dateTimeValueToUrlString({
        type: 'absolute',
        absoluteValue: new Date(dateString),
      });
      expect(result).toBe(expectedTimestamp);
    });

    it('should handle missing absoluteValue by using current time', () => {
      const beforeCall = Date.now();
      const result = dateTimeValueToUrlString({
        type: 'absolute',
      });
      const afterCall = Date.now();
      
      const timestamp = parseInt(result, 10);
      expect(timestamp).toBeGreaterThanOrEqual(beforeCall);
      expect(timestamp).toBeLessThanOrEqual(afterCall);
    });
  });

  describe('edge cases', () => {
    it('should handle zero value by defaulting to 1', () => {
      const result = dateTimeValueToUrlString({
        type: 'relative',
        relativeValue: 0,
        relativeFormat: 'Minutes ago',
        relativeNow: false,
      });
      // 0 is falsy, so it defaults to 1
      expect(result).toBe('now-1m');
    });

    it('should handle large values', () => {
      const result = dateTimeValueToUrlString({
        type: 'relative',
        relativeValue: '999',
        relativeFormat: 'Days ago',
        relativeNow: false,
      });
      expect(result).toBe('now-999d');
    });
  });
});

describe('urlStringToDateTimeValue', () => {
  describe('parsing "now"', () => {
    it('should parse "now" correctly', () => {
      const result = urlStringToDateTimeValue('now');
      expect(result).toEqual({
        type: 'relative',
        relativeNow: true,
      });
    });
  });

  describe('parsing Grafana format (ago)', () => {
    it('should parse "now-2w" to 2 Weeks ago', () => {
      const result = urlStringToDateTimeValue('now-2w');
      expect(result).toEqual({
        type: 'relative',
        relativeValue: '2',
        relativeFormat: 'Weeks ago',
        relativeNow: false,
      });
    });

    it('should parse "now-5m" to 5 Minutes ago', () => {
      const result = urlStringToDateTimeValue('now-5m');
      expect(result).toEqual({
        type: 'relative',
        relativeValue: '5',
        relativeFormat: 'Minutes ago',
        relativeNow: false,
      });
    });

    it('should parse "now-1h" to 1 Hours ago', () => {
      const result = urlStringToDateTimeValue('now-1h');
      expect(result).toEqual({
        type: 'relative',
        relativeValue: '1',
        relativeFormat: 'Hours ago',
        relativeNow: false,
      });
    });

    it('should parse "now-7d" to 7 Days ago', () => {
      const result = urlStringToDateTimeValue('now-7d');
      expect(result).toEqual({
        type: 'relative',
        relativeValue: '7',
        relativeFormat: 'Days ago',
        relativeNow: false,
      });
    });

    it('should parse "now-3M" to 3 Months ago', () => {
      const result = urlStringToDateTimeValue('now-3M');
      expect(result).toEqual({
        type: 'relative',
        relativeValue: '3',
        relativeFormat: 'Months ago',
        relativeNow: false,
      });
    });

    it('should parse "now-1y" to 1 Years ago', () => {
      const result = urlStringToDateTimeValue('now-1y');
      expect(result).toEqual({
        type: 'relative',
        relativeValue: '1',
        relativeFormat: 'Years ago',
        relativeNow: false,
      });
    });

    it('should parse "now-30s" to 30 Seconds ago', () => {
      const result = urlStringToDateTimeValue('now-30s');
      expect(result).toEqual({
        type: 'relative',
        relativeValue: '30',
        relativeFormat: 'Seconds ago',
        relativeNow: false,
      });
    });
  });

  describe('parsing Grafana format (future/from now)', () => {
    it('should parse "now+5m" as future time', () => {
      const result = urlStringToDateTimeValue('now+5m');
      expect(result).toEqual({
        type: 'relative',
        relativeValue: '+5m',
        relativeNow: false,
      });
    });

    it('should parse "now+1h" as future time', () => {
      const result = urlStringToDateTimeValue('now+1h');
      expect(result).toEqual({
        type: 'relative',
        relativeValue: '+1h',
        relativeNow: false,
      });
    });
  });

  describe('parsing Unix timestamp', () => {
    it('should parse Unix timestamp (milliseconds)', () => {
      const timestamp = 1709640000000; // 2026-03-05 12:00:00 UTC
      const result = urlStringToDateTimeValue(timestamp.toString());
      expect(result).toEqual({
        type: 'absolute',
        absoluteValue: new Date(timestamp),
      });
    });
  });

  describe('roundtrip conversion', () => {
    it('should maintain data through URL roundtrip for 2 weeks ago', () => {
      const original = {
        type: 'relative' as const,
        relativeValue: '2',
        relativeFormat: 'Weeks ago' as const,
        relativeNow: false,
      };
      const urlString = dateTimeValueToUrlString(original);
      const parsed = urlStringToDateTimeValue(urlString);
      
      expect(urlString).toBe('now-2w');
      expect(parsed).toEqual(original);
    });

    it('should maintain data through URL roundtrip for 5 minutes ago', () => {
      const original = {
        type: 'relative' as const,
        relativeValue: '5',
        relativeFormat: 'Minutes ago' as const,
        relativeNow: false,
      };
      const urlString = dateTimeValueToUrlString(original);
      const parsed = urlStringToDateTimeValue(urlString);
      
      expect(urlString).toBe('now-5m');
      expect(parsed).toEqual(original);
    });

    it('should maintain data through URL roundtrip for "now"', () => {
      const original = {
        type: 'relative' as const,
        relativeNow: true,
      };
      const urlString = dateTimeValueToUrlString(original);
      const parsed = urlStringToDateTimeValue(urlString);
      
      expect(urlString).toBe('now');
      expect(parsed).toEqual(original);
    });
  });
});
