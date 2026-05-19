import { z } from 'zod';

/**
 * DataOverview Panel Options Schema
 * Defines the configuration options for DataOverview panel type
 */
export const DataOverviewPanelOptionsSchema = z.object({
  // Data configurations
  datas: z.array(z.object({
    name: z.string(),
    path: z.string(),
  })).optional(),
  
  // Layout options
  layoutCol: z.boolean().optional(),
  titleTop: z.boolean().optional(),
  largeText: z.boolean().optional(),
  right: z.boolean().optional(),
  
  // Custom message when data doesn't exist
  chartDataNotExistMessage: z.string().optional(),
});

/**
 * DataOverview Panel Options Type
 * Inferred from the Zod schema
 */
export type DataOverviewPanelOptions = z.infer<typeof DataOverviewPanelOptionsSchema>;
