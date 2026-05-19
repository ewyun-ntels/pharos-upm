import {z} from 'zod';

// Zod schema for BarGaugePanelOptions
export const BarGaugePanelOptionsSchema = z.object({
  legendRule: z.string().optional(),
  unit: z.string().optional(),
  chartDataNotExistMessage: z.string().optional(),
});

// TypeScript type (inferred from Zod schema)
export type BarGaugePanelOptions = z.infer<typeof BarGaugePanelOptionsSchema>;
