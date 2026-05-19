'use client';

import {X} from 'lucide-react';
import {Button} from '@pharos/shared/components/ui';
import {formatDateTime} from '@lib/format-date';
import {useSheetStore} from '@components/sheet/sheetStore';
import type {ScheduleResultDetailRecord} from '../../../../providers';
import {calcElapsed} from '../schedule-result/utils';
import {ProgressStatusBadge} from './columns';

interface ScheduleResultDetailSheetProps {
  record: ScheduleResultDetailRecord;
}

export function ScheduleResultDetailSheet({record}: ScheduleResultDetailSheetProps) {
  const {setClose} = useSheetStore();

  // 결과 메시지 JSON 파싱 시도
  const parsedResultMessage = (() => {
    if (!record.result_message || record.result_code !== '1') {
      return null;
    }
    try {
      return JSON.parse(record.result_message);
    } catch {
      return null;
    }
  })();

  return (
    <div className="flex flex-col h-full w-full">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b shrink-0">
        <div className="flex items-center gap-2 min-w-0">
          <span className="font-semibold text-base truncate">제어 레코드</span>
          <ProgressStatusBadge status={record.progress_status} />
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
          {/* 기본 정보 */}
          <section className="space-y-3">
            <h3 className="text-sm font-semibold">기본 정보</h3>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">K8s Job:</span>
                <span className="font-mono">{record.k8s_job_name}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">스케줄 ID:</span>
                <span className="font-mono">{record.schedule_id}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">스케줄 이름:</span>
                <span>{record.schedule_name || '-'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Work Type:</span>
                <span>{record.work_type}</span>
              </div>
            </div>
          </section>

          {/* 상태 정보 */}
          <section className="space-y-3">
            <h3 className="text-sm font-semibold">상태 정보</h3>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground">진행 상태:</span>
                <ProgressStatusBadge status={record.progress_status} />
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">결과 코드:</span>
                <span>{record.result_code || '-'}</span>
              </div>
            </div>
          </section>

          {/* 대상 정보 */}
          <section className="space-y-3">
            <h3 className="text-sm font-semibold">대상 정보</h3>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">모델명:</span>
                <span className="font-mono">{record.stb_mdl_nm}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">CM MAC:</span>
                <span className="font-mono">{record.cm_mac_addr}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">STB MAC:</span>
                <span className="font-mono">{record.stb_mac_addr}</span>
              </div>
            </div>
          </section>

          {/* 시간 정보 */}
          <section className="space-y-3">
            <h3 className="text-sm font-semibold">시간 정보</h3>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">시작 시간:</span>
                <span>{formatDateTime(record.request_time)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">제어 시작:</span>
                <span>{formatDateTime(record.control_start_time)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">제어 종료:</span>
                <span>{formatDateTime(record.control_end_time)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">경과 시간:</span>
                <span>{calcElapsed(record.control_start_time, record.control_end_time)}</span>
              </div>
            </div>
          </section>

          {/* 결과 메시지 */}
          {record.result_message && (
            <section className="space-y-3">
              <h3 className="text-sm font-semibold">결과 메시지</h3>
              {parsedResultMessage ? (
                <div className="p-3 bg-muted rounded-md text-sm">
                  <div className="grid grid-cols-2 gap-3">
                    {Object.entries(parsedResultMessage).map(([key, value]) => (
                      <div key={key} className="space-y-1 min-w-0">
                        <p className="text-xs text-muted-foreground font-medium">{key}</p>
                        <p className="font-mono text-xs break-all">{String(value)}</p>
                      </div>
                    ))}
                  </div>
                </div>
              ) : (
                <div className="p-3 bg-muted rounded-md text-sm break-all font-mono">
                  {record.result_message}
                </div>
              )}
            </section>
          )}
        </div>
      </div>
    </div>
  );
}
