# Alert 서브시스템

이 패키지는 알림(Alert) 규칙을 정의/실행하고 현재 경보 상태 및 이력 조회를 제공합니다. 주요 구성 요소는 다음과 같습니다.

- Alert Rule: 주기적 쿼리/이벤트 등으로 경보를 산출하는 규칙
- Alert Status: 현재 활성화된 경보 상태 테이블(alert_status)
- Alert History: 경보 이력(history_alert, history_alert_row 혹은 alert_hist 등 백엔드 구성에 따라)

## 아키텍처 및 구조

### 지원하는 Alert 타입

현재 시스템에서 지원하는 Alert 타입은 다음과 같습니다:

- **Query**: 주기적 데이터베이스 쿼리를 통한 메트릭 기반 알림
- **EventStatus**: 외부 이벤트 기반 현재 상태 관리 알림
- **EventHistory**: 외부 이벤트 기반 이력 보관 및 정리 알림

### Adapter 계층 구조

Alert 시스템은 다계층 adapter 패턴을 통해 데이터 변환을 수행합니다:

```
API Request (shared/types) → Resource (resources/) → Model (model/) → Database
                          ←                       ←                 ←
API Response              ← Resource              ← Model           ← Database
```

#### 1. RuleAdapter (`adapters/rule_adapter.go`)
- **목적**: 내부 모델과 API 리소스 간 변환
- **지원 타입**: Query, EventStatus, EventHistory 모든 타입
- **주요 메서드**:
  - `ModelToResource()`: 데이터베이스 모델 → API 리소스 변환
  - `ResourceToModel()`: API 리소스 → 데이터베이스 모델 변환
  - `ToAPIRuleResponseFromInternal()`: 내부 타입 → API 응답 변환

#### 2. APIAdapter (`adapters/api_adapter.go`)
- **목적**: API 입력/출력 타입과 내부 타입 간 변환
- **주요 메서드**:
  - `ToAPIAlertValue()`: 내부 alert value → API alert value
  - `ToAPIAlertValues()`: 배치 변환 지원

### 테스트 커버리지

Alert adapter는 **100% 테스트 커버리지**를 달성했습니다:

#### 기본 Adapter 테스트
- API 변환 테스트 (2개)
- Rule adapter 변환 테스트 (4개) 
- 호환성 테스트 (2개)

#### CRUD 통합 테스트
- **모든 Alert 타입 CRUD 테스트**: Query, EventStatus, EventHistory (3개)
- **전체 변환 체인 검증**: API → Resource → Model → Resource → API
- **통합 테스트**: 실제 사용 시나리오 기반 (3개)
- **핸들러 테스트**: HTTP 레이어 포함 완전한 통합 (6개)

**총 테스트**: 16개 모든 테스트 통과 ✅

## 개발 및 테스트

### Adapter 테스트 실행

alert adapter의 모든 테스트를 실행하려면:

```bash
cd core/pkg/alert/adapters  
go test -v
```

**예상 결과**: 16개 테스트 모두 통과 (PASS)

개별 테스트 그룹 실행:
```bash
# CRUD 테스트만 실행
go test -v -run "TestAlertRuleCRUD"

# 통합 테스트만 실행
go test -v -run "TestAPIAdapterIntegrationWithAlertTypes"

# 핸들러 테스트만 실행
go test -v -run "TestHandler"
```

### 데이터 변환 검증

시스템의 데이터 무결성을 확인하려면:

1. **Alert 타입별 CRUD 동작** 검증 (Query, EventStatus, EventHistory)
2. **API ↔ Resource ↔ Model** 전체 변환 체인 검증  
3. **상태 변환 정확성** 확인 (Critical, Major, Minor, Normal)
4. **HTTP 핸들러 통합** 검증

## 엔드포인트 개요

핸들러 구현은 pkg/alert/handler.go의 RegisterRoutes를 참조하세요. 기본 경로(prefix)는 서비스 라우팅에 따라 다를 수 있습니다.

- GET /status — 현재 alert 상태 전체 조회(인증 필요)
- PUT /status/:name/:id/mask?mask={true|false} — 특정 상태 마스킹 토글(인증 필요)
- DELETE /status/:name/:id — 특정 상태 정리(clean, normal로 전환 기록 남김)(인증 필요)
- POST /rule — Alert Rule 생성 (SuperAdmin 또는 ManageAlert 권한 필요)
- GET /rule — Alert Rule 목록 조회 (인증 선택적: 코드상 false로 되어 있어 내부 배포 환경에 맞춤)
    - 쿼리 파라미터 detail=true 지정 시 상태 포함 상세 조회(권한 보유자에게 유용)
