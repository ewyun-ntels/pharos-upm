import * as chrono from 'chrono-node';
import {useState, useEffect} from 'react';
import {cn} from '../../../lib';
import {Popover, PopoverContent, PopoverTrigger} from '@pharos/shared/components/ui';
import {Button} from '@pharos/shared/components/ui';
import {Tooltip, TooltipContent, TooltipProvider, TooltipTrigger} from '../tooltip-custom';
import {CalendarDays} from 'lucide-react';
import {Tabs, TabsContent, TabsList, TabsTrigger} from '@pharos/shared/components/ui';
import {DateTimePanel} from './datetime-panel';
import {Input} from '@pharos/shared/components/ui';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@pharos/shared/components/ui';
import {Switch} from '@pharos/shared/components/ui';
import {
  getHighestPriority,
  getMaxAllowedRange,
  getRangeErrorMessage as utilsGetRangeErrorMessage,
  getMaxRelativeValue,
} from './date-range-utils';
import {DateTimeRangeType, DateTimeRelativeFormat, ValidationErrorType} from './Datetime';

export const dateTimeRelativeFormats: DateTimeRelativeFormat[] = [
  'Seconds ago',
  'Minutes ago',
  'Hours ago',
  'Days ago',
  'Weeks ago',
  'Months ago',
  // 'Years ago',
  // 'Seconds from now',
  // 'Minutes from now',
  // 'Hours from now',
  // 'Days from now',
  // 'Weeks from now',
  // 'Months from now',
  // 'Years from now',
];
const getDateTimeRelativeFormatLabel = (
  value: DateTimeRelativeFormat | 'Now',
  _locale: string,
): string => {
  switch (value) {
    case 'Now':
      return 'Now';
    case 'Seconds ago':
      return 'Seconds ago';
    case 'Minutes ago':
      return 'Minutes ago';
    case 'Hours ago':
      return 'Hours ago';
    case 'Days ago':
      return 'Days ago';
    case 'Weeks ago':
      return 'Weeks ago';
    case 'Months ago':
      return 'Months ago';
    case 'Years ago':
      return 'Years ago';
    // case 'Seconds from now':
    //   return 'Seconds from now';
    // case 'Minutes from now':
    //   return 'Minutes from now';
    // case 'Hours from now':
    //   return 'Hours from now';
    // case 'Days from now':
    //   return 'Days from now';
    // case 'Weeks from now':
    //   return 'Weeks from now';
    // case 'Months from now':
    //   return 'Months from now';
    // case 'Years from now':
    //   return 'Years from now';
    default:
      return value;
  }
};

type DateTimeRangeValue = {
  type: DateTimeRangeType;
  absoluteValue?: Date;
  relativeValue?: number | string;
  relativeFormat?: DateTimeRelativeFormat;
  relativeNow?: boolean;
  customValue?: string;
};

/**
 * DateTimeRangeValue를 Unix 타임스탬프(밀리초)로 변환
 * 
 * @param value - DateTimeRangeValue (absolute 또는 relative)
 * @returns Unix timestamp in milliseconds
 * 
 * @example
 * getAbsoluteValueTimestampMs({type: 'relative', relativeNow: true})  // 1234567890000
 * getAbsoluteValueTimestampMs({type: 'absolute', absoluteValue: new Date()})  // 1234567890000
 */
export function getAbsoluteValueTimestampMs(value: DateTimeRangeValue): number {
  if (value.type === 'absolute') {
    if (value.absoluteValue) {
      if (value.absoluteValue instanceof Date) {
        return value.absoluteValue.getTime();
      } else {
        return new Date(value.absoluteValue).getTime();
      }
    }
    return Date.now();
  } else if (value.type === 'relative') {
    // "now"
    if (value.relativeNow) {
      return Date.now();
    }

    // 두 가지 형식 지원:
    // 1. String format: relativeValue = "-1h", relativeFormat = undefined
    // 2. Number format: relativeValue = 1, relativeFormat = "Hours ago"
    let relativeValue: string;
    if (value.relativeFormat) {
      relativeValue = `${value.relativeValue} ${value.relativeFormat}`;
    } else {
      relativeValue = String(value.relativeValue);
    }

    const parsedDate = chrono.parseDate(relativeValue);
    if (parsedDate) {
      return parsedDate.getTime();
    }
  }
  
  return Date.now();
}

