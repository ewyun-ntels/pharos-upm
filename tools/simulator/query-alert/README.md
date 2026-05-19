# Alert Query Rule 시뮬레이터

## 개요
- REST API를 통해 Query alert rule을 등록합니다.
- 지정된 기간(기본 10분) 동안 규칙 평가 간격(기본 10초)으로 실행됩니다.
- ClickHouse 쿼리를 사용하여 10초마다 숫자 값을 교대로 변경하여 alert 발생/해제 주기를 만듭니다.
- arrayJoin과 여러 라벨 값(예: 여러 호스트)을 사용하여 평가당 여러 행을 생성하므로 라벨별 다중 alert 동작을 검증할 수 있습니다.
- 주기적으로 /alert/status를 폴링하여 라벨당 예상 활성 alert를 검증하고, 불일치가 감지되면 실패합니다.
- 종료 시 생성된 규칙을 삭제하여 정리합니다.

## 요구사항
- Tarzan 서버가 실행 중이고 접근 가능해야 합니다(기본 http://127.0.0.1:8080).
- master_config.toml에 제공된 인스턴스에 ClickHouse가 구성되어 있어야 합니다.
- OAuth 클라이언트 자격 증명이 사용 가능해야 합니다. 기본적으로 config의 client-id=master와 client-secret=master_client_secret을 사용합니다.

## 설정
- config.toml을 통해 설정을 제공합니다. 시뮬레이터는 이 파일만 읽습니다.
- -config 플래그만 지원되며 다른 config 파일을 가리킬 수 있습니다. 모든 옵션은 TOML에서 설정해야 합니다.

## 설정 파일 예시 (config.toml):

  [server]
  base = "http://127.0.0.1:8080"
  client_id = "test"         # or "master"
  client_secret = ""          # empty allowed for password grant public client
  username = "admin"
  password = "admin"

  [clickhouse]
  database = "tarzan"
  schema = "http"
  host = "192.168.15.101"
  http_port = 31478
  username = "default"
  password = "B8i4G17XrC"

  [simulator]
  rule_name = "sim-query"
  duration = "1m"
  interval = 10
  poll = "10s"
  label_key = "host"
  labels = ["sim-a", "sim-b", "sim-c"]
  datasource = "clickhouse-01"

## 사용법
```bash
# 컴파일
go build -o alert-simulator main.go

# 실행
./alert-simulator -config ./config.toml

# 또는 직접 실행
go run main.go -config ./config.toml
```

## 인증 동작
- config.toml의 [server] 값으로 OAuth Resource Owner Password Credentials grant (grant_type=password)를 사용합니다.

## 플래그
- `-config`: 시뮬레이터 TOML 파일의 경로. 기본값: `./config.toml`

## 규칙의 동작
- **데이터소스**: clickhouse (서버 config를 사용하여 연결)
- **쿼리** (단순화):
  ```sql
  SELECT ts,
         toString(CASE mixed condition using idx parity and time modulo 20
                  WHEN active THEN 80 + (idx % 20) ELSE 1 + (idx % 5) END) AS val,
         label AS <label_key>
  FROM (
    SELECT arrayJoin(labels) AS label, arrayEnumerate(labels) AS idx
  )
  ```
  여기서 labels는 [simulator.labels]에서 구축됩니다 (예: ['sim-a','sim-b']).
  이는 매 평가마다 혼합된 상태를 생성합니다: 대략 절반의 라벨은 임계값 이상이고 절반은 이하입니다. 활성 값은 라벨마다 다릅니다.
- **임계값**:
  condition=is_above, start_value=80, severity=critical

## 검증
- 각 폴링에서 동일한 idx parity 규칙과 현재 시간을 사용하여 활성화되어야 할 라벨을 계산합니다.
- 예상된 라벨만 활성화되고(각각 정확히 한 번씩), 다른 것들은 비활성화되어 있는지 확인합니다.
- 여러 개가 활성화된 경우 활성 값이 모두 동일하지 않은지 확인합니다.
- 어설션이 통과하면 폴링마다 성공(OK)을 로그하고, 그렇지 않으면 어설션 실패를 로그합니다. 불일치가 관찰되면 0이 아닌 코드로 종료합니다. 생성된 규칙은 삭제됩니다.
