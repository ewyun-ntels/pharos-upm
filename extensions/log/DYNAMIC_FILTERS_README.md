# 동적 필터 시스템 가이드

## 개요
log extension은 이제 설정 파일 기반의 동적 필터 시스템을 지원합니다. 새로운 인덱스를 추가하고 싶을 때 코드 수정 없이 설정 파일만 편집하면 됩니다.

## 설정 방법

### 1. master_config.toml에 인덱스 스키마 추가

```toml
[log.opensearch]
address = "https://opensearch.example.com"
username = "admin"
password = "password"
insecure_skip_verify = true
indices = ["app-*", "node-*", "nginx-*"]  # 사용 가능한 인덱스 목록

# app-* 인덱스: Kubernetes 애플리케이션 로그
[[log.opensearch.index_schema]]
pattern = "app-*"
filters = [
  { name = "namespace", field = "kubernetes.namespace_name.keyword", display = "Namespace" },
  { name = "pod",       field = "kubernetes.pod_name.keyword",       display = "Pod" },
  { name = "level",     field = "level.keyword",                    display = "Level", case_insensitive = true }
]

# node-* 인덱스: 시스템 로그 (systemd journal)
[[log.opensearch.index_schema]]
pattern = "node-*"
filters = [
  { name = "hostname", field = "_HOSTNAME.keyword",      display = "Hostname" },
  { name = "unit",     field = "_SYSTEMD_UNIT.keyword",  display = "Unit" },
  { name = "priority", field = "PRIORITY",              display = "Priority" }
]

# nginx-* 인덱스 추가 예시
[[log.opensearch.index_schema]]
pattern = "nginx-*"
filters = [
  { name = "host",        field = "host.keyword",         display = "Host" },
  { name = "status_code", field = "status_code",          display = "Status" },
  { name = "method",      field = "request_method.keyword", display = "Method" }
]
```

### 2. 필터 정의 필드 설명

- **name**: 백엔드/프론트엔드에서 사용하는 필터 식별자
- **field**: OpenSearch에서 aggregation 및 필터링에 사용할 실제 필드명
- **display**: UI에 표시될 필터 이름
- **case_insensitive** (선택): true로 설정하면 대소문자 구분 없이 검색 (예: "info"와 "INFO" 모두 매칭)

## 동작 방식

### 백엔드 (api.go)

1. **설정 로딩**: `opensearchConfig.IndexSchema`에 모든 인덱스 스키마가 저장됩니다
2. **스키마 매칭**: `findSchemaForIndex(index string)` 함수가 인덱스 패턴과 일치하는 스키마를 찾습니다
3. **동적 필터 생성**: `buildTermFilters(req)` 함수가 스키마를 기반으로 OpenSearch 쿼리 필터를 동적으로 생성합니다
4. **aggregations API**: 스키마에 정의된 모든 필터 필드에 대해 자동으로 aggregation을 생성합니다

### 프론트엔드 (log-dashboard.tsx)

1. **스키마 로딩**: `/log/config` API에서 `index_schema`를 받아와 저장합니다
2. **인덱스 변경 감지**: 선택된 인덱스가 변경되면 해당 인덱스의 스키마를 찾아 `currentSchema`에 설정합니다
3. **동적 필터 렌더링**: `currentSchema.filters`를 순회하며 각 필터에 대한 드롭다운을 동적으로 생성합니다
4. **검색 요청**: 필터 값들을 `filters` 객체에 담아 백엔드로 전송합니다

## 새 인덱스 추가 절차

1. **설정 파일 편집**: `master_config.toml`에 새로운 `[[log.opensearch.index_schema]]` 섹션 추가
2. **인덱스 목록 업데이트**: `indices` 배열에 새 인덱스 패턴 추가
3. **서버 재시작**: pharos 재시작하여 설정 적용
4. **확인**: UI에서 새 인덱스 선택 시 정의한 필터들이 자동으로 표시되는지 확인

## 예시: 새로운 audit-* 인덱스 추가

```toml
# audit-* 인덱스: 감사 로그
[[log.opensearch.index_schema]]
pattern = "audit-*"
filters = [
  { name = "user",      field = "user.keyword",      display = "User" },
  { name = "action",    field = "action.keyword",    display = "Action" },
  { name = "resource",  field = "resource.keyword",  display = "Resource" },
  { name = "result",    field = "result.keyword",    display = "Result" }
]
```

이제 `audit-*` 인덱스를 선택하면 User, Action, Resource, Result 필터가 자동으로 나타납니다!

## 트러블슈팅

### 필터가 표시되지 않음
- 인덱스 패턴이 정확한지 확인 (`pattern = "app-*"`)
- 서버를 재시작했는지 확인
- 브라우저 개발자 도구에서 `/log/config` API 응답 확인

### aggregation 값이 없음
- OpenSearch에 해당 필드가 실제로 존재하는지 확인
- 필드명이 정확한지 확인 (`.keyword` suffix 포함 여부)
- 시간 범위 내에 데이터가 있는지 확인

### 대소문자 구분 문제
- `case_insensitive = true` 옵션 추가
- 단, OpenSearch 쿼리가 복잡해지므로 성능에 영향이 있을 수 있음
