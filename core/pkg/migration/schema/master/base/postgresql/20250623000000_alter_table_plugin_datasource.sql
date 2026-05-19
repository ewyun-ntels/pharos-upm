-- +goose Up
ALTER TABLE plugin_datasource DROP COLUMN secure_data;

DELETE FROM plugin_datasource WHERE name='sample-clickhouse';
DELETE FROM plugin_datasource WHERE name='sample-postgresql';
DELETE FROM plugin_datasource WHERE name='sample-http-receiver';

-- +goose Down
ALTER TABLE plugin_datasource ADD COLUMN secure_data JSONB;
