# Pharos Plugins API

Pharos 플러그인 API를 통해 데이터소스를 관리하고 쿼리를 실행할 수 있습니다.

## 목차
- [개요](#개요)
- [지원 플러그인](#지원-플러그인)
- [플러그인 관리](#플러그인-관리)
- [데이터소스 관리](#데이터소스-관리)
- [쿼리 실행](#쿼리-실행)
- [프로비저닝](#프로비저닝)
- [스트리밍](#스트리밍)

## 개요

Pharos 플러그인 시스템은 다양한 데이터소스와의 연동을 지원하며, 다음 두 가지 플러그인 아키텍처를 제공합니다:

- **Built-in 플러그인**: 바이너리에 내장되어 별도 프로세스 없이 실행
- **HashiCorp 플러그인**: 별도 프로세스로 실행되는 외부 플러그인 (go-plugin 사용)

### 주요 기능

- **동적 플러그인 로딩**: 실행 시 플러그인 자동 탐색 및 로드
- **JSON Schema 기반 검증**: React JSON Schema Form을 통한 동적 폼 생성 및 유효성 검증
- **다중 쿼리 실행**: 여러 데이터소스에 대한 동시 쿼리 실행
- **프로비저닝**: 설정 파일 기반 데이터소스 자동 구성
- **스트리밍 지원**: 실시간 데이터 스트리밍 (WebSocket 기반)

## 지원 플러그인

### 데이터 플러그인 (Data Client)
SQL 쿼리를 실행하고 결과를 반환하는 플러그인:

| 플러그인 | ID | 타입 | 설명 |
|---------|-----|------|------|
| ClickHouse | `clickhouse` | Built-in | ClickHouse 데이터베이스 |
| PostgreSQL | `postgresql` | Built-in | PostgreSQL 데이터베이스 |
| Altibase | `altibase` | Built-in | Altibase 데이터베이스 |
| Vertica | `vertica` | Built-in | Vertica 데이터베이스 |
| Prometheus | `prometheus` | Built-in | Prometheus 메트릭 쿼리 |
| Elasticsearch | `elasticsearch` | Built-in | Elasticsearch 검색 및 쿼리 (ES\|QL 지원) |

### 스트리밍 플러그인 (Stream Client)
실시간 데이터 수신 및 발행을 지원하는 플러그인:

| 플러그인 | ID | 타입 | 설명 |
|---------|-----|------|------|
| HTTP Receiver | `http-receiver` | Built-in | HTTP 엔드포인트로 데이터 수신 및 WebSocket으로 발행 |

## 플러그인 관리

### 사용 가능한 플러그인 목록 조회

플러그인 설정 정보를 포함하여 사용 가능한 모든 플러그인을 조회합니다.

**요청**
```http
GET /plugins
```

**설명**
- [React JSON Schema Form](https://rjsf-team.github.io/react-jsonschema-form/)을 통해 동적 폼 생성
- JSON Schema: 데이터소스 POST API의 요청 본문 구조 정의 및 유효성 검증
- UI Schema: 폼 렌더링 옵션 정의 (위젯 타입, 레이아웃 등)

**응답**
```json
[
  {
    "type": "datasource",
    "id": "clickhouse",
    "name": "Clickhouse",
    "settings": {
      "jsonSchema": {
        "type": "object",
        "required": [
          "type",
          "name", 
          "host",
          "port",
          "max_open_connection",
          "max_lifetime",
          "username",
          "password"
        ],
        "properties": {
          "type": {
            "type": "string",
            "title": "Type",
            "default": "clickhouse",
            "readOnly": true
          },
          "name": {
            "type": "string",
            "title": "Name",
            "default": "clickhouse-01"
          },
          "host": {
            "type": "string",
            "title": "Host",
            "default": "localhost"
          },
          "port": {
            "type": "integer",
            "title": "Port",
            "default": 8123
          },
          "max_open_connection": {
            "type": "integer",
            "title": "Max open connection",
            "default": 10
          },
          "max_lifetime": {
            "type": "integer",
            "title": "Max lifetime",
            "default": 14400
          },
          "username": {
            "type": "string",
            "title": "Username",
            "default": "default"
          },
          "password": {
            "type": "string", 
            "title": "Password",
            "default": "default"
          }
        }
      },
      "uiSchema": {
        "password": {
          "ui:widget": "password"
        }
      }
    }
  }
]
```

## 데이터소스 관리

### 데이터소스 목록 조회

모든 데이터소스(DB 저장 + 프로비저닝)를 조회합니다.

**요청**
```http
GET /plugins/datasources
```

**응답**
```json
[
  {
    "name": "clickhouse-01",
    "type": "clickhouse",
    "data": {
      "host": "localhost",
      "port": 8123,
      "max_open_connection": 10,
      "max_lifetime": 14400,
      "username": "default",
      "password": "default"
    },
    "provisioning": false
  }
]
```

**응답 필드**
- `provisioning`: `true`인 경우 프로비저닝 데이터소스 (수정/삭제 불가)

### 특정 데이터소스 조회

**요청**
```http
GET /plugins/datasources/:name
```

**파라미터**
- `name`: 데이터소스 이름

**응답**
```json
{
  "name": "clickhouse-01",
  "type": "clickhouse",
  "data": {
    "host": "localhost",
    "port": 8123,
    "max_open_connection": 10,
    "max_lifetime": 14400,
    "username": "default",
    "password": "default"
  },
  "provisioning": false
}
```

### 데이터소스 생성

새로운 데이터소스를 생성합니다.

**요청**
```http
POST /plugins/datasources
Content-Type: application/json
```
**요청 본문**
```json
{
  "type": "clickhouse",
  "name": "clickhouse-01", 
  "data": {
    "host": "localhost",
    "port": 8123,
    "max_open_connection": 10,
    "max_lifetime": 14400,
    "username": "default",
    "password": "default"
  }
}
```

**예제**
```bash
curl -X POST ${master-address}/plugins/datasources \
     -H 'Accept: application/json' \
     -H 'Content-Type: application/json' \
     -d '{
  "type": "clickhouse",
  "name": "clickhouse-01",
  "data": {
    "host": "localhost",
    "port": 8123,
    "max_open_connection": 10,
    "max_lifetime": 14400,
    "username": "default",
    "password": "default"
  }
}'
```

### 데이터소스 수정

기존 데이터소스 설정을 수정합니다.

**요청**
```http
PUT /plugins/datasources/:name
Content-Type: application/json
```

**파라미터**
- `name`: 수정할 데이터소스 이름 (프로비저닝 데이터소스는 수정 불가)

**요청 본문**
```json
{
  "type": "clickhouse",
  "data": {
    "host": "new-host",
    "port": 8123,
    "max_open_connection": 20,
    "max_lifetime": 14400,
    "username": "default",
    "password": "new-password"
  }
}
```

**참고**: `name` 필드는 URL 파라미터로 지정되며 변경되지 않습니다.

### 데이터소스 삭제

**요청**
```http
DELETE /plugins/datasources/:name
```

**파라미터**
- `name`: 삭제할 데이터소스 이름 (프로비저닝 데이터소스는 삭제 불가)

**응답**
```http
HTTP/1.1 200 OK
```

## 쿼리 실행

### 다중 데이터소스 쿼리

여러 데이터소스에 대해 동시에 쿼리를 실행합니다.

**요청**
```http
POST /ds/query
Content-Type: application/json
```
**요청 본문**
```json
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
      "datasourceName": "invalid",
      "sql": "SELECT * FROM pg_settings LIMIT 2"
    }
  ]
}
```

**예제**
```bash
curl -X POST ${master-address}/plugins/ds/query \
     -H 'Content-Type: application/json' \
     -d '{
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
      "datasourceName": "invalid",
      "sql": "SELECT * FROM pg_settings LIMIT 2"
    }
  ]
}'
```

**응답**
**응답**
```json
{
  "Results": {
    "A": {
      "Frame": {
        "meta": [
          {
            "name": "Timestamp",
          },
          {
            "name": "Host", 
          },
          {
            "name": "Database",
          },
          {
            "name": "Encoding",
          },
          {
            "name": "Collate",
          },
          {
            "name": "Ctype",
          },
          {
            "name": "ConnectionLimit",
          },
          {
            "name": "Size",
          }
        ],
        "data": [
          {
            "Collate": "en_US.UTF-8",
            "ConnectionLimit": -1,
            "Ctype": "en_US.UTF-8",
            "Database": "postgres",
            "Encoding": "UTF8",
            "Host": "lsson",
            "Size": 2396668387,
            "Timestamp": "2025-04-12T09:42:45Z"
          },
          {
            "Collate": "en_US.UTF-8",
            "ConnectionLimit": -1,
            "Ctype": "en_US.UTF-8",
            "Database": "dskim-dev",
            "Encoding": "UTF8",
            "Host": "lsson",
            "Size": 7754211,
            "Timestamp": "2025-04-12T09:42:45Z"
          }
        ],
        "rows": 2,
        "sql": "SELECT * FROM tarzan.database LIMIT 2",
        "statistics": {
          "elapsed": 0.105339966
        }
      },
      "Error": ""
    },
    "B": {
      "Frame": {
        "meta": [
          {
            "name": "name",
          },
          {
            "name": "setting",
          },
          {
            "name": "unit", 
          },
          {
            "name": "category",
          },
          {
            "name": "short_desc",
          },
          {
            "name": "extra_desc",
          },
          {
            "name": "context",
          },
          {
            "name": "vartype",
          },
          {
            "name": "source",
          },
          {
            "name": "min_val",
          },
          {
            "name": "max_val",
          },
          {
            "name": "enumvals",
          },
          {
            "name": "boot_val",
          },
          {
            "name": "reset_val",
          },
          {
            "name": "sourcefile",
          },
          {
            "name": "sourceline",
          },
          {
            "name": "pending_restart",
          }
        ],
        "data": [
          {
            "boot_val": "off",
            "category": "Developer Options",
            "context": "superuser",
            "enumvals": null,
            "extra_desc": null,
            "max_val": null,
            "min_val": null,
            "name": "allow_in_place_tablespaces",
            "pending_restart": false,
            "reset_val": "off",
            "setting": "off",
            "short_desc": "Allows tablespaces directly inside pg_tblspc, for testing.",
            "source": "default",
            "sourcefile": null,
            "sourceline": null,
            "unit": null,
            "vartype": "bool"
          },
          {
            "boot_val": "off",
            "category": "Developer Options",
            "context": "superuser",
            "enumvals": null,
            "extra_desc": null,
            "max_val": null,
            "min_val": null,
            "name": "allow_system_table_mods",
            "pending_restart": false,
            "reset_val": "off",
            "setting": "off",
            "short_desc": "Allows modifications of the structure of system tables.",
            "source": "default",
            "sourcefile": null,
            "sourceline": null,
            "unit": null,
            "vartype": "bool"
          }
        ],
        "rows": 2,
        "sql": "SELECT * FROM pg_settings LIMIT 2",
        "statistics": {
          "elapsed": 0.01580569
        }
      },
      "Error": ""
    },
    "C": {
      "Frame": {
        "meta": null,
        "data": null,
        "rows": 0,
        "sql": "",
        "statistics": {
          "elapsed": 0
        }
      },
      "Error": "not exist datasource"
    }
  }
}
```

### 응답 구조

각 쿼리 결과는 다음 구조를 가집니다:

- **Frame**: 쿼리 결과 데이터
  - **meta**: 컬럼 메타데이터 배열
    - **name**: 컬럼 이름
  - **data**: 실제 데이터 배열 (JSON 객체 배열)
  - **rows**: 반환된 행 수
  - **sql**: 실행된 SQL 쿼리
  - **statistics**: 실행 통계
    - **elapsed**: 쿼리 실행 시간 (초)
- **Error**: 오류 메시지 (성공 시 빈 문자열)

### 에러 처리

쿼리 실행 중 발생할 수 있는 에러:

| 에러 메시지 | 설명 |
|------------|------|
| `not exist datasource` | 존재하지 않는 데이터소스 |
| `not supported datasource type` | 지원하지 않는 데이터소스 타입 |
| `not supported data server` | Data Server를 지원하지 않는 플러그인 |
| `not supported` | 지원하지 않는 기능 |
| `not implemented` | 구현되지 않은 기능 |

**특징**:
- 각 쿼리는 독립적으로 처리됩니다
- 일부 쿼리 실패가 다른 쿼리에 영향을 주지 않습니다
- 같은 데이터소스에 대한 쿼리는 자동으로 배치 처리됩니다

## 프로비저닝

설정 파일을 통해 데이터소스를 자동으로 구성할 수 있습니다.

### 설정 방법

**config.yaml**
```yaml
provisioning:
  plugins:
    datasources:
      - type: clickhouse
        name: clickhouse-prod
        data:
          host: prod-server
          port: 8123
          max_open_connection: 20
          max_lifetime: 14400
          username: admin
          password: secure-password
      - type: postgresql
        name: postgres-prod
        data:
          host: postgres-server
          port: 5432
          max_open_connection: 10
          max_lifetime: 3600
          username: postgres
          password: postgres-password
```

### 샘플 데이터소스

개발 및 테스트를 위해 모든 플러그인 타입의 샘플 데이터소스를 자동 생성할 수 있습니다:

**config.yaml**
```yaml
plugins:
  sample: true
```

샘플 데이터소스는 각 플러그인의 `schema/sample-form-data.json`에서 로드됩니다.

### 프로비저닝 데이터소스 특징

- **읽기 전용**: 수정 및 삭제 불가
- **우선순위**: DB에 저장된 데이터소스보다 우선 적용
- **충돌 방지**: DB에 동일한 이름이 있으면 에러 로그 출력 후 스킵

## 스트리밍

특정 플러그인은 실시간 데이터 스트리밍을 지원합니다.

### HTTP Receiver 플러그인

HTTP 엔드포인트로 데이터를 수신하고 WebSocket을 통해 구독자에게 발행합니다.

**데이터소스 생성**
```bash
curl -X POST ${master-address}/plugins/datasources \
     -H 'Content-Type: application/json' \
     -d '{
  "type": "http-receiver",
  "name": "metrics-receiver",
  "data": {
    "host": "0.0.0.0",
    "port": 9090,
    "path": "/metrics"
  }
}'
```

**데이터 전송**
```bash
curl -X POST http://localhost:9090/metrics \
     -H 'Content-Type: application/json' \
     -d '{
  "timestamp": "2024-12-02T10:00:00Z",
  "value": 42.5,
  "labels": {
    "host": "server-01",
    "metric": "cpu_usage"
  }
}'
```

**WebSocket 구독**
```javascript
const ws = new WebSocket('ws://master-address/ws');
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);
};
```

### 스트리밍 API

플러그인에서 구현 가능한 Stream 인터페이스:

- `RunStream`: 스트리밍 시작 및 설정
- `SubscribeStream`: 클라이언트 구독
- `UnsubscribeStream`: 클라이언트 구독 해제
- `PublishStream`: 데이터 발행
- `RemoveDatasource`: 스트리밍 정리

## 플러그인 개발

### Built-in 플러그인 구조

```
plugin/datasource/mydb/
├── mydb.go              # 플러그인 초기화
├── schema/
│   ├── json.json        # JSON Schema
│   ├── ui.json          # UI Schema
│   └── sample-form-data.json  # 샘플 데이터
└── main/
    └── main.go          # (선택) HashiCorp 플러그인용
```

**mydb.go 예시**
```go
package mydb

import (
    _ "embed"
    "ntels.com/pharos/core/pkg/common"
    "ntels.com/pharos/core/pkg/plugins/internal/plugin/datasource"
    "ntels.com/pharos/core/pkg/plugins/internal/plugin/datasource/sql"
    "ntels.com/pharos/core/pkg/plugins/model"
)

//go:embed schema/json.json
var jsonSchema string

//go:embed schema/ui.json
var uiSchema string

//go:embed schema/sample-form-data.json
var sampleFormData string

func GetPlugin(config common.Config) (*model.Plugin, error) {
    datasourceClient := &model.DatasourceClient{
        DataClient: datasource.GetDataClient(
            datasource.PluginLocationBuiltIn, 
            &sql.DataClient{Config: config}, 
            config,
        ),
    }
    
    return model.MakePlugin(
        model.PluginTypeDatasource, 
        "mydb", 
        "My Database", 
        jsonSchema, 
        uiSchema, 
        datasourceClient, 
        sampleFormData,
    )
}
```

### HashiCorp 플러그인

외부 바이너리로 실행되는 플러그인:

1. `plugins.path` 디렉토리에 실행 파일 배치
2. 안전성 검증:
   - 파일명이 안전한 패턴인지 확인
   - 심볼릭 링크 해석 및 경로 검증
   - 실행 권한 확인
3. go-plugin을 통한 RPC 통신

**플러그인 로딩 순서**:
1. Built-in 플러그인 로드
2. HashiCorp 플러그인 스캔 및 로드
3. 프로비저닝 데이터소스 로드
4. 스트리밍 시작

## 에러 코드

| 에러 | 설명 |
|------|------|
| `ErrorNotExistDatasource` | 데이터소스가 존재하지 않음 |
| `ErrorNotSupportedDatasourceType` | 지원하지 않는 데이터소스 타입 |
| `ErrorNotSupportedDataServer` | Data Server 미지원 |
| `ErrorNotSupportedStreamServer` | Stream Server 미지원 |
| `ErrorInvalidDatasourceType` | 유효하지 않은 데이터소스 타입 |
| `ErrorNotSupported` | 지원하지 않는 기능 |
| `ErrorNotImplemented` | 구현되지 않음 |
| `ErrorNotExistPluginJSONSchema` | JSON Schema가 없음 |

## 참고 자료

- [React JSON Schema Form](https://rjsf-team.github.io/react-jsonschema-form/)
- [JSON Schema Specification](https://json-schema.org/)
- [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin)
- [Centrifuge (WebSocket)](https://github.com/centrifugal/centrifuge)
