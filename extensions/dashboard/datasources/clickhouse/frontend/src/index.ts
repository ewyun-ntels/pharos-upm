import { ClickHouseDatasourceEditor } from './plugins/clickhouse/ClickHouseDatasourceEditor';
import { datasourceEditorRegistry } from '@pharos/core/datasource-editor-registry';

export { ClickHouseDatasourceEditor } from './plugins/clickhouse/ClickHouseDatasourceEditor';
export type { ClickHouseDataSourceConfig, ClickHouseQuery } from './types';

// Register ClickHouse datasource editor directly
datasourceEditorRegistry.register({
  datasourceType: 'clickhouse',
  name: 'ClickHouse Query Builder',
  description: 'Advanced SQL editor for ClickHouse with schema assistance',
  component: ClickHouseDatasourceEditor,
  priority: 20
});

console.log('[ClickHouse Extension] Datasource editor registered');

// Extension entry point for metadata
export default {
  name: 'clickhouse-datasource',
  version: '1.0.0',
  type: 'datasource'
};