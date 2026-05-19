/**
 * Datasource Plugin System
 * 
 * This module exports the datasource editor registry and types,
 * and imports core datasource editors to trigger their registration.
 * 
 * 내부(core) 및 외부(extension) 모두에서 사용 가능한 통합 export
 */

// Export registry and types
export { datasourceEditorRegistry } from './registry/DatasourceEditorRegistry';
export type {
  DatasourceEditorPlugin,
  DatasourceEditorComponentProps,
} from './registry/DatasourceEditorRegistry';

// Re-export DefaultSQLEditor for extensions
export { DefaultSQLEditor } from './plugins/default/DefaultSQLEditor';

// Import core datasource editors to trigger registration
import './registry/coreDatasourceEditors';
