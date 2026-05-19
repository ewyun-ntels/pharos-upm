/**
 * Core Datasource Editor Registration
 *
 * Register built-in datasource editors that are part of the core system
 */

import { datasourceEditorRegistry } from './DatasourceEditorRegistry';
import { DefaultSQLEditor } from '@features/dashboard/datasources';

// Register default SQL editor for common SQL databases
datasourceEditorRegistry.register({
  datasourceType: 'default',
  name: 'Default SQL Editor',
  description: 'Standard SQL editor for SQL databases',
  component: DefaultSQLEditor,
  priority: 1
});

// Register SQL editor for specific database types
datasourceEditorRegistry.register({
  datasourceType: 'postgresql',
  name: 'PostgreSQL Editor',
  description: 'SQL editor for PostgreSQL databases',
  component: DefaultSQLEditor,
  priority: 10
});

datasourceEditorRegistry.register({
  datasourceType: 'mysql',
  name: 'MySQL Editor',
  description: 'SQL editor for MySQL databases',
  component: DefaultSQLEditor,
  priority: 10
});

console.log('[CoreDatasourceEditors] ✅ Core datasource editors registered');
