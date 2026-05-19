import {DateTimeRelativeFormat, RangeType, ValidationErrorType} from './Datetime'

const priorityOrder: RangeType[] = ["1year", "1month", "1day", "1hour"];

export function getHighestPriority(options: string[]): string | null {
  for (const priority of priorityOrder) {
    if (options.includes(priority)) {
      return priority;
    }
  }
  return null;
}

export function getMaxAllowedRange(rangeType: string | null): number {
  switch (rangeType) {
    case '1year':
      return 365 * 24 * 60 * 60 * 1000;
    case '1month':
      return 30 * 24 * 60 * 60 * 1000;
    case '1day':
      return 24 * 60 * 60 * 1000;
    case '1hour':
      return 60 * 60 * 1000;
    default:
      return Infinity;
  }
}

export const getMaxRelativeValue = (format: DateTimeRelativeFormat, checkRangeDate: string | null): number => {
    if (!checkRangeDate) return Infinity;

    const SECONDS_IN = {
      SECOND: 1,
      MINUTE: 60,
      HOUR: 3600,          // 60 * 60
      DAY: 86400,          // 24 * 60 * 60
      WEEK: 604800,        // 7 * 24 * 60 * 60
      MONTH: 2592000,      // 30 * 24 * 60 * 60 
      YEAR: 31536000       // 365 * 24 * 60 * 60
    };
    
    const maxRangeSeconds = {
      '1hour': SECONDS_IN.HOUR,
      '1day': SECONDS_IN.DAY,
      '1month': SECONDS_IN.MONTH,
      '1year': SECONDS_IN.YEAR
    };
    
    const maxSeconds = maxRangeSeconds[checkRangeDate as keyof typeof maxRangeSeconds];
    if (!maxSeconds) return Infinity;
    
    let unitSeconds: number;
    if (format.includes('Seconds')) unitSeconds = SECONDS_IN.SECOND;
    else if (format.includes('Minutes')) unitSeconds = SECONDS_IN.MINUTE;
    else if (format.includes('Hours')) unitSeconds = SECONDS_IN.HOUR;
    else if (format.includes('Days')) unitSeconds = SECONDS_IN.DAY;
    else if (format.includes('Weeks')) unitSeconds = SECONDS_IN.WEEK;
    else if (format.includes('Months')) unitSeconds = SECONDS_IN.MONTH;
    else if (format.includes('Years')) unitSeconds = SECONDS_IN.YEAR;
    else return Infinity;
    
    const maxValue = Math.floor(maxSeconds / unitSeconds);
    return maxValue;
  };

export function getRangeErrorMessage(
  errorType: ValidationErrorType, 
  title?: string,
  checkRangeDate?: string | null
): string {
  if (errorType === 'empty') {
    return 'Please enter the time.';
  } else if (errorType === 'validation') {
    return title === 'Start Date' 
      ? 'Start time must be earlier than the end time.'
      : 'End time must be later than the start time.';
  } else if (errorType === 'range') {
    switch (checkRangeDate) {
      case '1year':
        return 'The date can be set for up to 365 days.';
      case '1month':
        return 'The date can be set for up to 30 days.';
      case '1day':
        return 'The date can be set for up to 24 hours.';
      case '1hour':
        return 'The date can be set for up to 60 minutes.';
      default:
        return 'Invalid date range.';
    }
  }
  return 'Invalid date range.';
}

// TODO translate
// const getDateTimeRelativeFormatLabel = (value:  DateTimeRelativeFormat | "Now", locale: string): string => {
//   switch (value) {
//     case "Now":
//       return "Now";
//     case "Seconds ago":
//       return "초 전";
//     case "Minutes ago":
//       return "분 전";
//     case "Hours ago":
//       return "시간 전";
//     case "Days ago":
//       return "일 전";
//     case "Weeks ago":
//       return "주 전";
//     case "Months ago":
//       return "달 전";
//     case "Years ago":
//       return "년 전";
//     case "Seconds from now":
//       return "초 후";
//     case "Minutes from now":
//       return "분 후";
//     case "Hours from now":
//       return "시간 후";
//     case "Days from now":
//       return "일 후";
//     case "Weeks from now":
//       return "주 후";
//     case "Months from now":
//       return "달 후";
//     case "Years from now":
//       return "년 후";
//     default:
//       return value;
//   }
// }