import {ColumnDef} from '@tanstack/react-table';
import {Badge, Tooltip, TooltipContent, TooltipTrigger} from '@pharos/shared/components/ui';
import {formatDateTime} from '@lib/format-date';
import {calcElapsed} from './utils';
import {
  SCHEDULE_TYPE_LABELS,
  TimeFieldSize,
  STATUS_PENDING,
  STATUS_RUNNING,
  STATUS_SUCCEEDED,
} from '../constants';

export interface ScheduleResultRow {
  result_id: string; // ScheduleResult.id 에 매핑
  work_type: string;
  status: string;
  request_time: string;
  update_time: string;
  schedule_name?: string;
  schedule_type?: string;
  total_count: number;
  succeeded_count: number;
  failed_count: number;
}

type ScheduleStatus = typeof STATUS_PENDING | typeof STATUS_RUNNING | typeof STATUS_SUCCEEDED;

const STATUS_CONFIG: Record<
  ScheduleStatus,
  {label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline'; className?: string}
> = {
  pending: {label: '대기', variant: 'outline', className: 'text-muted-foreground'},
  running: {label: '진행 중', variant: 'default'},
  succeeded: {label: '완료', variant: 'outline', className: 'text-green-600 border-green-600'},
};

function ScheduleStatusBadge({status}: {status: string}) {
  const config = STATUS_CONFIG[status as ScheduleStatus] ?? {
    label: status,
    variant: 'outline' as const,
  };
  return (
    <Badge variant={config.variant} className={config.className}>
      {config.label}
    </Badge>
  );
}

function ProgressBadges({
  succeeded,
  failed,
  total,
}: {
  succeeded: number;
  failed: number;
  total: number;
}) {
  return (
    <div className="flex gap-1 flex-wrap">
      <Badge variant="outline" className="text-green-600 border-green-600">
        완료 {succeeded}
      </Badge>
      {failed > 0 && <Badge variant="destructive">실패 {failed}</Badge>}
      <Badge variant="secondary">전체 {total}</Badge>
    </div>
  );
}

export function getScheduleResultColumns(): ColumnDef<ScheduleResultRow>[] {
  return [
    {
      id: 'request_time',
      accessorKey: 'request_time',
      header: '시작 시간',
      size: TimeFieldSize,
      cell: ({row}) => {
        const {request_time, result_id} = row.original;
        return (
          <Tooltip>
            <TooltipTrigger className="cursor-default text-left">
              {formatDateTime(request_time)}
            </TooltipTrigger>
            <TooltipContent>
              <p className="font-mono text-xs">ID: {result_id}</p>
            </TooltipContent>
          </Tooltip>
        );
      },
    },
    {
      id: 'schedule_name',
      accessorKey: 'schedule_name',
      header: '스케줄 이름',
      size: 150,
      cell: ({getValue}) => {
        const name = getValue<string | undefined>();
        return name ? (
          <span className="font-mono text-xs">{name}</span>
        ) : (
          <span className="text-muted-foreground">-</span>
        );
      },
    },
    {
      id: 'schedule_type',
      accessorKey: 'schedule_type',
      header: '스케줄 타입',
      size: 120,
      cell: ({getValue}) => {
        const type = getValue<string | undefined>();
        if (!type) return <span className="text-muted-foreground">-</span>;
        return <Badge variant="outline">{SCHEDULE_TYPE_LABELS[type] || type}</Badge>;
      },
    },
    {
      id: 'work_type',
      accessorKey: 'work_type',
      header: '제어 명령',
      meta: {flex: 1},
    },
    {
      id: 'elapsed',
      header: '경과 시간',
      size: 120,
      cell: ({row}) => {
        const {request_time, update_time} = row.original;
        return calcElapsed(request_time, update_time);
      },
    },
    {
      id: 'status',
      accessorKey: 'status',
      header: '상태',
      size: 100,
      cell: ({getValue}) => <ScheduleStatusBadge status={getValue<string>()} />,
    },
    {
      id: 'progress',
      header: '진행 현황',
      meta: {flex: 0.5},
      minSize: 250,
      cell: ({row}) => {
        const {total_count, succeeded_count, failed_count} = row.original;
        return (
          <ProgressBadges succeeded={succeeded_count} failed={failed_count} total={total_count} />
        );
      },
    },
  ];
}
