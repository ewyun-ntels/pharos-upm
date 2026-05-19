import React, {useMemo} from 'react';
import {createColumnHelper, ColumnDef, VisibilityState} from '@tanstack/react-table';
import {DataTableColumnSortHeader} from '../../components/ui-extension';
import {Button, Checkbox, Badge} from '../../components/ui';
import {Download} from 'lucide-react';
import {ColumnConfig, DialogConfig, PdfDownloadConfig, CheckboxConfig, TableList} from './types';
import {getSelectionKey} from './utils/selectionUtils';
import {
  formatValueByUnitType,
  settingStyles,
  createCellContent,
  getFileName,
  calculateRowSpan,
  groupColumnsByGroup,
  createInitialVisibility,
  findInitialSorting,
} from './utils/columnsUtils';

interface UseTableColumnsProps {
  columnConfigs: ColumnConfig<TableList>[];
  keyColumnName?: string;
  enableDialog?: boolean;
  dialogConfig?: DialogConfig;
  enablePdfDownload?: boolean;
  pdfDownloadConfig?: PdfDownloadConfig;
  showCheckbox?: boolean;
  checkboxConfig?: CheckboxConfig;
}

interface UseTableColumnsReturn {
  columns: ColumnDef<TableList, any>[];
  columnVisibility: VisibilityState;
  setColumnVisibility: React.Dispatch<React.SetStateAction<VisibilityState>>;
  initialSorting: Array<{id: string; desc: boolean}>;
}

const COLUMN_SIZES = {
  CHECKBOX: 40,
  DIALOG: 120,
  DOWNLOAD: 150,
} as const;

const createCheckboxColumn = (
  columnHelper: ReturnType<typeof createColumnHelper<TableList>>,
  config: CheckboxConfig,
  keyColumnName?: string,
  hasGroups = false,
): ColumnDef<TableList, any> => {
  const {
    checkboxKey,
    selectedItems = [],
    currentPageData = [],
    handleSelectAll,
    handleSelectOne,
    isSelectable = true,
  } = config;

  const selectableCount = currentPageData.filter(() => isSelectable).length;
  const isAllSelected = selectableCount > 0 && selectedItems.length === selectableCount;

  return columnHelper.display({
    id: 'checkbox',
    header: () => (
      <Checkbox
        checked={isAllSelected}
        onCheckedChange={(checked) => handleSelectAll?.(!!checked, currentPageData.filter(() => isSelectable))}
        className="translate-y-0.5 border-primary shadow-none cursor-pointer"
        disabled={selectableCount === 0}
      />
    ),
    cell: ({row}) => {
      const selectionKey = String(getSelectionKey?.(row.original, checkboxKey, keyColumnName));
      const isSelected = selectedItems.includes(selectionKey);
      const disabled =
        typeof isSelectable === 'function' ? !isSelectable(row.original) : !isSelectable;

      return (
        <Checkbox
          checked={isSelected}
          onCheckedChange={(checked) => handleSelectOne?.(row.original, !!checked)}
          className="translate-y-0.5 border-primary shadow-none cursor-pointer"
          disabled={disabled}
        />
      );
    },
    size: COLUMN_SIZES.CHECKBOX,
    maxSize: COLUMN_SIZES.CHECKBOX,
    meta: {
      rowSpan: calculateRowSpan(hasGroups, false),
    },
  });
};

const createDialogColumn = (
  columnHelper: ReturnType<typeof createColumnHelper<TableList>>,
  config: DialogConfig,
  keyColumnName?: string,
  hasGroups = false,
): ColumnDef<TableList, any> => {
  const {dialogTitle = 'View', dialogBtnName = dialogTitle, handleModalOpen} = config;

  return columnHelper.display({
    id: 'dialog',
    header: () => dialogTitle,
    cell: ({row}) => {
      const fileName = getFileName(row.original, keyColumnName);
      return (
        <Button
          variant="link"
          className="cursor-pointer p-0"
          size="default"
          onClick={() => handleModalOpen(fileName)}
        >
          {dialogBtnName}
        </Button>
      );
    },
    size: COLUMN_SIZES.DIALOG,
    maxSize: COLUMN_SIZES.DIALOG,
    meta: {
      rowSpan: calculateRowSpan(hasGroups, false),
    },
  });
};

