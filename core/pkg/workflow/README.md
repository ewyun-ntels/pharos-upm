# API
 - task
   - GET /workflow/task/:name
   - GET /workflow/task
     - 
     ```bash
     [
       {
         "dataFormat": {
           "field-01": "value-01"
         },
         "name": "sample-task-01"
       },
     ...
       {
         "dataFormat": "",
         "name": "collect-wal"
       }
     ]
     ```
 - job
   - GET /workflow/job/:name
   - GET /workflow/job
   - POST /workflow/job
     - 
     ```bash
     curl -X 'POST' http://192.168.15.101:30996/workflow/job \
       -H 'accept: application/json' \
       -H 'Content-Type: application/json' \
       -d '
     {
       "name": "test-01",
       "hosts": ["tarzan-agent-01-0"],
       "schedule": "*/10 * * * * *",
       "active": true,
       "taskNames": ["sample-task-01"],
       "taskDatas": {"sample-task-01":{"field-01":"value-01"}}
     }'
     ```
   - PUT /workflow/job/:name
     - 
     ```bash
     curl -X 'PUT' http://192.168.15.101:30996/workflow/job/test-01 \
       -H 'accept: application/json' \
       -H 'Content-Type: application/json' \
       -d '
     {
       "hosts": ["tarzan-agent-01-0"],
       "schedule": "*/10 * * * * *",
       "active": false,
       "taskNames": ["sample-task-01"],
       "taskDatas": {"sample-task-01":{"field-01":"value-01"}}
     }'
     ```
   - DELETE /workflow/job/:name
 - execution
   - GET /workflow/execution/:job
   - GET /workflow/execution
     - 
     ```bash
     [
       {
         "id": "8256ed50-47ed-487c-af1a-c9608278a69a",
         "job": "collect-agent-*1m",
         "modifiedTs": "2025-05-21T07:15:00.080044877Z",
         "startTs": "2025-05-21T07:15:00.001667917Z",
         "state": "successful",
         "tasks": [{"name":"collect-xid","state":"successful","startTs":"2025-05-21T07:15:00.017171949Z"}]
       },
     ...
     ]
     ```
