import { z } from 'zod';

/**
 * BarPercentGauge Panel Options Schema
 * Defines the configuration options for BarPercentGauge panel type
 */
export const BarPercentGaugePanelOptionsSchema = z.object({
  // Chart color
  chartColor: z.string().optional(),
  
  // Show type
  showType: z.string().optional(),
  
  // Legend rule
  legendRule: z.string().optional(),
  
  // Sub name
  subName: z.string().optional(),
  
  // Gauge max value
  gaugeMaxValue: z.number().optional(),
  
  // Unit type
  unitType: z.string().optional(),
  
  // Custom message when data doesn't exist
  chartDataNotExistMessage: z.string().optional(),
});

/**
 * BarPercentGauge Panel Options Type
 * Inferred from the Zod schema
 */
export type BarPercentGaugePanelOptions = z.infer<typeof BarPercentGaugePanelOptionsSchema>;
