import {useMemo} from 'react';
import {useList} from '@/lib/data-provider';
import {PLUGIN_PROVIDER_NAME, PLUGIN_RESOURCES} from '@providers/plugin-provider';

/**
 * Datasource 리스트를 가져오는 커스텀 훅
 * pluginProvider를 사용하여 /plugins/datasources API 호출
 */
export function useDatasourceList() {
  const {query: {data: datasourceData, isLoading, error}} = useList({
    resource: PLUGIN_RESOURCES.DATASOURCES,
    dataProviderName: PLUGIN_PROVIDER_NAME,
    pagination: {mode: 'off'},
  });

  const dataSourceList = useMemo(() => {
    if (!datasourceData?.data) return [];

    const result = datasourceData.data.map((item: any) => ({
      label: item.name,
      value: item.name,
      type: item.type || 'default', // Add type field with default fallback
    }));

    result.sort((a, b) => a.label.localeCompare(b.label));

    return result;
  }, [datasourceData]);

  return {
    dataSourceList,
    isLoading,
    error,
  };
}
