export type RefreshIntervalType =
  | 'Off'
  | '1s'
  | '5s'
  | '10s'
  | '30s'
  | '1m'
  | '5m'
  | '15m'
  | '30m'
  | '1h'
  | '2h'
  | '1d';

export const REFRESH_INTERVAL: RefreshIntervalType[] = [
  'Off',
  '1s',
  '5s',
  '10s',
  '30s',
  '1m',
  '5m',
  '15m',
  '30m',
  '1h',
  '2h',
  '1d',
];

export const REFRESH_INTERVAL_DEFAULT: RefreshIntervalType = '1m';

export interface RefreshFilterOptions {
  /**
   * Initial refresh interval value
   * @default '1m'
   */
  initValue?: RefreshIntervalType;

  /**
   * Available refresh interval options
   * If not specified, all default options are available
   * @example ['Off', '30s', '1m', '5m', '1h']
   */
  refreshOptions?: RefreshIntervalType[];
}
