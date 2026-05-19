import type {ScheduleResultRecord} from '../../../../providers';
import type {ScheduleResultRow} from './columns';
import {STATUS_SUCCEEDED, STATUS_RUNNING, STATUS_PENDING} from '../constants';

const MS_PER_SEC = 1000;
const SECS_PER_MIN = 60;
const MINS_PER_HOUR = 60;

/**
 * ScheduleResult API 응답 → 테이블 행 형식으로 변환 (공통 유틸)
 * ScheduleResult.tsx와 ScheduleDetailSheet.tsx에서 공유
 */
export function toScheduleResultRow(item: ScheduleResultRecord): ScheduleResultRow {
  const counts = item.count_by_status;
  const pending = counts.pending || 0;
  const running = counts.running || 0;
  const succeeded = counts.succeeded || 0;
  const failed = counts.failed || 0;
  const total = pending + running + succeeded + failed;

  // 모든 셋탑박스가 완료(성공 또는 실패) → 완료
  // running이 있으면 → 진행 중
  // 모두 pending이면 → 대기
  // 일부 완료 + 일부 pending (엣지케이스) → 진행 중
  let status: string;
  if (pending === 0 && running === 0) {
    status = STATUS_SUCCEEDED;
  } else if (running > 0) {
    status = STATUS_RUNNING;
  } else if (succeeded === 0 && failed === 0) {
    status = STATUS_PENDING;
  } else {
    status = STATUS_RUNNING;
  }

  return {
    result_id: item.id,
    work_type: item.work_type,
    status,
    request_time: item.request_time,
    update_time: item.last_update_time,
    schedule_name: item.schedule_name,
    schedule_type: item.schedule_type,
    total_count: total,
    succeeded_count: succeeded,
    failed_count: failed,
  };
}

export function calcElapsed(start: string | null, end: string | null): string {
  if (!start || !end) return '-';
  const diff = new Date(end).getTime() - new Date(start).getTime();
  if (diff < 0) return '-';
  const secs = Math.floor(diff / MS_PER_SEC);
  if (secs < SECS_PER_MIN) return `${secs}초`;
  const mins = Math.floor(secs / SECS_PER_MIN);
  const remSecs = secs % SECS_PER_MIN;
  if (mins < MINS_PER_HOUR) return `${mins}분${remSecs > 0 ? ` ${remSecs}초` : ''}`;
  const hours = Math.floor(mins / MINS_PER_HOUR);
  const remMins = mins % MINS_PER_HOUR;
  return `${hours}시간${remMins > 0 ? ` ${remMins}분` : ''}`;
}
