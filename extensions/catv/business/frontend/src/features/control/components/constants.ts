export const WORK_TYPE_OPTIONS = [
  {label: 'stb_request_info (전송 요청)', value: 'stb_request_info'},
  {label: 'smartReboot (스마트 리부팅)', value: 'smartReboot'},
] as const;

export const SCHEDULE_TYPE_LABELS: Record<string, string> = {
  immediately: '즉시 실행',
  once: '한번 실행',
  repeat: '반복 실행',
};

export const AREA_TYPE_LABELS: Record<string, string> = {
  so: 'SO',
  l3: 'L3',
  cell: 'Cell',
  settopbox: 'STB',
  cm: 'CM',
  stb: 'STB',
  all: '전체',
};

export const POLL_INTERVAL_MS = 5000;
export const SEARCH_DEBOUNCE_MS = 300;

export const TimeFieldSize = 210;

export const STATUS_SUCCEEDED = 'succeeded';
export const STATUS_RUNNING = 'running';
export const STATUS_PENDING = 'pending';
export const STATUS_FAILED = 'failed';
