export interface ClickHouseDataSourceConfig {
  host: string;
  port: number;
  database: string;
  username: string;
  password: string;
  max_open_connection: number;
  max_lifetime: number;
}

export interface ClickHouseQuery {
  sql: string;
  timeout?: number;
}