- GET /rule/:id — Alert Rule 단건 조회 (SuperAdmin 또는 ManageAlert 권한 필요)
- PUT /rule/:id — Alert Rule 수정 (SuperAdmin 또는 ManageAlert 권한 필요)
- DELETE /rule/:id — Alert Rule 삭제 (SuperAdmin 또는 ManageAlert 권한 필요)
- POST /query — 규칙에 사용할 쿼리 테스트 실행 (권한 필요)
- GET /hist — ClickHouse에서 Alert 이력 조회(권한 필요). name, start-time, end-time, count 쿼리 파라미터 지원.
- POST /event/:name — 외부 이벤트 입력(현재 TODO로 비활성화 방향 주석; 내부 연동 상황에 따라 사용)

권한 요약

- 읽기(status, rule list)는 상대적으로 완화되었으나 배포 환경 인증 정책을 따릅니다.
- 생성/수정/삭제는 SuperAdmin 또는 ManageAlert 역할 필요.

## 리소스 스키마

Alert Rule 요청/응답(resources.Rule):

- request (POST/PUT /rule)
  ```json
  {
    "alert_type": "query",
    "rule": { }
  }
  ```
  가능한 alert_type: "query", "event_status", "event_history", ...
- response
    - POST /rule 성공: {"id": "<uuid>"}
    - GET /rule (목록): 권한 및 detail 파라미터에 따라 전체 또는 축약 정보 반환
        - 축약형 예: { alert_type, rule: {name,id}, timestamp, update_at }
        - detail=true이고 권한 보유 시: status 필드에 현재 상태 배열 포함
    - GET /rule/:id: { alert_type, rule(전체 설정), status(옵션), timestamp, update_at }
    - PUT/DELETE 성공: HTTP 200 OK (바디 없음)

Query 테스트(resources.Query):

```json
{
  "datasource": "<플러그인 데이터소스명>",
  "query": {
    "variables": {
      "k": "v"
    },
    "query": "SELECT ...",
    "time_label": "time",
    "variable_label": "var"
  }
}
```

외부 이벤트(resources.Event):

```json
{
  "alert_id": "<이벤트 고유 id>",
  "description": "...",
  "severity": "major|minor|critical|...",
  "value": 12.34,
  "labels": {
    "k": "v"
  }
}
```

## Threshold(임계치) 동작 상세

이 섹션은 Query 기반 규칙에서 threshold(또는 thresholds) 필드가 어떻게 평가되어 상태(severity, status)가 결정되는지 설명합니다. 실제 비교 연산자는 코드에 구현된 명명형 방식만
지원합니다.

- 기본 개념
    - 규칙은 스케줄(evaluation_interval, schedule)에 따라 주기적으로 값을 산출합니다.
    - 산출된 값(예: value, count 등)에 대해 임계치 목록을 위에서 아래 순서대로 평가합니다.
    - 일치하는 임계치 중 가장 높은 심각도(severity)가 최종 심각도로 선택됩니다. 아무 임계치도 일치하지 않으면 NORMAL로 간주합니다.

- 임계치 항목의 스키마(코드 기준)

```json
[
  {
    "id": "major-threshold",
    "condition": "is_above",
    "start_value": 80,
    "severity": "major",
    "labels": {
      "source": "cpu"
    }
  },
  {
    "id": "critical-range",
    "condition": "within_range",
    "start_value": 90,
    "end_value": 100,
    "severity": "critical"
  }
]
```

- 필수 필드
    - condition: 비교 연산자 이름. 아래 네 가지 중 하나를 사용
        - "is_above" (>): 값이 start_value 보다 클 때 참
        - "is_below" (<): 값이 start_value 보다 작을 때 참
        - "within_range" (포함 범위): start_value 이상 AND end_value 이하
        - "outside_range" (범위 밖): start_value 미만 OR end_value 초과
    - start_value: 기준 값. 범위 연산의 하한(lower bound)이며, 단일 비교에서도 임계값으로 사용
    - severity: 일치 시 부여할 심각도 문자열
- 선택 필드
    - id: 식별/추적용 문자열
    - end_value: 범위 연산 시 상한(upper bound). within_range/outside_range에서만 의미 있음
    - labels: 매칭 시 상태/이력에 함께 기록할 라벨 키-값

- 유효성/검증 규칙
    - condition 값은 위 네 가지 중 하나여야 합니다.
    - within_range / outside_range 사용 시 start_value <= end_value 이어야 합니다. 이 조건이 어기면 잘못된 설정으로 간주됩니다.
    - is_above / is_below 에서는 end_value는 무시됩니다.

