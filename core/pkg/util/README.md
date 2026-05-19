# example
 - postgresql_metric_collect
   - `./main util postgresql_metric_collect 192.168.15.103 5432 admin admin postgres`
 - certification
   - `./main util certification localhost 192.168.15.101,127.0.0.1 2125-01-01 /pharos/cert`
 - convert_metric_entries
   - `./main util convert_metric_entries TMO_Metric_Table_20250423_rev2.xlsx ./config`
 - centrifuge_publish
   - `./main util centrifuge_publish ws://192.168.15.101:30996/websocket/centrifuge http-receiver:sample-http-receiver {"body":"","fullPath":"/receive","method":"GET"}`
