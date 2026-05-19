import {format} from 'date-fns';
import {Table} from '@tanstack/react-table';
import {ColumnConfig, ExportConfig} from '../types';
import {formatValueByUnitType} from './columnsUtils';

// Core 전용 Export는 ColumnConfig 기반 (Shared는 기본 버전만 제공)

export const exportTableData = (
  table: Table<any>,
  columnConfigs: ColumnConfig[],
  exportConfig?: ExportConfig,
  title?: string,
) => {
  if (!table) return;

  // 현재 테이블에서 정렬/필터된 모든 행을 가져옴
  const filteredAndSortedRows = table.getSortedRowModel().rows;

  const exportData = filteredAndSortedRows.map((row) => {
    const originalData = row.original;

    let processedRow;
    try {
      processedRow = exportConfig?.mapData ? exportConfig.mapData(originalData) : {...originalData};

      const hasValidData =
        processedRow &&
        Object.values(processedRow).some((value) => value !== undefined && value !== null);
      if (!hasValidData) {
        processedRow = {...originalData};
      }
    } catch (error) {
      processedRow = {...originalData};
    }

    const formattedRow = {...processedRow};

    columnConfigs.forEach((col) => {
      const key = typeof col.key === 'string' ? col.key : String(col.key);

      if (processedRow[key] !== undefined) {
        if (
          (col.properties?.sortingFn === 'datetime' || col.properties?.unit === 'local_time') &&
          processedRow[key]
        ) {
          let value = processedRow[key];

          if (col.properties?.unit) {
            value = formatValueByUnitType(value, col.properties.unit);
          }

          formattedRow[key] = `\t${value}`;
        } else if (col.properties?.unit) {
          formattedRow[key] = formatValueByUnitType(processedRow[key], col.properties.unit);
        } else if (col.properties?.badge) {
          const value = String(processedRow[key] || '');
          formattedRow[key] = value.charAt(0).toUpperCase() + value.slice(1).toLowerCase();
        }
      }
    });

    return formattedRow;
  });

  if (exportData.length === 0) {
    console.error('Export: No data to export');
    return;
  }

  const configKeys = columnConfigs.map((col) =>
    typeof col.key === 'string' ? col.key : String(col.key),
  );
  const dataKeys = Object.keys(exportData[0] || {});

  const headers = [
    ...configKeys.filter((key) => dataKeys.includes(key)),
    ...dataKeys.filter((key) => !configKeys.includes(key)),
  ];

  const csvContent = [
    headers.join(','),
    ...exportData.map((row) =>
      headers
        .map((header) => {
          const value = row[header];
          // 값에 쉼표나 따옴표가 있으면 따옴표로 감싸기(현재 Date를 위한 처리)
          return typeof value === 'string' && (value.includes(',') || value.includes('"'))
            ? `"${value.replace(/"/g, '""')}"`
            : value;
        })
        .join(','),
    ),
  ].join('\n');

  const blob = new Blob([csvContent], {type: 'text/csv;charset=utf-8;'});
  const link = document.createElement('a');
  const url = URL.createObjectURL(blob);
  link.setAttribute('href', url);
  link.setAttribute(
    'download',
    `${exportConfig?.fileName || title || 'export'}_${format(new Date(), 'yyyyMMdd_HHmmss')}.csv`,
  );
  link.style.visibility = 'hidden';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
};