/**
 * DateTimeRangeValue를 Unix 타임스탬프(초)로 변환
 * 
 * @param value - DateTimeRangeValue
 * @returns Unix timestamp in seconds
 * 
 * @example
 * getAbsoluteValueTimestamp({type: 'relative', relativeNow: true})  // 1234567890 (seconds)
 */
export const getAbsoluteValueTimestamp = (value: DateTimeRangeValue): number => {
  return Math.floor(getAbsoluteValueTimestampMs(value) / 1000);
};

const getAbsoluteValue = (value: DateTimeRangeValue): Date | undefined => {
  if (value.type === 'absolute') {
    if (value.absoluteValue) {
      // 문자열이면 Date로 변환
      if (value.absoluteValue instanceof Date) {
        return value.absoluteValue;
      } else {
        return new Date(value.absoluteValue);
      }
    }
  } else if (value.type === 'relative') {
    if (value.relativeNow) {
      return new Date();
    }

    // Handle both formats:
    // 1. String format: relativeValue = "-1h", relativeFormat = undefined
    // 2. Number format: relativeValue = 1, relativeFormat = "Hours ago"
    let relativeValue: string;
    if (value.relativeFormat) {
      relativeValue = `${value.relativeValue} ${value.relativeFormat}`;
    } else {
      relativeValue = String(value.relativeValue);
    }

    const parsedDate = chrono.parseDate(relativeValue);
    if (parsedDate) {
      return parsedDate;
    }
  }
  return new Date();
};

const getStringValue = (value: DateTimeRangeValue, locale: string = navigator.language): string => {
  if (value.type === 'absolute') {
    return value.absoluteValue
      ? value.absoluteValue.toLocaleString(locale, {
          year: 'numeric',
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit',
        })
      : 'None';
  } else if (value.type === 'relative') {
    if (value.relativeNow) {
      return locale === '' ? 'Now' : getDateTimeRelativeFormatLabel('Now', locale);
    }

    if (locale === '') {
      return `${value.relativeValue}${value.relativeFormat ? ' ' + value.relativeFormat : ''}`;
    } else {
      return `${value.relativeValue}${value.relativeFormat ? ' ' + getDateTimeRelativeFormatLabel(value.relativeFormat, locale) : ''}`;
    }
  }
  return '';
};

const getAbsoluteRangeValue = (value: Date): DateTimeRangeValue => {
  return {
    type: 'absolute',
    absoluteValue: value,
  };
};

export function getRelativeRangeValue(isNow: boolean): DateTimeRangeValue;
export function getRelativeRangeValue(value: string): DateTimeRangeValue;
export function getRelativeRangeValue(
  value: number,
  format: DateTimeRelativeFormat,
): DateTimeRangeValue;
export function getRelativeRangeValue(
  param: boolean | number | string,
  format?: DateTimeRelativeFormat,
): DateTimeRangeValue {
  if (typeof param === 'boolean') {
    return {
      type: 'relative',
      relativeNow: param,
    };
  } else if (typeof param === 'string') {
    return {
      type: 'relative',
      relativeValue: param,
    };
  } else {
    return {
      type: 'relative',
      relativeValue: param,
      relativeFormat: format,
    };
  }
}

interface DateTimeSelectorProps {
  value: DateTimeRangeValue;
  min?: Date;
  max?: Date;
  title: string;
  isError?: boolean;
  locale?: string;
  size?: 'default' | 'small';
  compareValue?: DateTimeRangeValue;
  onChange?: (value: DateTimeRangeValue) => void;
  rangeOptions?: string[];
}

