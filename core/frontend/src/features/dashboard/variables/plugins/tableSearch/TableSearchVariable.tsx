import React from 'react';
import {VariablePlugin, variablePluginRegistry, VariableProps} from '@features/dashboard/variables';
import {Search, X} from '@pharos/shared/components';
import {Input} from '@pharos/shared/components/ui';
import {IconButton} from '@pharos/shared/components/ui-extension';
import {useDashboardStore} from '@features/dashboard/hooks/use-dashboard-store';
import type {TableSearchFilterOptions} from './types';

/**
 * TableSearch Variable Component
 */
export const TableSearchVariable: React.FC<VariableProps<TableSearchFilterOptions>> = (props) => {
  const {options} = props;

  const filterMetas = useDashboardStore((state) => state.filterState?.filterMetas ?? new Map());
  const setFilter = useDashboardStore((state) => state.setFilter);

  const searchValue = filterMetas.get('search')?.value;
  const inputValue = (typeof searchValue === 'string' ? searchValue : '') || '';

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const targetValue = event.target.value;
    setFilter('search', targetValue); // ✅ URL 자동 동기화 (var-search)
  };

  const handleKeyDown: React.KeyboardEventHandler<HTMLInputElement> = (event) => {
    if (event.key === 'Enter') {
      if (inputValue) {
        setFilter('search', inputValue);
      }
    } else if (event.key === 'Escape') {
      handleClear();
    }
  };

  const handleClear = () => {
    setFilter('search', '');
  };

  return (
    <div className="search-area layout-flex-row relative flex-wrap lg:flex-nowrap gap-[6px]">
      <div className="search-input relative">
        <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
        <Input
          autoComplete="off"
          className="w-full sm:w-64 pl-8"
          value={inputValue || ''}
          placeholder={options?.placeholder || 'Search'}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown}
        />
        {inputValue.length !== 0 && (
          <IconButton
            onClick={handleClear}
            icon={<X />}
            className="absolute top-1/2 right-2.5 transform -translate-y-1/2 w-5 h-5 p-0 rounded-full"
          >
            초기화
          </IconButton>
        )}
      </div>
    </div>
  );
};

/**
 * TableSearch Variable Plugin
 */
export const tableSearchVariablePlugin: VariablePlugin<TableSearchFilterOptions> = {
  info: {
    id: 'tableSearch',
    label: 'Table Search',
    description: 'Search input for table filtering',
    category: 'input',
    requiresData: false,
  },
  component: TableSearchVariable,
};

// Register
variablePluginRegistry.register(tableSearchVariablePlugin);
