import {ColumnConfig} from '@pharos/shared/hooks/table-columns';
import {z} from 'zod';

// Filter Item Types
export interface FilterItem {
  id: string;
  type: 'search' | 'export';
  placeholder?: string;
  disabled?: boolean;
}

// Data Link — 셀 클릭 시 이동할 링크 설정
// URL에 {{value}}, {{row.컬럼명}}, {{대시보드변수}} 사용 가능
export interface DataLink {
  title: string;
  url: string;
  targetBlank?: boolean;
}

// Zod schema for TablePanelOptions (내부 타입 정의)
// CardTable과 호환되는 타입 사용
export const TablePanelOptionsSchema = z.object({
  columnConfigs: z.custom<ColumnConfig[]>().optional(),
  leftItems: z.array(z.object({
    id: z.string(),
    type: z.enum(['search', 'export']),
    placeholder: z.string().optional(),
    disabled: z.boolean().optional(),
  })).optional(),
  rightItems: z.array(z.object({
    id: z.string(),
    type: z.enum(['search', 'export']),
    placeholder: z.string().optional(),
    disabled: z.boolean().optional(),
  })).optional(),
  checkboxConfig: z.object({
    showCheckbox: z.boolean().optional(),
    checkboxKey: z.string().optional(),
  }).optional(),
  columnFilters: z.array(z.object({
    id: z.string(),
    value: z.any(),
  })).optional(),
  properties: z.record(z.string(), z.any()).optional(),
  usePagenation: z.boolean().optional(),
  columnDataLinks: z.record(
    z.string(),
    z.array(z.object({
      title: z.string(),
      url: z.string(),
      targetBlank: z.boolean().optional(),
    }))
  ).optional(),
});

// TypeScript type (inferred from Zod schema)
export type TablePanelOptions = z.infer<typeof TablePanelOptionsSchema>;

