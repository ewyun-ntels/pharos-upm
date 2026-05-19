import { z } from 'zod';

/**
 * Histogram Panel Options Schema
 * Defines the configuration options for Histogram panel type
 */
export const HistogramPanelOptionsSchema = z.object({
  // Bucket count (number of histogram bins)
  bucketCount: z.number().optional(),

  // Legend rule
  legendRule: z.string().optional(),

  // Y-axis display
  yaxis: z.boolean().optional(),

  // Tick font size
  tickFont: z.number().optional(),

  // Stack ID for stacked bars
  stackId: z.boolean().optional(),

  // Fill opacity
  fillOpacity: z.number().optional(),

  // Show average
  showAverage: z.boolean().optional(),

  // Show last
  showLast: z.boolean().optional(),

  // Show max
  showMax: z.boolean().optional(),

  // Show min
  showMin: z.boolean().optional(),

  // Show total
  showTotal: z.boolean().optional(),

  // Legend as table
  legendAsTable: z.boolean().optional(),

  // Legend enabled
  legendEnabled: z.boolean().optional(),

  // Legend option (position)
  legendOption: z.enum(['bottom', 'right']).optional(),

  // Show grid lines
  showGrid: z.boolean().optional(),

  // Unit
  unit: z.string().optional(),

  // Selected unit
  selectedUnit: z.string().optional(),
});

/**
 * Histogram Panel Options Type
 * Inferred from the Zod schema
 */
export type HistogramPanelOptions = z.infer<typeof HistogramPanelOptionsSchema>;
