import { z } from 'zod';

// Zod schema for color mapping
export const ColorMappingSchema = z.object({
  legendName: z.string(),
  color: z.string(),
});

export type ColorMapping = z.infer<typeof ColorMappingSchema>;

// Zod schema for alias colors (colors는 필수)
const AliasColorsSchema = z.object({
  colors: z.array(ColorMappingSchema),
});

// Zod schema for PiePanelOptions (내부 타입 정의)
export const PiePanelOptionsSchema = z.object({
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
  // Chart appearance
  chartType: z.enum(['pie', 'donut']).optional(),
  showLabel: z.enum(['none', 'value', 'percent']).optional(),
  innerRadius: z.number().optional(),
  outerRadius: z.number().optional(),
  aliasColors: AliasColorsSchema.optional(),
  // Pie stroke width (panel border)
  strokeWidth: z.number().min(0).max(10).optional(),
  // Additional fields from SharedChartOptions
  showType: z.string().optional(),
  useStatusPalette: z.boolean().optional(),
  chartDataNotExistMessage: z.any().optional(),
  unit: z.string().optional(),
  decimals: z.number().min(0).max(10).optional(),
});

// TypeScript type (inferred from Zod schema)
export type PiePanelOptions = z.infer<typeof PiePanelOptionsSchema>;
