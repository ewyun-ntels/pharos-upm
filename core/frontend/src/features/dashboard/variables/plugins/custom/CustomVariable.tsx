import React from 'react';
import {VariablePlugin, variablePluginRegistry, VariableProps} from '@features/dashboard/variables';
import type {CustomFilterOptions} from './types';

/**
 * Custom Variable Component
 * Renders a custom React component provided in options
 */
export const CustomVariable: React.FC<VariableProps<CustomFilterOptions>> = (props) => {
  const {options} = props;

  return <>{options?.customComponent}</>;
};

/**
 * Custom Variable Plugin
 */
export const customVariablePlugin: VariablePlugin<CustomFilterOptions> = {
  info: {
    id: 'custom',
    label: 'Custom',
    description: 'Renders a custom React component',
    category: 'custom',
    requiresData: false,
  },
  component: CustomVariable,
};

// Register
variablePluginRegistry.register(customVariablePlugin);
