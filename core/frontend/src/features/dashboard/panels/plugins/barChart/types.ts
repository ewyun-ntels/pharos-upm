import {z} from 'zod';

// Zod schema for threshold
const ThresholdSchema = z.object({
  value: z.string().optional(),
  color: z.string().optional(),
  type: z.enum(['solid', 'dash', 'dot']).optional(),
});

// Zod schema for scales
const ScalesSchema = z.object({
  yMin: z.string().optional(),
  yMax: z.string().optional(),
});

// Zod schema for BarChartPanelOptions
export const BarChartPanelOptionsSchema = z.object({
  // Legend settings
  legendRule: z.string().optional(),
  legendEnabled: z.boolean().optional(),
  legendAlign: z.enum(['right', 'bottom']).optional(),
  legendAsTable: z.boolean().optional(),
  // Bar chart specific settings
  barMode: z.enum(['single', 'grouped', 'stacked', 'percent']).optional(),
  orientation: z.enum(['auto', 'vertical', 'horizontal']).optional(),
  xAxis: z.string().optional(),
  showValueLabels: z.boolean().optional(),
  valueSortOrder: z.enum(['none', 'asc', 'desc']).optional(), // 값 정렬 순서 (가로 바 차트용)
  // Bar color settings
  barColorMode: z.enum(['single', 'category', 'series']).optional(), // 색상 모드: single(단일), category(카테고리별), series(시리즈별)
  barColor: z.string().optional(), // 단일 색상 모드일 때 사용할 색상 (예: 'hsl(221, 83%, 53%)')
  // Chart settings
  unit: z.string().optional(),
  scales: ScalesSchema.optional(),
  thresholds: z.array(ThresholdSchema).optional(),
  chartDataNotExistMessage: z.string().optional(),
  tooltipLimit: z.number().optional(),
  // Variable interaction
  onClickVariable: z.string().optional(), // variable ID to set on bar click
});

// TypeScript type (inferred from Zod schema)
export type BarChartPanelOptions = z.infer<typeof BarChartPanelOptionsSchema>;
