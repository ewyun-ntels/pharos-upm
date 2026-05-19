# Badges API (pkg/badges)

UI에서 배지를 쉽게 테스트/개발할 수 있도록 Badges API를 정리했습니다. 배지는 사용자가 정의한 쿼리를 등록(POST/PUT)해두고, 조회(GET) 시에는 이름만으로 저장된 쿼리를 실행하여 단일 값을
반환합니다. Dashboard와 비슷한 동작 모델을 따르지만, GET에서는 권한 문제로 쿼리를 직접 받지 않습니다.

## 핵심 개념

- 저장 엔드포인트(POST/PUT): 배지 이름(name), 데이터소스(datasource), 쿼리(query), 값 추출 컬럼(column)을 등록/업데이트합니다.
- 조회 엔드포인트(GET /badges/:name): 저장된 쿼리를 실행해 첫 번째 행의 column에 해당하는 값을 추출하여 반환합니다. 내부적으로 결과를 캐싱하여 동일 배지명에 대한 반복 조회 시 데이터소스
  쿼리를 줄입니다(POST/PUT/DELETE 시 해당 배지 캐시 무효화). 캐시 TTL은 설정 파일의 [badges].cache_ttl로 제어하며, 0으로 설정하면 캐시를 사용하지 않습니다. 응답에 cache_hit 필드를 포함하여 캐시 사용 여부를 확인할 수 있습니다.
- 목록(GET /badges): 등록된 배지 이름 목록을 반환합니다.
- 삭제(DELETE /badges/:name): 배지를 삭제합니다.
- 등록되지 않은 배지 조회: 배지가 존재하지 않을 경우 404 대신 204 No Content를 반환하여 클라이언트에서 에러 로그 없이 처리할 수 있도록 합니다.

## 엔드포인트 요약

- POST /badges (권한: SuperAdmin)
- PUT /badges/:name (권한: SuperAdmin)
- GET /badges (권한: SuperAdmin)
- GET /badges/:name (권한: 공개)
- DELETE /badges/:name (권한: SuperAdmin)

권한은 서버의 인증/인가 설정에 따라 달라질 수 있습니다. 현재 구현은 router.go 참고.

## 요청/응답 스키마

POST /badges, PUT /badges/:name

- 요청(JSON: BadgeSpec)
  - PUT의 경우: 경로 파라미터 :name과 요청 본문의 name이 둘 다 존재하면 동일해야 합니다. 본문 name이 비어 있으면 경로의 :name이 적용됩니다.
  ```json5
  {
    "name": "uptime", // 배지 이름 (필수)
    "datasource": "ds1", // 데이터소스 이름 (필수)
    "query": {
      "query": "select 1 as v", // 실행할 SQL 템플릿 (필수)
      "variables": { // 템플릿 변수 (선택)
        // 예: "__start_time": "2025-01-01T00:00:00Z"
        // 예: "__end_time": "2025-01-01T00:00:00Z"
      },
      "time_label": "", // 선택
      "variable_label": ""           // 선택
    },
    "column": "v"                    // 반환에 사용할 컬럼명(최상위) (필수)
  }
  ```
- 응답: 200 OK (본문 없음)
- 에러: 400 Bad Request (필수 필드 누락), 500 Internal Server Error (DB/실행 오류)

GET /badges/:name

- 응답(JSON: BadgeValueResponse)
  ```json5
  {
    "name": "uptime",
    "value": 123,
    "last_updated": "2025-01-15T10:30:00Z",  // 응답 생성 시각
    "cache_hit": false                        // 캐시에서 가져왔는지 여부 (true: 캐시, false: DB 쿼리)
  }
  ```
- 반환 규칙
    - 저장된 쿼리를 렌더링(query_builder) 후 plugins.DatasourceQuery로 실행합니다.
    - 첫 번째 행의 지정한 column 키의 값을 추출합니다. 행이 없거나 키가 없으면 value=null로 반환합니다.
    - cache_hit 필드를 통해 캐시에서 가져온 값인지 확인할 수 있습니다.
