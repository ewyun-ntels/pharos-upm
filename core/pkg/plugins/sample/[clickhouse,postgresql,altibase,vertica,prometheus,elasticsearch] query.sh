#!/bin/sh

curl -X POST http://192.168.15.101:30996/plugins/ds/query \
     -d '
{
  "queries": [
    {
      "id": "A",
      "datasourceName": "sample-clickhouse",
      "sql": "SELECT * FROM tarzan.database LIMIT 2"
    },
    {
      "id": "B",
      "datasourceName": "sample-postgresql",
      "sql": "SELECT * FROM pg_settings LIMIT 2"
    },
    {
      "id": "C",
      "datasourceName": "sample-altibase",
      "sql": "SELECT * FROM SYSTEM_.sys_tables_ LIMIT 2;"
    },
    {
      "id": "D",
      "datasourceName": "sample-vertica",
      "sql": "SELECT table_schema, table_name FROM v_catalog.tables LIMIT 2"
    },
    {
      "id": "E",
      "datasourceName": "sample-prometheus",
      "sql": "up"
    },
    {
      "id": "F",
      "datasourceName": "sample-elasticsearch",
      "sql": "FROM kafka-ckafka-*"
    },
    {
      "id": "G",
      "datasourceName": "invalid",
      "sql": "SELECT * FROM pg_settings LIMIT 2"
    }
  ]
}'
printf "\n"
