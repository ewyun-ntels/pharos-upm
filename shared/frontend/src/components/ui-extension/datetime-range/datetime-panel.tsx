'use client';

// 기존 https://github.com/huybuidac/shadcn-datetime-picker 에서 Popoevr를 제거후 Panel만 남긴 코드
// DateTimeRange에서는 panel만 필요하기 때문에 해당 코드만 남김 (UI 및 기능 일부 수정됨)
import * as React from 'react';
import {useCallback, useEffect, useMemo, useState} from 'react';
import {format, getYear, setYear, addMonths, subMonths} from 'date-fns';
import {ChevronDownIcon, ChevronLeftIcon, ChevronRightIcon, ChevronUpIcon} from 'lucide-react';
import {DayPicker, Matcher, TZDate} from 'react-day-picker';

import {cn} from '../../../lib';
import {Button, buttonVariants} from '@pharos/shared/components/ui';
import {MonthYearPicker, TimePicker} from './datetime-util';
import {ValidationErrorType} from './Datetime';

export type CalendarProps = Omit<React.ComponentProps<typeof DayPicker>, 'mode'>;

export type DateTimePanelProps = {
  /**
   * The modality of the popover. When set to true, interaction with outside elements will be disabled and only popover content will be visible to screen readers.
   * If you want to use the datetime picker inside a dialog, you should set this to true.
   * @default false
   */
  modal?: boolean;
  /**
   * The datetime value to display and control.
   */
  value: Date | undefined;
  /**
   * Callback function to handle datetime changes.
   */
  onChangeAction: (date: Date | undefined, isSuccess?: boolean) => void;
  /**
   * The minimum datetime value allowed.
   * @default undefined
   */
  min?: Date;
  /**
   * The maximum datetime value allowed.
   */
  max?: Date;
  /**
   * The timezone to display the datetime in, based on the date-fns.
   * For a complete list of valid time zone identifiers, refer to:
   * https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
   * @default undefined
   */
  timezone?: string;
  /**
   * Whether the datetime picker is disabled.
   * @default false
   */
  disabled?: boolean;
  /**
   * Whether to show the time picker.
   * @default false
   */
  hideTime?: boolean;
  /**
   * Whether to use 12-hour format.
   * @default false
   */
  use12HourFormat?: boolean;
  /**
   * Whether to show the clear button.
   * @default false
   */
  clearable?: boolean;
  /**
   * Custom class names for the component.
   */
  classNames?: {
    /**
     * Custom class names for the trigger (the button that opens the picker).
     */
    trigger?: string;
  };
  timePicker?: {
    hour?: boolean;
    minute?: boolean;
    second?: boolean;
  };
  hideButton?: boolean;
  onSelect?: (selectedDate: Date | undefined) => void;
  isError?: boolean;
  compareValue?: Date;
  title?: string;
  validateDateRange?: (
    currentValue: Date,
    compareValue?: Date,
  ) => {isValid: boolean; errorType: ValidationErrorType};
  getRangeErrorMessage?: (errorType: ValidationErrorType, title?: string) => string;
};

