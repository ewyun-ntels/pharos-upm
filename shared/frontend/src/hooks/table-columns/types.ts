import {UnitType} from '../../lib';
import type {CheckboxConfig as BaseCheckboxConfig, ExportConfig as BaseExportConfig} from '../../components/ui-extension/data-grid';

export type TableList = {
  [key: string]: any;
  select: string;
  actions: string;
};

// Re-export with TableList as default type for backward compatibility
export type CheckboxConfig = BaseCheckboxConfig<TableList>;
export type ExportConfig = BaseExportConfig<TableList>;

// Core 전용 타입 (PDF 다운로드는 Core에서만 사용)
export interface PdfDownloadConfig {
  isBulkCollecting?: boolean;
  downloadingFiles?: Set<string>;
  handleDownload: (fileNames: string[]) => void;
  PdfButtonSize?: 'default' | 'sm' | 'lg' | 'icon' | null;
  resource?: string;
  meta?: any;
  queryOptions?: any;
}

interface PropertiesDev {
  searchByFormatted?: boolean;
}

interface Properties extends PropertiesDev {
  title?: string;
  size?: number;
  defaultSort?: 'asc' | 'desc';
  unit?: UnitType;
  decimalPlaces?: number; // 소수점 자리수
  textStyle?: string;
  hide?: boolean;
  badge?: {
    [key: string]: {
      color: string;
      textColor?: string;
      displayText?: string;
    };
  };
  group?: string;
  enableSorting?: boolean; // tanstack 제공 기능
  sortingFn?: 'datetime' | 'alphanumeric' | 'textCaseSensitive' | 'text' | 'basic'; // tanstack 제공 기능
  colorRules?: Array<{
    condition: 'gt' | 'gte' | 'lt' | 'lte' | 'eq';
    value: number;
    color: string;
    textColor?: string;
  }>;
  gauge?: {
    min: number;
    max: number;
    color?: string;
  };
}

export interface ColumnConfig<T = any> {
  key: keyof T;
  properties?: Properties;
  customCell?: (value: T[keyof T], row: T) => React.ReactNode;
  onCellClick?: (value: T[keyof T], row: T) => void;
  cellType?: string /* sheet 사용을 위한 type */;
  filterFn?: (row: T, columnId: string, filterValue: any) => boolean;
  getUniqueValues?: (row: T) => any[];
}

export interface DialogConfig {
  dialogTitle?: string;
  dialogBtnName?: string;
  handleModalOpen: (fileName: string) => void;
}
