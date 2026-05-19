# API Migrations

## 개요

**API 마이그레이션**은 REST API 호출을 통해 애플리케이션의 초기 설정을 수행하는 마이그레이션입니다.

## 특징

- **실행 시점**: 서버 시작 **후** (백그라운드)
- **목적**: 사이트별 초기 설정, 기본 데이터 투입
- **테이블명**: `core_app_data_schema_migrations`, `core_app_data_migration_runs`, `core_app_data_stored_env`
- **형식**: YAML 파일 (apimigrate)
- **프레임워크**: [apimigrate](https://github.com/loykin/apimigrate)

## 구조

```
api/
├── runner.go           # 실행 로직
└── runner_test.go      # 테스트
```

## 작동 방식

1. **임시 사용자 생성**: SuperAdmin 권한을 가진 임시 계정 생성
2. **OAuth2 토큰 발급**: 임시 계정으로 인증 토큰 획득
3. **마이그레이션 실행**: YAML 파일에 정의된 API 호출 실행
4. **임시 사용자 삭제**: 마이그레이션 완료 후 임시 계정 삭제

## 설정 방법

### config 파일 (master_config.toml)

```toml
[migration]
disable = false                      # 마이그레이션 비활성화 (기본값: false)
wait = "5s"                         # 시작 전 대기 시간
migration_dir = "/path/to/migrations"  # YAML 마이그레이션 디렉토리
artifact_dir = "/path/to/artifacts"    # 응답 저장 디렉토리

[[migration.env]]
name = "SOME_VAR"
value = "some_value"

[[migration.env]]
name = "API_KEY"
value_from_env = "EXTERNAL_API_KEY"  # 환경변수에서 가져오기
```

## 마이그레이션 파일 작성

### YAML 형식 (apimigrate)

```yaml
# 0001_initial_setup.yaml
version: 1
name: "Initial Setup"

steps:
  - name: "Create default organization"
    method: POST
    url: "{{api_base}}/api/v1/organizations"
    headers:
      Content-Type: application/json
    body: |
      {
        "name": "Default Organization",
        "slug": "default"
      }
    extract:
      org_id: "$.data.id"
    
  - name: "Create default team"
    method: POST
    url: "{{api_base}}/api/v1/teams"
    body: |
      {
        "name": "Default Team",
        "organization_id": "{{org_id}}"
      }
```

### 사용 가능한 변수

- `{{api_base}}`: 서버 주소 (자동 설정)
- `{{artifact_dir}}`: 아티팩트 저장 경로
- `{{env.SOME_VAR}}`: config에 정의한 환경변수
- `{{org_id}}`: 이전 step에서 추출한 변수

## 실행 방법

### 자동 실행 (기본)

```go
// server.go (이미 적용됨)
import migrationAPI "ntels.com/pharos/core/pkg/migration/api"

func (s *Server) migration() {
    time.Sleep(200 * time.Millisecond)  // 서버 준비 대기
    migrationAPI.Run(s.config)           // 백그라운드 실행
}
```

### 수동 비활성화

```toml
# config/master_config.toml
[migration]
disable = true  # API 마이그레이션 비활성화
```

## 주의사항

⚠️ **비동기 실행**
- 서버 시작을 막지 않음 (백그라운드 실행)
- 실패해도 서버는 정상 동작함
- 로그를 확인해서 성공 여부 파악

⚠️ **인증 필요**
- OAuth2 토큰이 필요한 API만 호출 가능
- 임시 SuperAdmin 계정 자동 생성/삭제
- TLS 인증서 검증 건너뜀 (InsecureSkipVerify)

⚠️ **Schema/Data 마이그레이션과 분리**
- API 마이그레이션은 **별도 테이블**에 이력 저장
- Schema → Data → API 순서로 실행
- API 마이그레이션 실패해도 앱은 동작함

## 디버깅

### 로그 확인

```bash
# 마이그레이션 시작
migration: starting dir=/path/to/migrations env=... api_base=https://localhost:8443

# 각 버전 실행 결과
migration: applied version=0001 status=200 env=map[org_id:123]

# 완료
migration: completed successfully
```

### 응답 저장

```toml
[migration]
artifact_dir = "/tmp/migration_artifacts"
```

API 응답이 파일로 저장됨:
```
/tmp/migration_artifacts/
├── 0001_step1_response.json
└── 0001_step2_response.json
```

## 참고

- 상세 가이드: [MIGRATION_ARCHITECTURE.md](../../../../../docs/MIGRATION_ARCHITECTURE.md)
- Schema Migrations: [../schema/README.md](../schema/README.md)
- Data Migrations: [../data/README.md](../data/README.md)
- apimigrate 문서: https://github.com/loykin/apimigrate