const DateTimeSelector = ({
  value,
  min,
  max,
  onChange,
  isError,
  title,
  size = 'default',
  compareValue,
  locale = navigator.language,
  rangeOptions,
}: DateTimeSelectorProps) => {
  const checkRangeDate = rangeOptions ? getHighestPriority(rangeOptions) : null;

  const getAvailableFormats = (): DateTimeRelativeFormat[] => {
    if (!checkRangeDate) return dateTimeRelativeFormats;

    return dateTimeRelativeFormats.filter((format) => {
      const maxValue = getMaxRelativeValue(format, checkRangeDate);
      return maxValue > 0;
    });
  };

  const getInitialFormat = (): DateTimeRelativeFormat => {
    const availableFormats = getAvailableFormats();
    if (value.relativeFormat && availableFormats.includes(value.relativeFormat)) {
      return value.relativeFormat;
    }
    return availableFormats.length > 0 ? availableFormats[0] : 'Seconds ago';
  };

  const [relativeTimeFormat, setRelativeTimeFormat] =
    useState<DateTimeRelativeFormat>(getInitialFormat());
  const [relativeTimeValue, setRelativeTimeValue] = useState<string>(
    value.relativeValue !== undefined && value.relativeValue !== null
      ? String(value.relativeValue)
      : '1',
  );
  const [isNow, setIsNow] = useState<boolean>(value.relativeNow || false);
  const [isOpen, setIsOpen] = useState<boolean>(false);
  const [errorType, setErrorType] = useState<ValidationErrorType>(null);

  const [confirmedValues, setConfirmedValues] = useState({
    relativeTimeValue:
      value.relativeValue !== undefined && value.relativeValue !== null
        ? String(value.relativeValue)
        : '1',
    isNow: value.relativeNow || false,
    relativeTimeFormat: value.relativeFormat || 'Seconds ago',
  });
  const [shouldRestore, setShouldRestore] = useState<boolean>(true);

  useEffect(() => {
    if (value.type === 'relative') {
      setConfirmedValues({
        relativeTimeValue:
          value.relativeValue !== undefined && value.relativeValue !== null
            ? String(value.relativeValue)
            : '1',
        isNow: value.relativeNow || false,
        relativeTimeFormat: value.relativeFormat || 'Seconds ago',
      });
    }
  }, [value]);

  const validateDate = (
    currentValue: DateTimeRangeValue,
    compareValue?: DateTimeRangeValue,
  ): {isValid: boolean; errorType: ValidationErrorType} => {
    if (!compareValue) return {isValid: true, errorType: null};

    const currentParsed = getAbsoluteValue(currentValue);
    const compareParsed = getAbsoluteValue(compareValue);

    if (currentParsed !== undefined && compareParsed !== undefined) {
      if (title === 'Start Date') {
        if (currentParsed >= compareParsed) return {isValid: false, errorType: 'validation'};
      } else if (title === 'End Date') {
        if (currentParsed <= compareParsed) return {isValid: false, errorType: 'validation'};
      }

      // relative인 경우 range는 별도 검증
      if (currentValue.type === 'relative') {
        return {isValid: true, errorType: null};
      }

      if (checkRangeDate) {
        const maxRange = getMaxAllowedRange(checkRangeDate);
        const timeDiff = Math.abs(currentParsed.getTime() - compareParsed.getTime());

        if (timeDiff > maxRange) {
          return {isValid: false, errorType: 'range'};
        }
      }
    }
    return {isValid: true, errorType: null};
  };

  const getRangeErrorMessage = (errorType: ValidationErrorType): string => {
    return utilsGetRangeErrorMessage(errorType, title, checkRangeDate);
  };

  return (
    <Popover
      onOpenChange={(open) => {
        setIsOpen(open);

        if (open && value.type === 'relative') {
          setShouldRestore(true);

          if (value.relativeValue) {
            setRelativeTimeValue(value.relativeValue.toString());
          }
          if (value.relativeFormat && dateTimeRelativeFormats.includes(value.relativeFormat)) {
            setRelativeTimeFormat(value.relativeFormat);
          }
        } else if (!open && value.type === 'relative' && shouldRestore) {
          setRelativeTimeValue(confirmedValues.relativeTimeValue);
          setIsNow(confirmedValues.isNow);
          setRelativeTimeFormat(confirmedValues.relativeTimeFormat);
          setErrorType(null);
        }
      }}
      open={isOpen}
    >
      <PopoverTrigger asChild className="flex flex-1">
        <Button
          variant={'ghost'}
          className={cn(
            isError && 'text-destructive',
            'px-3 h-6 mx-0.75 rounded-sm',
            size === 'small' ? 'text-[13px]' : 'text-sm',
            isOpen && 'px-3 bg-accent',
          )}
        >
          {getStringValue(value, locale)}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="mr-9">
        <div className="text-sm text-muted-foreground font-medium pb-3">{title}</div>
        <Tabs
          defaultValue={value.type}
          onValueChange={(newValue) => {
            setErrorType(null);
            if (newValue === 'relative') {
              setRelativeTimeValue(confirmedValues.relativeTimeValue);
              setRelativeTimeFormat(confirmedValues.relativeTimeFormat);
              setIsNow(confirmedValues.isNow);
            }
          }}
        >
          <TabsList className={cn('grid w-full grid-cols-2')}>
            <TabsTrigger value={'absolute'}>Absolute</TabsTrigger>
            <TabsTrigger value={'relative'}>Relative</TabsTrigger>
          </TabsList>
          <TabsContent value={'absolute'} className="pb-0">
            <DateTimePanel
              min={min}
              max={max}
              value={getAbsoluteValue(value)}
              isError={isError}
              compareValue={compareValue ? getAbsoluteValue(compareValue) : undefined}
              title={title}
              validateDateRange={(currentValue: Date, compareValue?: Date) => {
                const currentRangeValue = getAbsoluteRangeValue(currentValue);
                const compareRangeValue = compareValue
                  ? getAbsoluteRangeValue(compareValue)
                  : undefined;
                return validateDate(currentRangeValue, compareRangeValue);
              }}
              getRangeErrorMessage={(errorType: ValidationErrorType) =>
                getRangeErrorMessage(errorType)
              }
              onChangeAction={(date, isSuccess) => {
                if (date !== undefined && isSuccess) {
                  setErrorType(null);
                  setConfirmedValues({
                    relativeTimeValue: '1',
                    relativeTimeFormat: 'Seconds ago',
                    isNow: false,
                  });
                  setRelativeTimeValue('1');
                  setRelativeTimeFormat('Seconds ago');
                  setIsNow(false);
                  onChange && onChange({type: 'absolute', absoluteValue: date});
                  setIsOpen(false);
                } else if (!isSuccess) {
                  setIsOpen(false);
                }
              }}
              // renderTrigger={renderTrigger}
            />
          </TabsContent>
          <TabsContent value={'relative'} className="flex flex-col gap-4 mt-4">
            <div className="flex flex-row gap-2">
              <Input
                type="text"
                inputMode="numeric"
                pattern="[0-9]*"
                className="w-40"
                disabled={isNow}
                value={relativeTimeValue}
                onChange={(e) => {
                  const val = e.target.value;
                  if (/^\d*$/.test(val)) {
                    setRelativeTimeValue(val);
                  }
                }}
                onWheel={(e) => {
                  e.currentTarget.blur();
                }}
                onKeyDown={(e) => {
                  if (
                    ['e', 'E', '+', '-', '.'].includes(e.key) ||
                    (e.key.length === 1 && /[ㄱ-ㅎㅏ-ㅣ가-힣]/.test(e.key))
                  ) {
                    e.preventDefault();
                  }
                }}
              />
              <Select
                disabled={isNow}
                defaultValue={relativeTimeFormat}
                value={relativeTimeFormat}
                onValueChange={(value) => {
                  const newFormat = value as DateTimeRelativeFormat;
                  setRelativeTimeFormat(newFormat);
                }}
              >
                <SelectTrigger className="text-sm text-left leading-tight">
                  <SelectValue>
                    {getDateTimeRelativeFormatLabel(relativeTimeFormat, navigator.language)}
                  </SelectValue>
                </SelectTrigger>
                <SelectContent>
                  {getAvailableFormats().map((format) => (
                    <SelectItem key={`select-relative-format-${format}`} value={format}>
                      {getDateTimeRelativeFormatLabel(format, navigator.language)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-4">
              <div className="flex flex-row gap-2 items-center">
                <Switch checked={isNow} onCheckedChange={(checked) => setIsNow(checked)} />
                <div className="flex flex-col text-muted-foreground">
                  <label className="font-medium text-sm">Now</label>
                  <span className="text-xs">
                    Set to current time
                    {/* TODO translate 현재 시간으로 설정 */}
                  </span>
                </div>
              </div>
              {(isError || errorType !== null) && (
                <div className="py-3 px-4 border border-destructive rounded-lg  max-w-full">
                  <span className="text-xs text-destructive leading-snug block">
                    {getRangeErrorMessage(errorType)}
                  </span>
                </div>
              )}
            </div>
            <div className="flex gap-2 justify-end">
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  setRelativeTimeValue(confirmedValues.relativeTimeValue);
                  setIsNow(confirmedValues.isNow);
                  setRelativeTimeFormat(confirmedValues.relativeTimeFormat);
                  setErrorType(null);
                  setShouldRestore(false);
                  setIsOpen(false);
                }}
              >
                Cancel
                {/* TODO translate {t('label.common.cancel')} */}
              </Button>
              <Button
                size="sm"
                onClick={() => {
                  if (!isNow && (!relativeTimeValue || relativeTimeValue.trim() === '')) {
                    setErrorType('empty');
                    return;
                  }

                  // range에 따른 상대 시간 값 체크
                  if (!isNow) {
                    const maxValue = getMaxRelativeValue(relativeTimeFormat, checkRangeDate);
                    const currentValue = parseInt(relativeTimeValue);

                    if (currentValue > maxValue) {
                      setErrorType('range');
                      return;
                    }
                  }

                  const newValue = {
                    type: 'relative',
                    relativeValue: relativeTimeValue,
                    relativeFormat: relativeTimeFormat,
                    relativeNow: isNow,
                  } as DateTimeRangeValue;

                  const validation = validateDate(newValue, compareValue);
                  setErrorType(validation.isValid ? null : validation.errorType);

                  if (validation.isValid) {
                    setErrorType(null);
                    setConfirmedValues({
                      relativeTimeValue,
                      relativeTimeFormat,
                      isNow,
                    });
                    onChange && onChange(newValue);
                    setShouldRestore(false);
                    setIsOpen(false);
                  }
                }}
              >
                Done
                {/* TODO translate {t('label.common.done')} */}
              </Button>
            </div>
          </TabsContent>
        </Tabs>
      </PopoverContent>
    </Popover>
  );
};

interface DatetimeRangeProps {
  startTime: DateTimeRangeValue;
  endTime: DateTimeRangeValue;
  locale?: string;
  className?: string;
  size?: 'default' | 'small';
  onChange: (startTime: DateTimeRangeValue, endTime: DateTimeRangeValue) => void;
  rangeStepOptions?: Record<string, string[]>;
}

const DatetimeRange = ({
  startTime,
  endTime,
  onChange,
  className,
  size = 'default',
  locale = navigator.language,
  rangeStepOptions = {},
}: DatetimeRangeProps) => {
  const rangeKeys = Object.keys(rangeStepOptions ?? {});

  return (
    <div
      className={cn(
        'flex w-full items-center rounded-lg bg-background border border-input focus-within:border-foreground transition-all',
        size === 'small' ? 'h-8 text-[13px]' : 'h-9 text-sm',
        className,
      )}
    >
      {/* 임시 주석 처리 */}
      {/* <Select
        value={''}
        onValueChange={(value) => {
          switch (value) {
            case 'Last 5 minutes':
              handleDateChange(
                getRelativeRangeValue(5, 'Minutes ago'),
                getRelativeRangeValue(true),
              );
              break;
            case 'Last 15 minutes':
              handleDateChange(
                getRelativeRangeValue(15, 'Minutes ago'),
                getRelativeRangeValue(true),
              );
              break;
            case 'Last 30 minutes':
              handleDateChange(
                getRelativeRangeValue(30, 'Minutes ago'),
                getRelativeRangeValue(true),
              );
              break;
            case 'Last 1 hour':
              handleDateChange(getRelativeRangeValue(1, 'Hours ago'), getRelativeRangeValue(true));
              break;
            case 'Last 3 hour':
              handleDateChange(getRelativeRangeValue(3, 'Hours ago'), getRelativeRangeValue(true));
              break;
            case 'Last 6 hours':
              handleDateChange(getRelativeRangeValue(6, 'Hours ago'), getRelativeRangeValue(true));
              break;
            case 'Last 12 hours':
              handleDateChange(getRelativeRangeValue(12, 'Hours ago'), getRelativeRangeValue(true));
              break;
            case 'Last 24 hours':
              handleDateChange(getRelativeRangeValue(24, 'Hours ago'), getRelativeRangeValue(true));
              break;
            case "Last 48 hours":
              handleDateChange(getRelativeRangeValue(48, "Hours ago"), getRelativeRangeValue(true));
              break;
            case 'Last 7 days':
              handleDateChange(getRelativeRangeValue(7, 'Days ago'), getRelativeRangeValue(true));
              break;
            case 'Last 14 days':
              handleDateChange(getRelativeRangeValue(14, 'Days ago'), getRelativeRangeValue(true));
              break;
          }
        }}
      >
        <SelectTrigger
          className={
            'felx flex-row border-none px-0 w-[30px] h-[30px] rounded-r-sm shadow-none focus:ring-0 hover:bg-muted'
          }
        >
          <div
            className="flex justify-center items-center gap-2 w-16 [&+svg]:hidden"
            title="Quick select"
          >
            <Clock3 className={cn('h-4 w-4')} />
          </div>
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="Last 5 minutes">{'Last 5 minutes'}</SelectItem>
          <SelectItem value="Last 15 minutes">{'Last 15 minutes'}</SelectItem>
          <SelectItem value="Last 30 minutes">{'Last 30 minutes'}</SelectItem>
          <SelectItem value="Last 1 hour">{'Last 1 hour'}</SelectItem>
          <SelectItem value="Last 3 hour">{'Last 3 hour'}</SelectItem>
          <SelectItem value="Last 6 hours">{'Last 6 hours'}</SelectItem>
          <SelectItem value="Last 12 hours">{'Last 12 hours'}</SelectItem>
          <SelectItem value="Last 24 hours">{'Last 24 hours'}</SelectItem>
          <SelectItem value="Last 48 hours">{"Last 48 hours"}</SelectItem>
          <SelectItem value="Last 7 days">{'Last 7 days'}</SelectItem>
          <SelectItem value="Last 14 days">{'Last 14 days'}</SelectItem> */}
      {/* TODO translate <SelectItem value="Last 5 minutes">{"마지막 5분"}</SelectItem>
          <SelectItem value="Last 15 minutes">{"마지막 15분"}</SelectItem>
          <SelectItem value="Last 30 minutes">{"마지막 30분"}</SelectItem>
          <SelectItem value="Last 1 hour">{"지난 1시간"}</SelectItem>
          <SelectItem value="Last 3 hour">{"지난 3시간"}</SelectItem>
          <SelectItem value="Last 6 hours">{"지난 6시간"}</SelectItem>
          <SelectItem value="Last 12 hours">{"지난 12시간"}</SelectItem>
          <SelectItem value="Last 24 hours">{"지난 24시간"}</SelectItem>
          <SelectItem value="Today">{"오늘"}</SelectItem>
          <SelectItem value="Yesterday">{"어제"}</SelectItem>
          <SelectItem value="Last 7 days">{"지난 7일"}</SelectItem>
          <SelectItem value="Last 14 days">{"지난 14일"}</SelectItem> */}
      {/* </SelectContent>
      </Select> */}
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <div className="flex items-center pl-2 pr-1">
              <CalendarDays size={16} className="text-foreground" />
            </div>
          </TooltipTrigger>
          <TooltipContent variant="icon" side="bottom" className="mt-0.5 font-normal">
            Date range picker
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
      <DateTimeSelector
        title="Start Date"
        value={startTime}
        size={size}
        compareValue={endTime}
        max={getAbsoluteValue(endTime)}
        onChange={(date) => {
          onChange(date, endTime);
        }}
        locale={locale}
        rangeOptions={rangeKeys}
      />
      <div>-</div>
      <DateTimeSelector
        title="End Date"
        value={endTime}
        size={size}
        compareValue={startTime}
        min={getAbsoluteValue(startTime)}
        onChange={(date) => {
          onChange(startTime, date);
        }}
        locale={locale}
        rangeOptions={rangeKeys}
      />
    </div>
  );
};

/**
 * Converts DateTimeRangeValue to URL string format (Grafana-compatible)
 * - Absolute time: Unix timestamp in milliseconds
 * - Relative time: "now", "now-5m", "now-1h", etc.
 * 
 * @example
 * dateTimeValueToUrlString({type: 'relative', relativeNow: true}) // "now"
 * dateTimeValueToUrlString({type: 'relative', relativeValue: 1, relativeFormat: 'Hours ago'}) // "now-1h"
 * dateTimeValueToUrlString({type: 'absolute', absoluteValue: new Date()}) // "1709654400000"
 */
export function dateTimeValueToUrlString(value: DateTimeRangeValue): string {
  if (value.type === 'absolute') {
    // Absolute time → Unix timestamp (milliseconds)
    const date = value.absoluteValue instanceof Date 
      ? value.absoluteValue 
      : value.absoluteValue 
        ? new Date(value.absoluteValue) 
        : new Date();
    return date.getTime().toString();
  } else if (value.type === 'relative') {
    // Relative time → Grafana format
    if (value.relativeNow) {
      return 'now';
    }
    
    // If relativeValue is already a string with format like "-5m", use it directly
    // BUT only if there's no relativeFormat (which needs conversion)
    if (typeof value.relativeValue === 'string' && !value.relativeFormat) {
      // If it starts with 'now', return as-is, otherwise prepend 'now'
      if (value.relativeValue.startsWith('now')) {
        return value.relativeValue;
      }
      // If it's just a relative offset like "-5m", prepend 'now'
      return `now${value.relativeValue}`;
    }
    
    // Convert relativeFormat to Grafana time unit
    const formatToUnit: Record<string, string> = {
      'Seconds ago': 's',
      'Minutes ago': 'm',
      'Hours ago': 'h',
      'Days ago': 'd',
      'Weeks ago': 'w',
      'Months ago': 'M',
      'Years ago': 'y',
    };
    
    const unit = value.relativeFormat ? formatToUnit[value.relativeFormat] : 'h';
    const amount = value.relativeValue || 1;
    
    return `now-${amount}${unit}`;
  }
  
  return 'now';
}

/**
 * Converts URL string to DateTimeRangeValue (Grafana-compatible)
 * - "now" → relative now
 * - "now-5m", "now-1h" → relative time
 * - Unix timestamp → absolute time
 * 
 * @example
 * urlStringToDateTimeValue("now") // {type: 'relative', relativeNow: true}
 * urlStringToDateTimeValue("now-1h") // {type: 'relative', relativeValue: "-1h"}
 * urlStringToDateTimeValue("1709654400000") // {type: 'absolute', absoluteValue: Date}
 */
export function urlStringToDateTimeValue(urlString: string): DateTimeRangeValue {
  // Case 1: "now"
  if (urlString === 'now') {
    return {
      type: 'relative',
      relativeNow: true,
    };
  }
  
  // Case 2: "now-5m", "now-1h", "now+5m" (Grafana format)
  if (urlString.startsWith('now')) {
    // Extract the relative part (e.g., "-5m", "+1h")
    const relativeMatch = urlString.match(/^now([+-]?)(\d+)([smhdwMy])$/);
    if (relativeMatch) {
      const sign = relativeMatch[1] || '-'; // default to '-' for 'ago'
      const amount = relativeMatch[2];
      const unit = relativeMatch[3];
      
      // Convert Grafana unit to relativeFormat
      const unitToFormat: Record<string, DateTimeRelativeFormat> = {
        's': 'Seconds ago',
        'm': 'Minutes ago',
        'h': 'Hours ago',
        'd': 'Days ago',
        'w': 'Weeks ago',
        'M': 'Months ago',
        'y': 'Years ago',
      };
      
      const relativeFormat = unitToFormat[unit];
      
      if (relativeFormat && sign === '-') {
        // "ago" format (most common)
        return {
          type: 'relative',
          relativeValue: amount,
          relativeFormat,
          relativeNow: false,
        };
      } else if (sign === '+') {
        // "from now" format (future) - store as raw string
        return {
          type: 'relative',
          relativeValue: `+${amount}${unit}`,
          relativeNow: false,
        };
      }
    }
    
    // Fallback: treat as "now"
    return {
      type: 'relative',
      relativeNow: true,
    };
  }
  
  // Case 3: Unix timestamp (milliseconds)
  const timestamp = parseInt(urlString, 10);
  if (!isNaN(timestamp)) {
    return {
      type: 'absolute',
      absoluteValue: new Date(timestamp),
    };
  }
  
  // Fallback: return "now"
  return {
    type: 'relative',
    relativeNow: true,
  };
}

export {
  DatetimeRange,
  getAbsoluteValue,
};
export type {DateTimeRangeValue};
