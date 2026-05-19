import { z } from 'zod';
import { AnnotationSchema, type Annotation } from '../../../types/dashboard';
export type { Annotation } from '../../../types/dashboard';

// Zod schema for threshold
const ThresholdSchema = z.object({
  value: z.string().optional(),
  color: z.string().optional(),
  type: z.enum(['solid', 'dash', 'dot']).optional(),
  lineWidth: z.number().optional(),
});

// Zod schema for scales
const ScalesSchema = z.object({
  yMin: z.string().optional(),
  yMax: z.string().optional(),
});

// Zod schema for TimeSeriesPanelOptions
export const TimeSeriesPanelOptionsSchema = z.object({
  // Legend settings
  legendRule: z.string().optional(),
  legendEnabled: z.boolean().optional(),
  legendAlign: z.enum(['right', 'bottom']).optional(),
  legendAsTable: z.boolean().optional(),
  // Metrics display
  showAverage: z.boolean().optional(),
  showLast: z.boolean().optional(),
  showMax: z.boolean().optional(),
  showMin: z.boolean().optional(),
  showTotal: z.boolean().optional(),
  // Chart settings
  chartType: z.enum(['line', 'points', 'bars', 'area']).optional(),
  showPoints: z.boolean().optional(),
  lineWidth: z.number().optional(),
  pointSize: z.number().optional(),
  barWidth: z.number().optional(),
  barStack: z.boolean().optional(),
  decimals: z.number().optional(),
  chartColor: z.string().optional(),
  fillOpacity: z.number().optional(),
  unit: z.string().optional(),
  scales: ScalesSchema.optional(),
  syncId: z.string().optional(),
  thresholds: z.array(ThresholdSchema).optional(),
  annotations: z.array(AnnotationSchema).optional(),
  chartDataNotExistMessage: z.string().optional(),
  showLink: z.boolean().optional(),
  tooltipLimit: z.number().optional(),
  showAnnotationButton: z.boolean().optional(),
  annotationLabelPosition: z.enum(['top', 'bottom']).optional(),
});

// TypeScript type (inferred from Zod schema)
export type TimeSeriesPanelOptions = z.infer<typeof TimeSeriesPanelOptionsSchema>;

// Additional types
export type Threshold = z.infer<typeof ThresholdSchema>;
export type Scales = z.infer<typeof ScalesSchema>;

// ============================================================================
// Chart Data Types - External data injection
// ============================================================================

/**
 * Chart Metric - 시계열 데이터 포인트
 * 
 * timestamp는 밀리초 단위 Unix timestamp
 */
export interface ChartMetric {
  timestamp: number;  // Unix timestamp in milliseconds
  [key: string]: string | number | null;
}

/**
 * Chart Metric Data - 차트에 표시할 순수 데이터
 * TimeSeries는 이 데이터만 받아서 렌더링합니다.
 * 
 * **타임스탬프 단위: 모든 타임스탬프는 밀리초(milliseconds)**
 * - startTime: Unix timestamp (ms)
 * - endTime: Unix timestamp (ms)
 * - step: 데이터 포인트 간격 (ms)
 * - ChartMetric.timestamp: Unix timestamp (ms)
 * 
 * @see /docs/TIMESTAMP_DESIGN.md
 */
export interface ChartMetricData {
  chartMetric: ChartMetric[];  // 시계열 데이터 포인트
  startTime: number;           // Unix timestamp in milliseconds
  endTime: number;             // Unix timestamp in milliseconds
  step: number;                // 데이터 간격 (milliseconds)
  chartType: string;           // 차트 타입 (timeseries, line, area 등)
  uniqueKeys?: string[];       // 고유 메트릭 키 목록
  chartMetricMap?: Map<number, ChartMetric>; // 선택: 빠른 조회용 Map
}

/**
 * TimeSeriesChart Props - 순수 차트 컴포넌트
 * 데이터를 외부에서 주입받아 렌더링만 수행합니다.
 */
export interface TimeSeriesChartProps {
  // 필수: 차트 데이터 (외부 주입)
  data?: ChartMetricData;

  // 필수: 로딩 상태
  loading?: boolean;

  // 차트 옵션
  options?: TimeSeriesPanelOptions;

  // 차트 크기
  chartWidth?: number;
  chartHeight?: number;

  // Callbacks
  onTimeRangeChange?: (start: number, end: number) => void;

  // Utility functions (optional with defaults)
  formatLocalTime?: (date: Date) => string;
  convertUnit?: (value: number, unit?: string, decimals?: number) => string;
  translate?: (key: string) => string;

  // Panel-scoped annotations (from panel.options.annotations)
  annotations?: Annotation[];

  // Dashboard-scoped annotations (from DashboardConfig.annotations)
  globalAnnotations?: Annotation[];

  // Optional: Component ID for scroll position tracking
  id?: string;

  // Called when user clicks '+ Add Annotation' button inside the tooltip
  onChartClick?: (timeMs: number) => void;

  // Called when user clicks the delete button on an annotation in the tooltip
  onDeleteAnnotation?: (id: string) => void;
}
