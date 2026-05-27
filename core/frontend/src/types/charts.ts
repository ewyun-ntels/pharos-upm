import type { ChartQuery as BaseChartQuery, Kind } from '@pharos/shared/types/dashboard'

// ============================================================================
// 차트 플러그인 레지스트리 패턴용 타입
// ============================================================================
// 
// 아키텍처 원칙:
// - ChartOptions (shared): DB 저장용 JSON 직렬화 타입
// - 각 차트의 PropsSchema: 런타임 실행용 (함수 포함 가능)
// - 차트는 플러그인으로 등록되어 동적으로 로드됨
//
// ============================================================================

// ============================================================================
// Query Types - 차트 데이터 로딩용
// ============================================================================

/**
 * Chart query arguments (시간 범위, step, 필터 값 등)
 * Map 사용 이유:
 * - 필터 스토어와 일관성
 * - O(1) 접근 성능
 * - 메모리 효율적
 */
type ChartQueryArgs = Map<string, string | number | (() => number)>;

/**
 * Chart Query Request
 * 
 * dashboardProvider가 실제로 받는 타입:
 * 
 * 1. Panel Mode (viewer 권한):
 *    - id + kind: 저장된 패널 실행
 *    - query/datasourceName 불필요
 * 
 * 2. Run Mode (editor/owner 권한):
 *    - query + datasourceName: 쿼리 직접 실행
 *    - id/kind 불필요
 * 
 * 3. 공통:
 *    - dashboardId: 필수 (API 경로)
 *    - args: 변수 치환용 (선택)
 */
type ChartQueryRequest = {
  // 필수: 모든 모드에서 필요
  dashboardId: string;
  
  // Panel Mode 필드 (viewer용)
  id?: string;
  kind?: Kind;
  
  // Run Mode 필드 (editor/owner용)
  query?: string;
  datasourceName?: string;
  datasourceType?: string;
  templateStyle?: string;
  
  // 공통 선택 필드
  label?: string;
  queryName?: string;
  forceRefresh?: number;
  args?: ChartQueryArgs;
};

/**
 * Chart Query (DB 저장용)
 * BaseChartQuery를 확장하여 로컬 런타임 필드 추가
 */
type ChartQuery = Partial<BaseChartQuery> & {
  dashboardId?: string;
  id?: string;
  kind?: Kind;
  queryName?: string;
  query: string; // 최소한 query는 필수
};

// ============================================================================
// Response Types - 차트 데이터 응답용 (Re-exported from shared library)
// ============================================================================

// Chart metric label (범례용 메타데이터)
export type {ChartMetricLabel} from '@pharos/shared/components/charts/utils/chartType';

// Chart metric & data types (시계열 데이터)
export type {ChartMetric, ChartMetricData} from '@pharos/shared/components/charts';

// ============================================================================
// Exports
// ============================================================================

export type {
  // Query types
  ChartQuery,
  ChartQueryArgs,
  ChartQueryRequest,
  // Response types are re-exported above:
  // - ChartMetric (from shared)
  // - ChartMetricData (from shared)
  // - ChartMetricLabel (from shared)
};
