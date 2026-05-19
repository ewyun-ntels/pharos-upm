import React from 'react';
import {convertUnit, UnitType} from '../../../lib';
import {ColumnConfig, TableList} from '../types';

export const formatValueByUnitType = (
  value: any,
  unit: UnitType,
  decimalPlaces?: number,
): string => {
  if (!value || !unit) return value;

  switch (unit) {
    case 'local_time':
      return value === 'ᴺᵁᴸᴸ' ? '-' : convertUnit(value, unit).toString();
    case 'milliseconds': {
      const msVal = Number(value);
      return !isNaN(msVal) ? convertUnit(msVal, unit).toString() : 'NaN';
    }
    case 'Base64_decoding': {
      return convertUnit(value, unit).toString();
    }
    case 'numeric': {
      return convertUnit(value, unit, decimalPlaces).toString();
    }
    default:
      return convertUnit(Number(value), unit).toString();
  }
};

export const settingStyles = (value: any, style?: string): React.ReactNode | string => {
  if (!style) return value;

  switch (style) {
    case 'ellipsis':
      return React.createElement(
        'div',
        {
          className: 'text-ellipsis overflow-hidden whitespace-nowrap',
          title: value,
        },
        value,
      );
    default:
      return value;
  }
};

export const createCellContent = (
  value: any,
  row: any,
  config: any,
  formatValueByUnitType: (value: any, unit: UnitType, decimalPlaces: number) => string,
  settingStyles: (value: any, style?: string) => React.ReactNode | string,
) => {
  if (config.customCell) {
    return config.customCell(value, row);
  }

  let formattedValue = value;
  if (config.properties?.unit) {
    formattedValue = formatValueByUnitType(
      value,
      config.properties?.unit,
      config.properties?.decimalPlaces || 2,
    );
  }
  if (config.properties?.textStyle) {
    return settingStyles(formattedValue, config.properties?.textStyle);
  }
  return formattedValue;
};

export const getFileName = (row: TableList, keyColumnName?: string): string => {
  return row[keyColumnName || ''] || row.fileName || '';
};

export const calculateRowSpan = (hasGroups: boolean, isGrouped: boolean): number => {
  const DEFAULT_ROW_SPAN = 1;
  const GROUPED_ROW_SPAN = 2;

  return hasGroups && !isGrouped ? GROUPED_ROW_SPAN : DEFAULT_ROW_SPAN;
};

export const groupColumnsByGroup = (columnConfigs: ColumnConfig<TableList>[]) => {
  const grouped: Record<string, ColumnConfig<TableList>[]> = {};
  const ungrouped: ColumnConfig<TableList>[] = [];

  columnConfigs.forEach((config) => {
    const groupName = config.properties?.group;
    if (groupName) {
      if (!grouped[groupName]) {
        grouped[groupName] = [];
      }
      grouped[groupName].push(config);
    } else {
      ungrouped.push(config);
    }
  });

  return {grouped, ungrouped};
};

export const createInitialVisibility = (columnConfigs: ColumnConfig<TableList>[]) => {
  return columnConfigs.reduce(
    (acc, config) => {
      acc[String(config.key)] = !config.properties?.hide;
      return acc;
    },
    {} as Record<string, boolean>,
  );
};

export const findInitialSorting = (columnConfigs: ColumnConfig<TableList>[]) => {
  const defaultSortCol = columnConfigs.find((col) => col.properties?.defaultSort);
  return defaultSortCol
    ? [{id: String(defaultSortCol.key), desc: defaultSortCol.properties?.defaultSort === 'desc'}]
    : [];
};
