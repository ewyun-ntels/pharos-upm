/**
 * Extension Registry - Main Export
 */

export {
  ExtensionRegistry,
  registerExtension,
  getExtensions,
  getExtensionPages,
  getExtensionWidgets,
  getMenuItems,
  getExtensionPanels,
  getExtensionDataProviders,
  getBranding,
} from './ExtensionRegistry';

export type {
  Extension,
  ExtensionMetadata,
  ExtensionPage,
  ExtensionWidget,
  ExtensionPanelPlugin,
  Branding,
} from './ExtensionRegistry';

// Re-export MenuItem type for convenience
export type { MenuItem } from '@pharos/shared/features/extension';
