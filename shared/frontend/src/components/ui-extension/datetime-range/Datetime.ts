export type DateTimeRangeType = 'absolute' | 'relative';

export type TimeUnit = 'seconds' | 'milliseconds' | 'nanoseconds';

export type RangeType = '1year' | '1month' | '1day' | '1hour';

export type ValidationErrorType = 'validation' | 'range' | 'empty' | null;

export type DateTimeRelativeFormat =
  | 'Seconds ago'
  | 'Minutes ago'
  | 'Hours ago'
  | 'Days ago'
  | 'Weeks ago'
  | 'Months ago'
  | 'Years ago'
  // | 'Seconds from now'
  // | 'Minutes from now'
  // | 'Hours from now'
  // | 'Days from now'
  // | 'Weeks from now'
  // | 'Months from now'
  // | 'Years from now';