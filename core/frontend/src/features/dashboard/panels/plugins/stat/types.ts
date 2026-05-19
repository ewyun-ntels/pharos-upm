import { z } from 'zod';

// Zod schema for StatPanelOptions
export const StatPanelOptionsSchema = z.object({
  chartColor: z.string().optional(),
  chartType: z.enum(['area', 'none']).optional(),
  fillOpacity: z.number().min(0).max(1).optional(),
  showType: z.enum(['last', 'first', 'average', 'max', 'min']).optional(),
  chartDataNotExistMessage: z.string().optional(),
  textMode: z.enum(['auto', 'value', 'value_and_name', 'name', 'none']).optional(),
  textAlignment: z.enum(['center', 'justify']).optional(),
  displayName: z.string().optional(),
  decimals: z.number().min(0).max(10).optional(),
  wideLayout: z.boolean().optional(),
  thresholds: z.array(z.object({
    value: z.number(),
    color: z.string(),
  })).optional(),
});

// TypeScript type (inferred from Zod schema)
export type StatPanelOptions = z.infer<typeof StatPanelOptionsSchema>;
