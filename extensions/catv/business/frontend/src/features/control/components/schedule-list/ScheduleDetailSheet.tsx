'use client';

import {useMemo} from 'react';
import {X} from 'lucide-react';
import {useList} from '@lib/data-provider';
import {formatDateTime} from '@lib/format-date';
import {Badge, Button} from '@pharos/shared/components/ui';
import {useSheetStore} from '@components/sheet/sheetStore';
import {DataGrid} from '@pharos/shared/components/ui-extension/data-grid';
import {
  CATV_PROVIDER_NAME,
  CATV_RESOURCES,
  Schedule,
  ScheduleResultRecord,
} from '../../../../providers';
import {getScheduleResultColumns, ScheduleResultRow} from '../schedule-result/columns';
import {toScheduleResultRow} from '../schedule-result/utils';
import {SCHEDULE_TYPE_LABELS, AREA_TYPE_LABELS} from '../constants';

interface ScheduleDetailSheetProps {
  schedule: Schedule;
}

export function ScheduleDetailSheet({schedule}: ScheduleDetailSheetProps) {
  const {setClose} = useSheetStore();

  // 해당 스케줄로 실행된 스케줄 목록 조회
  const resultsQuery = useList<ScheduleResultRecord>({
    resource: CATV_RESOURCES.SCHEDULE_RESULTS,
    dataProviderName: CATV_PROVIDER_NAME,
    filters: [{field: 'schedule_id', operator: 'eq', value: schedule.id}],
    pagination: {
      currentPage: 1,
      pageSize: 100,
      mode: 'server' as const,
    },
  });

  const displayData: ScheduleResultRow[] = useMemo(() => {
    return ((resultsQuery.query.data?.data as ScheduleResultRecord[]) || []).map(
      toScheduleResultRow,
    );
  }, [resultsQuery.query.data]);

  const columns = getScheduleResultColumns();

  return (
    <div className="flex flex-col h-full w-full">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b shrink-0">
        <div className="flex items-center gap-2 min-w-0">
          <span className="font-semibold text-base truncate">{schedule.name}</span>
          <Badge variant="outline">
            {SCHEDULE_TYPE_LABELS[schedule.schedule_type] || schedule.schedule_type}
          </Badge>
        </div>
        <Button
          size="icon"
          variant="outline"
          className="h-7 w-7 shrink-0 ml-2"
          onClick={() => setClose()}
        >
          <X className="h-4 w-4" />
        </Button>
      </div>

      {/* Body */}
      <div className="flex-1 overflow-y-auto px-4 py-4">
        <div className="space-y-6">
          {/* 스케줄 정보 */}
          <section className="space-y-3">
            <h3 className="text-sm font-semibold">스케줄 정보</h3>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">스케줄 이름:</span>
                <span className="font-semibold">{schedule.name}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">실행 타입:</span>
                <Badge variant="outline">
                  {SCHEDULE_TYPE_LABELS[schedule.schedule_type] || schedule.schedule_type}
                </Badge>
              </div>
              {schedule.schedule_type === 'once' && schedule.schedule_spec_once && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">실행 스펙:</span>
                  <span className="font-mono text-xs">{schedule.schedule_spec_once}</span>
                </div>
              )}
              {schedule.schedule_type === 'repeat' && schedule.schedule_spec_repeat && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">실행 스펙:</span>
                  <span className="font-mono text-xs">{schedule.schedule_spec_repeat}</span>
                </div>
              )}
              <div className="flex justify-between">
                <span className="text-muted-foreground">제어 명령:</span>
                <span>{schedule.work_type}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">대상 타입:</span>
                <span>{AREA_TYPE_LABELS[schedule.area_type] || schedule.area_type}</span>
              </div>
              {schedule.area_ids && schedule.area_ids.length > 0 && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">대상 개수:</span>
                  <span>{schedule.area_ids.length}개</span>
                </div>
              )}
              {schedule.create_time && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">생성 시간:</span>
                  <span>{formatDateTime(schedule.create_time)}</span>
                </div>
              )}
              {schedule.update_time && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">수정 시간:</span>
                  <span>{formatDateTime(schedule.update_time)}</span>
                </div>
              )}
            </div>
          </section>

          {/* 스케줄 결과 목록 */}
          <section className="space-y-3">
            <div className="flex justify-between items-center">
              <h3 className="text-sm font-semibold">스케줄 결과 목록</h3>
              {resultsQuery.query.data?.total !== undefined && (
                <span className="text-xs text-muted-foreground">
                  전체 {resultsQuery.query.data.total}개
                </span>
              )}
            </div>
            <div className="border rounded-md">
              <DataGrid
                tableKey="schedule-result-list"
                data={displayData}
                columns={columns}
                rowCursor={false}
                getRowId={(row: ScheduleResultRow) => row.result_id}
                isLoading={resultsQuery.query.isLoading}
                enablePagination={false}
              />
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}