const createDownloadColumn = (
  columnHelper: ReturnType<typeof createColumnHelper<TableList>>,
  config: PdfDownloadConfig,
  keyColumnName?: string,
  hasGroups = false,
): ColumnDef<TableList, any> => {
  const {
    isBulkCollecting = false,
    downloadingFiles = new Set(),
    handleDownload,
    PdfButtonSize = 'default',
  } = config;

  return columnHelper.display({
    id: 'download',
    header: () => 'Download',
    cell: ({row}) => {
      const fileName = getFileName(row.original, keyColumnName);
      const isDownloading = downloadingFiles.has(fileName);

      return (
        <Button
          size={PdfButtonSize}
          variant="outline"
          disabled={isDownloading || isBulkCollecting}
          onClick={() => handleDownload([fileName])}
        >
          <Download className="w-4 h-4" />
          Download
        </Button>
      );
    },
    size: COLUMN_SIZES.DOWNLOAD,
    maxSize: COLUMN_SIZES.DOWNLOAD,
    meta: {
      rowSpan: calculateRowSpan(hasGroups, false),
    },
  });
};

const createDataColumn = (
  columnHelper: ReturnType<typeof createColumnHelper<TableList>>,
  config: ColumnConfig<TableList>,
  hasGroups: boolean,
): ColumnDef<TableList, any> => {
  const processedConfig = {
    ...config,
    properties: {
      ...config.properties,
      hide: config.properties?.hide ?? false,
    },
  };

  const headerValue = processedConfig.properties?.title || String(processedConfig.key);
  const rowSpan = calculateRowSpan(hasGroups, !!processedConfig.properties?.group);

  return columnHelper.accessor(processedConfig.key as string, {
    id: String(processedConfig.key),
    header: ({column}) =>
      (config.properties?.enableSorting ?? (config as any).enableSorting ?? true) ? (
        <DataTableColumnSortHeader title={headerValue} column={column} />
      ) : (
        headerValue
      ),
    cell: ({row}) => {
      const value = row.original[processedConfig.key];

      if (processedConfig.properties?.gauge) {
        const { min, max, color } = processedConfig.properties.gauge;
        const numVal = parseFloat(String(value));
        const pct = isNaN(numVal) ? 0 : Math.min(100, Math.max(0, ((numVal - min) / (max - min)) * 100));
        const barColor = color || '#3b82f6';
        return (
          <div className="relative w-full h-5 rounded overflow-hidden" style={{ backgroundColor: 'hsl(var(--muted))' }}>
            <div className="absolute left-0 top-0 h-full rounded" style={{ width: `${pct}%`, backgroundColor: barColor }} />
            <span className="absolute inset-0 flex items-center justify-center text-xs font-medium">
              {String(value ?? '')}
            </span>
          </div>
        );
      }

      if (processedConfig.properties?.colorRules?.length) {
        const numVal = parseFloat(String(value));
        const matched = processedConfig.properties.colorRules.find((rule) => {
          switch (rule.condition) {
            case 'gt': return numVal > rule.value;
            case 'gte': return numVal >= rule.value;
            case 'lt': return numVal < rule.value;
            case 'lte': return numVal <= rule.value;
            case 'eq': return numVal === rule.value;
          }
        });
        if (matched) {
          return (
            <div
              className="px-2 py-0.5 rounded text-center text-xs font-medium"
              style={{ backgroundColor: matched.color, color: matched.textColor || '#fff' }}
            >
              {String(value ?? '')}
            </div>
          );
        }
      }

      if (processedConfig.properties?.badge) {
        const badgeKey = String(value);
        const badgeInfo = Object.entries(processedConfig.properties.badge).find(
          ([key]) => key.toLowerCase() === badgeKey.toLowerCase(),
        )?.[1];

        if (badgeInfo) {
          return (
            <div className="w-full flex justify-center items-center">
              <Badge
                variant="outline"
                className={`${badgeInfo.textColor ? 'border' : 'border-0'} text-white  shadow-none`}
                style={{
                  backgroundColor: badgeInfo.color,
                  color: badgeInfo.textColor || '#FFFFFF',
                  borderColor: badgeInfo.textColor,
                }}
              >
                {badgeInfo.displayText || badgeKey}
              </Badge>
            </div>
          );
        }
        return value;
      }

      // 일반 셀 내용 생성
      const cellContent = createCellContent(
        value,
        row.original,
        processedConfig,
        formatValueByUnitType,
        settingStyles,
      );

      return processedConfig.onCellClick ? (
        <div
          onClick={() => processedConfig.onCellClick?.(value, row.original)}
          className="cursor-pointer"
        >
          {cellContent}
        </div>
      ) : (
        cellContent
      );
    },
    ...processedConfig.properties,
    enableSorting: config.properties?.enableSorting ?? (config as any).enableSorting ?? true,
    size: processedConfig.properties?.size,
    // maxSize는 설정하지 않음: 저장된 size가 maxSize로 가면 리로드 후 콜럼을 늘릴 수 없어짐
    filterFn: config.filterFn
      ? (row, columnId, filterValue) => config.filterFn!(row.original, columnId, filterValue)
      : undefined,
    getUniqueValues: config.getUniqueValues
      ? (row) => config.getUniqueValues!(row)
      : undefined,
    meta: {
      rowSpan,
      searchByFormatted: processedConfig.properties?.searchByFormatted,
    },
  });
};

