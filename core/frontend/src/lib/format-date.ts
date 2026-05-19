import {DateTime} from 'luxon';

const DATE_TIME_OPTIONS: Intl.DateTimeFormatOptions = {
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
  second: '2-digit',
};

const DATE_TIME_SHORT_OPTIONS: Intl.DateTimeFormatOptions = {
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
};

/** yyyy. mm. dd. hh:mm:ss */
export function formatDateTime(date: Date | string | number | null | undefined): string {
  if (!date) return '-';
  return new Intl.DateTimeFormat(undefined, DATE_TIME_OPTIONS).format(new Date(date));
}

/** yyyy. mm. dd. hh:mm (seconds omitted) */
export function formatDateTimeShort(date: Date | string | number | null | undefined): string {
  if (!date) return '-';
  return new Intl.DateTimeFormat(undefined, DATE_TIME_SHORT_OPTIONS).format(new Date(date));
}

/** 상대 시간 ("3일 전", "2 hours ago" 등) */
export function formatRelativeTime(date: Date | string | number | null | undefined): string {
  if (!date) return '';
  const dt = DateTime.fromJSDate(new Date(date)).setLocale(navigator.language);
  return dt.toRelative() ?? '';
}