- 에러: 204 No Content (배지 미등록), 500 Internal Server Error (쿼리 실행 오류 등)

GET /badges

- 응답(JSON): 
  ```json5
  ["uptime", "..."]
  ```

DELETE /badges/:name

- 응답: 200 OK

## 예제(curl)

1) 배지 등록(또는 업데이트)

```bash
curl -X POST http://localhost:8080/badges \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <TOKEN>' \
  -d '{
    "name": "uptime",
    "datasource": "ds1",
    "query": {"query": "select 1 as v"},
    "column": "v"
  }'
```

2) 배지 수정(경로에 이름)

```bash
curl -X PUT http://localhost:8080/badges/uptime \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <TOKEN>' \
  -d '{
    "datasource": "ds1",
    "query": {"query": "select 2 as v"},
    "column": "v"
  }'
```

3) 배지 조회(이름으로 실행)

```bash
curl http://localhost:8080/badges/uptime
# -> {"name":"uptime","value":2,"last_updated":"2025-01-15T10:30:00Z","cache_hit":false}
```

4) 배지 목록

```bash
curl -H 'Authorization: Bearer <TOKEN>' http://localhost:8080/badges
# -> ["uptime", ...]
```

5) 배지 삭제

```bash
curl -X DELETE -H 'Authorization: Bearer <TOKEN>' http://localhost:8080/badges/uptime
```

## 템플릿/변수 사용

- query_builder를 통해 SQL 템플릿을 렌더링합니다. variables 객체에 제공한 키/값을 사용하여 문자열을 생성합니다.
- SQL 인젝션 방지를 위해 변수는 가능한 안전하게 다루고, 서버측 플러그인/드라이버의 바인딩 옵션이 있다면 활용하세요.

## 값 추출 규칙

- 쿼리 결과(plugins.DatasourceQuery 응답)의 첫 번째 행을 확인합니다.
- 행의 맵에서 column 키를 찾아 해당 값을 value로 반환합니다.
- 행이 없거나 키가 없다면 value=null로 반환합니다(200 OK).

## 자주 발생하는 오류와 해결

- 400 name/datasource/query/column is empty: 필수 필드 누락. 요청 JSON을 확인하세요.
- 204 No Content: 등록되지 않은 배지를 조회했습니다. 먼저 POST/PUT으로 등록하세요.
- 500 datasource query missing result: plugins.DatasourceQuery 응답에 지정 ID가 없습니다. 데이터소스 설정/플러그인을 확인하세요.
- 500 DatasourceQuery error: 실제 쿼리 실행 에러. SQL과 데이터소스 연결 정보를 확인하세요.

## 로컬 테스트 가이드(UI 개발자)

- 서버를 개발 모드로 실행 후 위의 curl 예제로 동작을 확인할 수 있습니다.
- 단위 테스트: 백엔드만으로 빠르게 확인하려면 다음을 사용하세요.
    - 특정 패키지 테스트: `go test ./pkg/badges -v`
    - 전체 테스트: `go test ./...`
- pkg/badges/handler_test.go는 실제 데이터소스 없이도 테스트할 수 있도록 내부적으로 dsQuery를 스텁하여 동작을 검증합니다.

## 내부 구현 참고

- 라우터: pkg/badges/router.go
- 핸들러: pkg/badges/handler.go
- 모델/테이블: pkg/badges/model.go (badges 테이블: name, datasource, query_json, column, updated_at)
- 리소스 타입: pkg/badges/resource.go

UI에서 필요한 최소 입력(이름, 쿼리, 컬럼)만으로 쉽게 배지 값을 제공할 수 있도록 설계되었습니다. 문제가 있으면 핸들러 로그 메시지와 응답 메시지를 참고해 원인을 파악하세요.