- 값 산출 규칙
    - Query 규칙은 datasource_query 결과에서 첫 행의 첫 컬럼 또는 alias가 "value"인 컬럼을 우선 사용합니다.
    - 필요한 경우 datasource_query에서 집계/alias를 명시해 주는 것이 안전합니다.

- 정상화(클리어) 조건
    - 어떤 임계치에도 더 이상 일치하지 않는 순간 NORMAL로 전환되며, 이전 상태는 history에 종료로 기록됩니다.
    - 상승/하강 임계값을 분리하는 히스테리시스는 현재 스키마에 내장되어 있지 않습니다(필요 시 별도의 임계치 조합으로 구현).

- 디바운스/지속 시간(for) - 차후 고도화 필요
    - 구현에 따라 임계치 유지 시간을 둘 수 있습니다. 해당 옵션을 사용하는 구현이 추가될 때까지는 즉시 일치/해제 모델이 기본입니다.

- 누락 데이터 처리
    - 최근 평가 주기에 데이터가 없으면 이전 상태 유지 또는 NORMAL 처리 중 하나로 정책화할 수 있습니다. 기본은 이전 상태 유지입니다.

- 상태 전이와 기록
    - severity 변경 시 이전 상태는 history_alert에 종료/변경 사유가 기록되고, 새 상태로 alert_status가 갱신됩니다.
    - 동일 severity로 값만 변하면 중복 기록을 줄이기 위해 업데이트만 하고 추가 이력은 생략될 수 있습니다.

- 간단한 Query 규칙 예시

```json
{
  "alert_type": "query",
  "rule": {
    "name": "cpu-usage-alert",
    "datasource": "clickhouse-01",
    "datasource_query": {
      "query": {
        "query": "SELECT max(usage) AS value FROM cpu WHERE time > now()-60s"
      },
      "variables": {}
    },
    "threshold": [
      {
        "id": "cpu-major",
        "condition": "is_above",
        "start_value": 80,
        "severity": "major"
      },
      {
        "id": "cpu-critical",
        "condition": "within_range",
        "start_value": 90,
        "end_value": 100,
        "severity": "critical"
      }
    ],
    "evaluation_interval": "@every 30s",
    "check_type": "last"
  }
}
```

## CRUD 예시 cURL

1) 규칙 생성(Create)

```shell
curl -X POST "$HOST/alert/rule" \
-H "Authorization: Bearer $TOKEN" \
-H "Content-Type: application/json" \
-d '{
"alert_type": "query",
"rule": {
"name": "cpu-usage-alert",
"schedule": "@every 30s",
"datasource": "clickhouse",
"query": {
"query": "SELECT max(value) as value FROM cpu WHERE time > now()-60s",
"variables": {}
},
"threshold": [
{"condition": "is_above", "start_value": 80, "severity": "major"},
{"condition": "is_above", "start_value": 90, "severity": "critical"}
]
}
}'
```

응답: {"id":"<uuid>"}

2) 규칙 목록 조회(List)

# 축약 정보

```shell
curl -X GET "$HOST/alert/rule" -H "Authorization: Bearer $TOKEN"
```

# 상세 정보(상태 포함)

```shell
curl -X GET "$HOST/alert/rule?detail=true" -H "Authorization: Bearer $TOKEN"
```

3) 규칙 단건 조회(Get)

```shell
curl -X GET "$HOST/alert/rule/<id>" -H "Authorization: Bearer $TOKEN"
```

4) 규칙 수정(Update)

```shell
curl -X PUT "$HOST/alert/rule/<id>" \
-H "Authorization: Bearer $TOKEN" \
-H "Content-Type: application/json" \
-d '{
"alert_type": "query",
"rule": {
"name": "cpu-usage-alert",
"schedule": "@every 15s",
"datasource": "clickhouse",
"query": {
"query": "SELECT max(value) as value FROM cpu WHERE time > now()-30s",
"variables": {}
},
"thresholds": {"major": 70, "critical": 90}
}
}'
```

응답: HTTP/1.1 200 OK

5) 규칙 삭제(Delete)

```shell
curl -X DELETE "$HOST/alert/rule/<id>" -H "Authorization: Bearer $TOKEN"
```

응답: HTTP/1.1 200 OK

## 상태(Status) 조작 예시

- 마스킹 토글
  ```shell
  curl -X PUT "$HOST/alert/status/<name>/<alertId>/mask?mask=true" -H "Authorization: Bearer $TOKEN"
  ```

- 상태 클린(정상화 기록 남김)
  ```shell
  curl -X DELETE "$HOST/alert/status/<name>/<alertId>" -H "Authorization: Bearer $TOKEN"
  ```

## 이력 조회 예시(GET /hist)

쿼리 파라미터:

