import {z} from 'zod';

// Zod schema for CustomAlertPanelOptions
export const CustomAlertPanelOptionsSchema = z.object({
  // custom_alert 패널은 특별한 옵션이 없음 (Alert 데이터를 표시하는 패널)
  chartDataNotExistMessage: z.string().optional(),
});

// TypeScript type (inferred from Zod schema)
export type CustomAlertPanelOptions = z.infer<typeof CustomAlertPanelOptionsSchema>;
