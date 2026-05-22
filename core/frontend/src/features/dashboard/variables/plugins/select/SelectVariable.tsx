import React, {useMemo, useCallback, useEffect} from 'react';
import {VariablePlugin, variablePluginRegistry, VariableProps} from '@features/dashboard/variables';
import {VariableSelect} from '@pharos/shared/components/ui-extension';
import {useDashboardQuery} from '@features/dashboard/hooks/useDashboardQuery';
import {useDashboardStore} from '@features/dashboard/hooks/use-dashboard-store';
import {useDatasourceList} from '@hooks/use-datasource-list';
import {DASHBOARD_RESOURCES} from '@providers/dashboard-provider';
import {SelectVariableEditor} from './SelectVariableEditor';
import {SelectVariablePreview} from './SelectVariablePreview';
import type {SelectFilterOptions} from './types';
import { ActionSchema } from '@pharos/shared/types/dashboard';

/**
 * Select Variable Component (React Query Pure)
 */
export const SelectVariable: React.FC<VariableProps<SelectFilterOptions>> = (props) => {
  const {id, label, datasourceName, query, options, onChange, dashboardId, kind, permission} = props;
  const isMultiSelect = options?.isMulti === true;
  const shouldDefaultToAll = isMultiSelect && options?.showAllOption && options?.defaultAllSelected;
  const shouldAutoSelectFirst = options?.autoSelectFirstOption !== false && !shouldDefaultToAll;
  const {dataSourceList} = useDatasourceList();

  // Query object 구성: 권한에 따라 Panel Mode vs Run Mode
  // - editor/owner (Run Mode): query + datasourceName로 직접 실행
  // - viewer (Panel Mode): id + kind로 저장된 variable 실행
  const useRunMode = permission === ActionSchema.enum.editor || permission === ActionSchema.enum.owner;
  const datasourceType = useMemo(
    () => dataSourceList.find(ds => ds.value === datasourceName)?.type,
    [dataSourceList, datasourceName],
  );

  useEffect(() => {
    if (!useRunMode || !datasourceName) return;
    useDashboardStore.getState().setFilterMeta(id, {
      id,
      query: query ?? '',
      datasourceName,
      datasourceType,
      options,
    });
  }, [id, query, useRunMode, datasourceName, datasourceType, options]);
  
  const queries = useMemo(() => {
    if (useRunMode) {
      // Run Mode: query 직접 실행 (editor/owner)
      return [{
        query: query ?? '',
        datasourceName: datasourceName ?? '',
        datasourceType,
        dashboardId: dashboardId,
      }];
    } else {
      // Panel Mode: 저장된 variable 실행 (viewer)
      return [{
        id: id,
        kind: kind,
        dashboardId: dashboardId,
        query: '',
        datasourceName: '',
      }];
    }
  }, [useRunMode, query, id, dashboardId, datasourceName, datasourceType, kind]);

  // ✅ useDashboardQuery 사용 (DAG Pipeline + Direct Provider)
  // dependency values는 내부에서 자동으로 queryKey에 추가됨
  const {
    data,
    isSuccess,
    isFetching,
    isError,
    value,
    setValue,
  } = useDashboardQuery<unknown[]>({
    id,
    dashboardId,
    queries,
    resource: DASHBOARD_RESOURCES.RAW,
    useFilterMeta: true,
    useDependencyGraph: true,
    autoSelectFirst: shouldAutoSelectFirst,
    labelKey: options?.labelKey,
    valueKey: options?.valueKey,
  });

  // ✅ Format options
  const optionList = useMemo(() => {
    if (!isSuccess || !data) return [];
    
    const labelKey = options?.labelKey ?? 'label';
    const valueKey = options?.valueKey ?? 'value';
    
    return data.map((item: any) => ({
      label: item[labelKey] ?? item[labelKey.toUpperCase()],
      value: item[valueKey] ?? item[valueKey.toUpperCase()],
    }));
  }, [data, isSuccess, options?.labelKey, options?.valueKey]);

  // Handle both single and multi values
  // IMPORTANT: Keep undefined as undefined so VariableSelect can distinguish
  // between "not yet set" (undefined) and "explicitly cleared" ([])
  const storeValue = useMemo(() => {
    if (value === undefined) return undefined;
    
    return isMultiSelect 
      ? (Array.isArray(value) ? value : [])
      : (typeof value === 'string' ? value : '');
  }, [isMultiSelect, value]);

  // ✅ 에러 발생 시 빈 옵션으로 표시 (console에는 이미 로그됨)
  const displayOptions = isError ? [] : optionList;

  const handleValueChange = useCallback((newValue: string | string[]) => {
    // Prevent setting the same value (infinite loop protection)
    const currentValue = isMultiSelect 
      ? (Array.isArray(value) ? [...value].sort().join(',') : '')
      : (typeof value === 'string' ? value : '');
    
    const newValueStr = Array.isArray(newValue) ? [...newValue].sort().join(',') : newValue;
    
    if (currentValue === newValueStr) {
      console.log(`[SelectVariable ${id}] Skipping duplicate value:`, newValueStr);
      return; // Same value, don't trigger update
    }
    
    console.log(`[SelectVariable ${id}] Value changed:`, currentValue, '→', newValueStr);
    setValue?.(newValue);
    onChange?.(newValue);
  }, [isMultiSelect, value, setValue, onChange, id]);

  // Don't render if hidden
  if (options?.isHidden) {
    return null;
  }

  return (
    <VariableSelect
      label={label}
      options={displayOptions}
      value={storeValue}
      onChange={handleValueChange}
      multiple={isMultiSelect}
      searchable={options?.isSearchable}
      disabled={isFetching}
      isLoading={isFetching}
      showLabel={true}
      className="min-w-50"
      displayMode={options?.displayMode || 'value-threshold'}
      displayThreshold={options?.displayThreshold || 1}
      showAllOption={options?.showAllOption}
      defaultAllSelected={options?.defaultAllSelected}
      customAllValue={options?.customAllValue}
      autoSelectFirstOption={shouldAutoSelectFirst}
    />
  );
};

/**
 * Select Variable Plugin
 */
export const selectVariablePlugin: VariablePlugin<SelectFilterOptions> = {
  info: {
    id: 'select',
    label: 'Select Box',
    description: 'Single selection dropdown',
    category: 'input',
    requiresData: true,
  },
  component: SelectVariable,
  editor: SelectVariableEditor,
  preview: SelectVariablePreview,
};

// Register
variablePluginRegistry.register(selectVariablePlugin);