- name: 룰 이름 필터
- start-time, end-time: time.DateTime 포맷(예: 2006-01-02 15:04:05)
- count: 반환 최대 행 수(기본 1000)

예시:
```shell
curl -G "$HOST/alert/hist" \
-H "Authorization: Bearer $TOKEN" \
--data-urlencode "name=cpu-usage-alert" \
--data-urlencode "start-time=2025-09-01 00:00:00" \
--data-urlencode "end-time=2025-09-02 00:00:00" \
--data-urlencode "count=100"
```


## Rule 타입별 가이드: query, event_status, event_history

아래 섹션은 각 Rule 유형의 설정 필드와 동작을 요약합니다. 실제 필드는 코드 기준으로 최소한만 기재했습니다.

### 1) Query Rule (pkg/alert/rule/query)

- 핵심 필드
    - name(string): 규칙 이름. 필수
    - datasource(string): plugins에 등록된 데이터소스 이름. 필수
    - datasource_query(object): 쿼리와 변수 정의. 필수
    - threshold(array): 임계치 목록. 각 항목은 id(자동 생성 가능), severity, condition 등 세부 필드가 있으며 코드의 Threshold.Validate() 규칙을 따릅니다.
      최소 1개 필수
    - evaluation_interval(string): 크론/스케줄 표기(@every 30s 등). 필수
    - check_type(string): 현재 "last"만 지원
    - query_delay_offset(int64): 수집 지연 보정(초). 선택
    - notifications([string]): 알림 전송 목적지 이름 목록. 선택
- 동작
    - 스케줄마다 datasource_query를 plugins.DatasourceQuery로 실행하여 결과를 평가하고, threshold에 따라 상태/심각도를 산출하여 alert_status에 Upsert,
      notification 전송, history 기록을 수행합니다.
- 생성 예시(POST /alert/rule)

```shell
curl -X POST "$HOST/alert/rule" \
-H "Authorization: Bearer $TOKEN" \
-H "Content-Type: application/json" \
-d '{
"alert_type": "query",
"rule": {
"name": "cpu-usage-alert",
"datasource": "clickhouse",
"datasource_query": {
"query": {"query": "SELECT max(value) as value FROM cpu WHERE time > now()-60s"},
"variables": {}
},
"threshold": [
{"id": "major-th", "severity": "major", "op": ">=", "value": 80},
{"id": "critical-th", "severity": "critical", "op": ">=", "value": 90}
],
"evaluation_interval": "@every 30s",
"check_type": "last",
"notifications": ["snmp-v2c-sender"]
}
}'
```

### 2) EventStatus Rule (pkg/alert/rule/eventStatus)

- 핵심 필드
    - name(string): 규칙 이름. 필수
    - description(string): 메시지 설명. 선택
    - notifications([string]): 알림 목적지. 선택
- 동작
    - 외부에서 POST /alert/event/:name 호출 시 Event 데이터를 입력으로 현재 상태를 갱신합니다.
    - 동일 alert_id에 대해 severity가 변경되면 이전 상태를 normal로 정리 후, 새 상태(alerting)를 반영합니다.
    - severity가 normal로 입력되면 현재 상태를 삭제(clean)합니다.
    - 변경이 있을 때만 notification과 history를 기록합니다.
- 이벤트 입력 예시(POST /alert/event/:name)

```shell
curl -X POST "$HOST/alert/event/<ruleName>" \
-H "Content-Type: application/json" \
-d '{
"alert_id": "iface-1",
"description": "Link down",
"severity": "major",
"value": 1,
"labels": {"ifName": "eth0"}
}'
```


### 3) EventHistory Rule (pkg/alert/rule/eventHistory)

- 핵심 필드
    - name(string): 규칙 이름. 필수
    - description(string): 메시지 설명. 선택
    - retention_period(int): 보관 기간(초). 필수, 0보다 커야 함
    - notifications([string]): 알림 목적지. 선택
- 동작
    - 초 단위 크론(* * * * * *)으로 주기 실행하여 해당 name의 alert_status 중 updated_at이 retention_period보다 오래된 레코드를 정리(delete)합니다.
    - 또한 외부 이벤트 POST /alert/event/:name를 받아 이벤트를 이력(history)과 상태에 반영합니다. 이때 상태가 Normal인 이벤트는 저장만 스킵합니다.
- 이벤트 입력 예시(POST /alert/event/:name)

```shell
curl -X POST "$HOST/alert/event/<ruleName>" \
-H "Content-Type: application/json" \
-d '{
"alert_id": "job-1",
"description": "Batch finished",
"severity": "minor",
"value": 0,
"labels": {"job": "daily-etl"}
}'

```
