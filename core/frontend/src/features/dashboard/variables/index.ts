/**
 * Variable Plugin System
 * 
 * This module exports the variable plugin registry and imports all variable plugins
 * to trigger their auto-registration.
 * 
 * Import this module to ensure all variable plugins are registered.
 */

// Export registry
export {variablePluginRegistry} from './registry/VariablePluginRegistry';
export type {
  VariablePlugin,
  VariablePluginInfo,
  VariableProps,
  VariableEditorProps,
} from './registry/VariablePluginRegistry';

// Import all plugins to trigger registration
import './plugins/select/SelectVariable';
import './plugins/datetime/DateTimeVariable';
import './plugins/step/StepVariable';
import './plugins/refresh/RefreshVariable';
import './plugins/custom/CustomVariable';
import './plugins/tableSearch/TableSearchVariable';
// Note: tagsinput plugin removed - now handled by select with isMulti option
