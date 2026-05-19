import {Badge} from '@pharos/shared/components/ui';
import {ColumnDef} from '@tanstack/react-table';
import {formatDateTime} from '@lib/format-date';
import type {ScheduleResultDetailRecord} from '../../../../providers';
import {calcElapsed} from '../schedule-result/utils';
import {
  SCHEDULE_TYPE_LABELS,
  TimeFieldSize,
  STATUS_PENDING,
  STATUS_RUNNING,
  STATUS_SUCCEEDED,
  STATUS_FAILED,
} from '../constants';

export type StbInfoRow = ScheduleResultDetailRecord & {parsedInfo: Record<string, string>};

export type ProgressStatus =
  | typeof STATUS_PENDING
  | typeof STATUS_RUNNING
  | typeof STATUS_SUCCEEDED
  | typeof STATUS_FAILED;

export const PROGRESS_STATUS_CONFIG: Record<
  ProgressStatus,
  {label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline'; className?: string}
> = {
  pending: {label: '대기', variant: 'outline', className: 'text-muted-foreground'},
  running: {label: '진행 중', variant: 'default'},
  succeeded: {label: '성공', variant: 'outline', className: 'text-green-600 border-green-600'},
  failed: {label: '실패', variant: 'destructive'},
};

export function ProgressStatusBadge({status}: {status: string}) {
  const config = PROGRESS_STATUS_CONFIG[status as ProgressStatus] ?? {
    label: status,
    variant: 'outline' as const,
  };
  return (
    <Badge variant={config.variant} className={config.className}>
      {config.label}
    </Badge>
  );
}

export const BASIC_COLUMN_DEFS: ColumnDef<ScheduleResultDetailRecord>[] = [
  // 신원 컬럼
  {
    id: 'control_start_time',
    accessorKey: 'control_start_time',
    header: '시작 시간',
    size: TimeFieldSize,
    enableHiding: true,
    cell: ({getValue}) => formatDateTime(getValue<string>()),
  },
  {
    id: 'control_end_time',
    accessorKey: 'control_end_time',
    header: '종료 시간',
    size: TimeFieldSize,
    enableHiding: true,
    cell: ({getValue}) => formatDateTime(getValue<string>()),
  },
  {
    id: 'elapsed',
    header: '경과 시간',
    size: 120,
    enableHiding: true,
    cell: ({row}) => calcElapsed(row.original.control_start_time, row.original.control_end_time),
  },
  {
    id: 'stb_mdl_nm',
    accessorKey: 'stb_mdl_nm',
    header: '모델명',
    size: 120,
    enableHiding: true,
    cell: ({getValue}) => getValue<string>() || '-',
  },
  {
    id: 'cm_mac_addr',
    accessorKey: 'cm_mac_addr',
    header: 'CM MAC',
    size: 160,
    enableHiding: true,
    cell: ({getValue}) => getValue<string>() || '-',
  },
  {
    id: 'stb_mac_addr',
    accessorKey: 'stb_mac_addr',
    header: 'STB MAC',
    size: 160,
    enableHiding: true,
    cell: ({getValue}) => getValue<string>() || '-',
  },
  {
    id: 'progress_status',
    accessorKey: 'progress_status',
    header: '진행 상태',
    size: 100,
    enableHiding: true,
    cell: ({getValue}) => {
      const status = getValue<string>();
      return status ? <ProgressStatusBadge status={status} /> : '-';
    },
  },
  // 결과 컬럼
  {
    id: 'result_code',
    accessorKey: 'result_code',
    header: '결과 코드',
    size: 100,
    enableHiding: true,
    cell: ({getValue}) => {
      const code = getValue<string>();
      if (!code) return '-';
      return <span className="font-mono">{code}</span>;
    },
  },
  {
    id: 'result_message',
    accessorKey: 'result_message',
    header: '결과 메시지',
    size: 300,
    enableHiding: true,
    cell: ({getValue}) => {
      const value = getValue<string>();
      if (!value) return '-';
      return (
        <div className="max-w-[300px] truncate" title={value}>
          <span className="font-mono text-xs">{value}</span>
        </div>
      );
    },
  },
  // 메타 컬럼
  {
    id: 'schedule_name',
    accessorKey: 'schedule_name',
    header: '스케줄 이름',
    size: 150,
    enableHiding: true,
    cell: ({getValue}) => getValue<string>() || '-',
  },
  {
    id: 'schedule_type',
    accessorKey: 'schedule_type',
    header: '스케줄 타입',
    size: 120,
    enableHiding: true,
    cell: ({getValue}) => {
      const type = getValue<string>();
      if (!type) return '-';
      return SCHEDULE_TYPE_LABELS[type] || type;
    },
  },
  {
    id: 'k8s_job_name',
    accessorKey: 'k8s_job_name',
    header: 'K8s Job',
    size: 220,
    enableHiding: true,
    cell: ({getValue}) => <span className="font-mono text-xs">{getValue<string>() || '-'}</span>,
  },
];

export const BASIC_COLUMN_IDS = BASIC_COLUMN_DEFS.map((c) => c.id as string);

export const DEFAULT_HIDDEN_BASIC_COLUMN_IDS = [
  'control_end_time',
  'elapsed',
  'schedule_name',
  'schedule_type',
  'k8s_job_name',
];

// 워크타입별 DEFAULT_HIDDEN 오버라이드
export const DEFAULT_HIDDEN_BASIC_COLUMN_IDS_BY_WORK_TYPE: Record<string, string[]> = {
  stb_request_info: [
    'control_end_time',
    'elapsed',
    'result_code',
    'result_message',
    'schedule_name',
    'schedule_type',
    'k8s_job_name',
  ],
};
