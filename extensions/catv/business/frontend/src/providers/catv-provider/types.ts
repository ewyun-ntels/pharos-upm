/**
 * CATV Provider Types
 *
 * Type definitions for CATV STB control data provider
 */

import type {BaseRecord} from '@lib/data-provider';

/**
 * Resource names for CATV provider
 */
export const CATV_RESOURCES = {
  SCHEDULE_RESULTS: 'schedule-results',
  SCHEDULE_RESULT_DETAILS: 'schedule-result-details',
  SCHEDULES: 'schedules',
  SOS: 'sos',
  L3S: 'l3s',
  CELLS: 'cells',
  SETTOPBOXES: 'settopboxes',
} as const;

/**
 * Hierarchy Item - represents SO/L3/Cell/Settopbox name
 */
export interface HierarchyItem extends BaseRecord {
  id: string; // item value (SO name, L3 ID, Cell ID, MAC address)
  name: string; // display name (same as id)
}

/**
 * Schedule Result - 스케줄 결과
 */
export interface ScheduleResultRecord extends BaseRecord {
  id: string;
  request_time: string;
  last_update_time: string;
  schedule_id: string;
  schedule_name: string;
  schedule_type: string;
  k8s_job_names: string[];
  work_type: string;
  count_by_status: {
    pending: number;
    running: number;
    succeeded: number;
    failed: number;
  };
}

/**
 * Schedule Result Detail - 개별 STB 제어 결과 상세
 */
export interface ScheduleResultDetailRecord extends BaseRecord {
  id: string;
  request_time: string;
  control_start_time: string | null;
  control_end_time: string | null;
  schedule_id: string;
  schedule_name: string;
  schedule_type: string;
  k8s_job_name: string;
  stb_mdl_nm: string;
  cm_mac_addr: string;
  cm_ip_addr: string;
  stb_mac_addr: string;
  src_ip_addr: string;
  work_type: string;
  work_value: string | null;
  progress_status: string;
  result_code: string;
  result_message: string;
}

/**
 * Job list API response
 */
/**
 * 결과 목록 API 응답 (/control/schedule/results)
 */
export interface ScheduleResultResponse {
  total: number;
  limit: number;
  offset: number;
  filters: {
    id?: string;
    work_type?: string;
    schedule_name?: string;
  };
  items: ScheduleResultRecord[];
}

/**
 * 결과 상세 API 응답 (/control/schedule/results/:id)
 */
export interface ScheduleResultDetailResponse {
  id: string;
  total: number;
  limit: number;
  offset: number;
  filters: {
    k8s_job_name?: string;
    cm_mac_addr?: string;
    stb_mac_addr?: string;
    progress_status?: string[];
    result_code?: string;
    search?: string;
  };
  items: ScheduleResultDetailRecord[];
}

/**
 * Schedule - 스케줄 정보 (생성/수정 API 및 목록 API 응답에 사용)
 */
export interface Schedule extends BaseRecord {
  id: string;
  name: string;
  schedule_type: 'immediately' | 'once' | 'repeat';
  schedule_spec_once?: string;
  schedule_spec_repeat?: string;
  sos?: string[];
  l3s?: string[];
  cells?: string[];
  settopboxes?: string[];
  area_type: string;
  area_ids: string[];
  work_type: string;
  work_value: string | null;
  update_time?: string;
  create_time?: string;
  is_deleted?: number;
}

/**
 * Schedule Result API 쿼리 파라미터
 */
export interface ScheduleResultParams {
  limit?: number;
  offset?: number;
  id?: string;
  schedule_id?: string;
  schedule_name?: string;
  work_type?: string;
  from?: string;
  to?: string;
}

/**
 *  Schedule Result Detail API 쿼리 파라미터
 */
export interface ScheduleResultDetailParams {
  limit?: number;
  offset?: number;
  k8s_job_name?: string;
  cm_mac_addr?: string;
  stb_mac_addr?: string;
  progress_status?: string | string[];
  result_code?: string;
  search?: string;
}