export function DateTimePanel({
  value,
  onChangeAction,
  // renderTrigger,
  min,
  max,
  timezone,
  hideTime,
  hideButton,
  use12HourFormat,
  disabled,
  clearable,
  classNames,
  timePicker,
  modal = false,
  isError,
  compareValue,
  title,
  validateDateRange,
  getRangeErrorMessage,
  ...props
}: DateTimePanelProps & CalendarProps) {
  const [monthYearPicker, setMonthYearPicker] = useState<'month' | 'year' | false>(false);
  const initDate = useMemo(() => new TZDate(value || new Date(), timezone), [value, timezone]);

  const [month, setMonth] = useState<Date>(initDate);
  const [date, setDate] = useState<Date>(initDate);
  const [errorType, setErrorType] = useState<ValidationErrorType>(null);
  const [previousDate, setPreviousDate] = useState<Date>(initDate);

  const endMonth = useMemo(() => {
    return setYear(month, getYear(month) + 1);
  }, [month]);
  const minDate = useMemo(() => (min ? new TZDate(min, timezone) : undefined), [min, timezone]);
  const maxDate = useMemo(() => (max ? new TZDate(max, timezone) : undefined), [max, timezone]);

  const validateDate = useCallback(
    (
      currentValue: Date,
      compareValue?: Date,
    ): {isValid: boolean; errorType: ValidationErrorType} => {
      if (validateDateRange) {
        return validateDateRange(currentValue, compareValue);
      }

      // 기본 검증 로직 (fallback)
      if (!compareValue) return {isValid: true, errorType: null};

      if (title === 'Start Date') {
        if (currentValue >= compareValue) return {isValid: false, errorType: 'validation'};
      } else if (title === 'End Date') {
        if (currentValue <= compareValue) return {isValid: false, errorType: 'validation'};
      }

      return {isValid: true, errorType: null};
    },
    [validateDateRange, title],
  );

  const onDayChanged = useCallback(
    (d: Date) => {
      d.setHours(date.getHours(), date.getMinutes(), date.getSeconds());
      if (min && d < min) {
        d.setHours(min.getHours(), min.getMinutes(), min.getSeconds());
      }
      if (max && d > max) {
        d.setHours(max.getHours(), max.getMinutes(), max.getSeconds());
      }
      setDate(d);
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [setDate, setMonth],
  );

  const onSubmit = useCallback(() => {
    const validation = validateDate(date, compareValue);
    setErrorType(validation.isValid ? null : validation.errorType);

    if (validation.isValid) {
      onChangeAction(new Date(date), true);
      setPreviousDate(new Date(date));
    }
  }, [date, onChangeAction, compareValue, validateDate]);

  const onCancel = useCallback(() => {
    setDate(previousDate);
    setMonth(previousDate);
    setErrorType(null);
    onChangeAction(undefined, false);
  }, [previousDate, onChangeAction]);

  const onMonthYearChanged = useCallback(
    (d: Date, mode: 'month' | 'year') => {
      setMonth(d);
      if (mode === 'year') {
        setMonthYearPicker('month');
      } else {
        setMonthYearPicker(false);
      }
    },
    [setMonth, setMonthYearPicker],
  );
  const onNextMonth = useCallback(() => {
    setMonth(addMonths(month, 1));
  }, [month]);
  const onPrevMonth = useCallback(() => {
    setMonth(subMonths(month, 1));
  }, [month]);

  // useEffect(() => {
  //   if (open) {
  //     setDate(initDate);
  //     setMonth(initDate);
  //     setMonthYearPicker(false);
  //   }
  // }, [open, initDate]);
  useEffect(() => {
    setDate(initDate);
    setMonth(initDate);
    setMonthYearPicker(false);
    setPreviousDate(initDate);
  }, [initDate]);

  return (
    <div>
      <div className="flex items-center justify-between mt-4">
        <div className="text-md font-bold ms-2 flex gap-2 items-center cursor-pointer">
          <div>
            <span onClick={() => setMonthYearPicker(monthYearPicker === 'month' ? false : 'month')}>
              {format(month, 'MMMM')}
            </span>
            <span
              className="ms-1"
              onClick={() => setMonthYearPicker(monthYearPicker === 'year' ? false : 'year')}
            >
              {format(month, 'yyyy')}
            </span>
          </div>
          <Button
            variant="ghost"
            size="icon"
            className="w-7 h-7"
            onClick={() => setMonthYearPicker(monthYearPicker ? false : 'year')}
          >
            {monthYearPicker ? (
              <ChevronUpIcon className="w-4 h-4 stroke-muted-foreground" />
            ) : (
              <ChevronDownIcon className="w-4 h-4 stroke-muted-foreground" />
            )}
          </Button>
        </div>
        <div className={cn('flex space-x-2', monthYearPicker ? 'hidden' : '')}>
          <Button variant="outline" size="icon" className="w-7 h-7" onClick={onPrevMonth}>
            <ChevronLeftIcon className="w-4 h-4 stroke-muted-foreground" />
          </Button>
          <Button variant="outline" size="icon" className="w-7 h-7" onClick={onNextMonth}>
            <ChevronRightIcon className="w-4 h-4 stroke-muted-foreground" />
          </Button>
        </div>
      </div>
      <div className="relative overflow-hidden">
        <DayPicker
          timeZone={timezone}
          mode="single"
          selected={date}
          modifiers={{
            selected: (sel: Date) => {
              const newSel = new Date(sel);
              const newDate = new Date(date);
              return newSel.setHours(0, 0, 0, 0) === newDate.setHours(0, 0, 0, 0);
            },
            range_middle: (cur: Date) => {
              const newDate = new Date(date);
              const newCur = new Date(cur);
              if (max !== undefined && cur < max && newCur > newDate) {
                return true;
              } else if (min !== undefined && cur > min && newCur < newDate) {
                return true;
              }
              return false;
            },
          }}
          onSelect={(d) => d && onDayChanged(d)}
          month={month}
          endMonth={endMonth}
          disabled={
            [max ? {after: max} : null, min ? {before: min} : null].filter(Boolean) as Matcher[]
          }
          onMonthChange={setMonth}
          classNames={{
            dropdowns: 'flex w-full gap-2',
            months: 'flex w-full h-fit',
            month: 'flex flex-col w-full',
            month_caption: 'hidden',
            button_previous: 'hidden',
            button_next: 'hidden',
            month_grid: 'w-full border-collapse',
            weekdays: 'flex justify-between mt-2',
            weekday: 'text-muted-foreground rounded-md w-9 font-normal text-[0.8rem]',
            week: 'flex w-full justify-between mt-2',
            day: 'h-9 w-9 text-center text-sm p-0 relative flex items-center justify-center [&:has([aria-selected].day-range-end)]:rounded-r-md last:[&:has([aria-selected])]:bg-primary [&:has([aria-selected].day-outside)]:bg-primary first:[&:has([aria-selected])]:rounded-l-md last:[&:has([aria-selected])]:rounded-r-md focus-within:relative focus-within:z-20 rounded-1 [&[aria-selected]>button:hover]:bg-primary [&[aria-selected]>button:hover]:text-primary-foreground',
            day_button: cn(buttonVariants({variant: 'ghost'}), 'size-9 p-0 font-normal'),
            range_end: 'day-range-end bg-primary',
            selected:
              'bg-primary text-primary-foreground hover:bg-primary hover:text-primary-foreground focus:bg-primary focus:text-primary-foreground rounded-l-md rounded-r-md',
            today: 'bg-accent text-accent-foreground rounded-sm',
            outside:
              'day-outside text-muted-foreground/80 aria-selected:bg-primary aria-selected:text-primary-foreground',
            disabled: 'text-muted-foreground opacity-50',
            // TODO w-full이 맞는지 확인 필요
            range_middle:
              'w-full bg-secondary aria-selected:bg-primary aria-selected:text-primary-foreground',
            hidden: 'invisible',
          }}
          showOutsideDays={true}
          {...props}
        />
        <div
          className={cn(
            'absolute top-0 left-0 bottom-0 right-0',
            monthYearPicker ? 'bg-popover' : 'hidden',
          )}
        ></div>
        <MonthYearPicker
          value={month}
          mode={monthYearPicker as any}
          onChange={onMonthYearChanged}
          minDate={minDate}
          maxDate={maxDate}
          className={cn('absolute top-0 left-0 bottom-0 right-0', monthYearPicker ? '' : 'hidden')}
        />
      </div>
      <div className="flex flex-col gap-6 mt-4">
        {!hideTime && (
          <TimePicker
            timePicker={timePicker}
            value={date}
            onChange={setDate}
            use12HourFormat={use12HourFormat}
            min={minDate}
            max={maxDate}
          />
        )}
        {(isError || errorType !== null) && (
          <div className="py-3 px-4 border border-destructive rounded-lg  max-w-full">
            <span className="text-xs text-destructive leading-snug block">
              {getRangeErrorMessage?.(errorType, title) || 'Invalid date range.'}
            </span>
          </div>
        )}
        <div className="flex items-center justify-end">
          {timezone && (
            <div className="text-sm">
              <span>Timezone:</span>
              <span className="font-semibold ms-1">{timezone}</span>
            </div>
          )}
          {!hideTime && (
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={onCancel}>
                Cancel
                {/* TODO translate {t('label.common.cancel')} */}
              </Button>
              <Button size="sm" onClick={onSubmit}>
                Done
                {/* TODO translate {t('label.common.done')} */}
              </Button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
