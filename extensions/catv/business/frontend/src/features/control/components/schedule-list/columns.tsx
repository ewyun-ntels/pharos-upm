import {ColumnDef} from '@tanstack/react-table';
import {Badge, Tooltip, TooltipContent, TooltipTrigger} from '@pharos/shared/components/ui';
import {IconButton} from '@pharos/shared/components/ui-extension';
import {Pencil, Trash2} from 'lucide-react';
import {formatDateTime} from '@lib/format-date';
import type {Schedule} from '../../../../providers';
import {SCHEDULE_TYPE_LABELS, AREA_TYPE_LABELS, TimeFieldSize} from '../constants';

interface ScheduleListPermissions {
  canUpdate: boolean;
  canDelete: boolean;
}

export function getScheduleListColumns(
  onEdit: (schedule: Schedule) => void,
  onDelete: (name: string) => void,
  permissions: ScheduleListPermissions = {canUpdate: true, canDelete: true},
): ColumnDef<Schedule>[] {
  const {canUpdate, canDelete} = permissions;
  return [
    {
      id: 'name',
      accessorKey: 'name',
      header: '스케줄 이름',
      meta: {flex: 1},
    },
    {
      id: 'schedule_type',
      accessorKey: 'schedule_type',
      header: '실행 타입',
      size: 100,
      cell: ({getValue}) => {
        const type = getValue<string>();
        return <Badge variant="outline">{SCHEDULE_TYPE_LABELS[type] || type}</Badge>;
      },
    },
    {
      id: 'schedule_spec',
      header: '실행 스펙',
      size: TimeFieldSize,
      cell: ({row}) => {
        const {schedule_type, schedule_spec_once, schedule_spec_repeat} = row.original;
        if (schedule_type === 'immediately') {
          return <span className="text-muted-foreground">즉시</span>;
        } else if (schedule_type === 'once' && schedule_spec_once) {
          return <span className="font-mono text-xs">{formatDateTime(schedule_spec_once)}</span>;
        } else if (schedule_type === 'repeat' && schedule_spec_repeat) {
          return (
            <Tooltip>
              <TooltipTrigger className="cursor-default">
                <code className="text-xs">{schedule_spec_repeat}</code>
              </TooltipTrigger>
              <TooltipContent>
                <p>Cron 표현식</p>
              </TooltipContent>
            </Tooltip>
          );
        }
        return <span className="text-muted-foreground">-</span>;
      },
    },
    {
      id: 'work_type',
      accessorKey: 'work_type',
      header: '제어 명령',
      meta: {flex: 1},
    },
    {
      id: 'area_type',
      accessorKey: 'area_type',
      header: '대상 타입',
      size: 80,
      cell: ({getValue}) => {
        const type = getValue<string>();
        return <Badge variant="secondary">{AREA_TYPE_LABELS[type] || type}</Badge>;
      },
    },
    {
      id: 'create_time',
      accessorKey: 'create_time',
      header: '생성 시간',
      size: TimeFieldSize,
      cell: ({getValue}) => formatDateTime(getValue<string>()),
    },
    ...(canUpdate || canDelete
      ? [
          {
            id: 'actions',
            header: '액션',
            size: 100,
            cell: ({row}: {row: {original: Schedule}}) => {
              const {schedule_type, schedule_spec_once} = row.original;
              const isImmediately = schedule_type === 'immediately';
              const isExpiredOnce =
                schedule_type === 'once' &&
                !!schedule_spec_once &&
                new Date(schedule_spec_once) < new Date();
              const showEdit = canUpdate && !isImmediately && !isExpiredOnce;
              const showDelete = canDelete;
              if (!showEdit && !showDelete) return null;
              return (
                <div className="flex gap-1" onClick={(e) => e.stopPropagation()}>
                  {showEdit && (
                    <IconButton
                      variant="ghost"
                      size="icon-xs"
                      icon={<Pencil />}
                      onClick={(e) => {
                        e.stopPropagation();
                        onEdit(row.original);
                      }}
                    >
                      수정
                    </IconButton>
                  )}
                  {showDelete && (
                    <IconButton
                      variant="ghost"
                      size="icon-xs"
                      icon={<Trash2 />}
                      className="text-destructive hover:text-destructive"
                      onClick={(e) => {
                        e.stopPropagation();
                        if (confirm(`"${row.original.name}" 스케줄을 삭제하시겠습니까?`)) {
                          onDelete(row.original.name);
                        }
                      }}
                    >
                      삭제
                    </IconButton>
                  )}
                </div>
              );
            },
          } as ColumnDef<Schedule>,
        ]
      : []),
  ];
}