export function useTableColumns(props: UseTableColumnsProps): UseTableColumnsReturn {
  const {
    keyColumnName,
    columnConfigs,
    enableDialog = false,
    dialogConfig,
    enablePdfDownload = false,
    pdfDownloadConfig,
    showCheckbox = false,
    checkboxConfig,
  } = props;

  // columnConfigs에 따른 가시성 상태
  const columnVisibility = useMemo(() => createInitialVisibility(columnConfigs), [columnConfigs]);

  // 컬럼 그룹핑
  const {grouped: groupedColumns, ungrouped: ungroupedColumns} = useMemo(
    () => groupColumnsByGroup(columnConfigs),
    [columnConfigs],
  );

  const hasGroups = Object.keys(groupedColumns).length > 0;
  const columnHelper = createColumnHelper<TableList>();

  const columns = useMemo(() => {
    const result: ColumnDef<TableList, any>[] = [];

    if (showCheckbox && checkboxConfig) {
      result.push(createCheckboxColumn(columnHelper, checkboxConfig, keyColumnName, hasGroups));
    }

    // 일반 컬럼들
    result.push(
      ...ungroupedColumns.map((config) => createDataColumn(columnHelper, config, hasGroups)),
    );

    // 그룹 컬럼들
    Object.entries(groupedColumns).forEach(([groupName, groupCols]) => {
      result.push(
        columnHelper.group({
          id: groupName,
          header: groupName,
          columns: groupCols.map((config) => createDataColumn(columnHelper, config, hasGroups)),
        }),
      );
    });

    if (enableDialog && dialogConfig) {
      result.push(createDialogColumn(columnHelper, dialogConfig, keyColumnName, hasGroups));
    }

    if (enablePdfDownload && pdfDownloadConfig) {
      result.push(createDownloadColumn(columnHelper, pdfDownloadConfig, keyColumnName, hasGroups));
    }

    return result;
  }, [
    columnHelper,
    showCheckbox,
    checkboxConfig,
    ungroupedColumns,
    groupedColumns,
    hasGroups,
    enableDialog,
    dialogConfig,
    enablePdfDownload,
    pdfDownloadConfig,
    keyColumnName,
  ]);

  // 초기 정렬
  const initialSorting = useMemo(() => findInitialSorting(columnConfigs), [columnConfigs]);

  return {
    columns,
    columnVisibility,
    setColumnVisibility: () => {}, // 임시로 빈 함수 (필요시 구현)
    initialSorting,
  };
}
