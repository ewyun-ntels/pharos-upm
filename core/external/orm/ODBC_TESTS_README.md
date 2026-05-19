# Database Pool ODBC Tests

이 디렉토리에는 ODBC 빌드 태그와 관련된 테스트 파일들이 포함되어 있습니다.

## 테스트 파일 구조

### 1. `database_pool_odbc_test.go`
- **빌드 태그**: `//go:build odbc`
- **목적**: ODBC 빌드 태그가 활성화된 환경에서의 테스트
- **주요 기능**:
  - Altibase 드라이버 지원 확인
  - ODBC 패키지 임포트 검증
  - init() 함수 실행 효과 확인
  - ODBC 환경에서의 통합 테스트

### 2. `database_pool_no_odbc_test.go`
- **빌드 태그**: `//go:build !odbc`
- **목적**: ODBC 빌드 태그가 비활성화된 환경에서의 테스트
- **주요 기능**:
  - Altibase 드라이버 미지원 확인
  - 기본 드라이버들의 정상 동작 확인
  - ODBC 없이도 기본 기능 동작 검증

## 테스트 실행 방법

### ODBC 환경에서 테스트 (Altibase 지원)
```bash
# ODBC 라이브러리가 설치된 환경에서
go test -tags odbc -v ./internal/orm -run TestODBC
```

### 일반 환경에서 테스트 (Altibase 미지원)
```bash
# 기본 환경에서 (ODBC 태그 없음)
go test -v ./internal/orm -run TestNonODBC
```

### 전체 테스트 실행
```bash
# 일반 환경에서 모든 테스트
go test -v ./internal/orm

# ODBC 환경에서 모든 테스트
go test -tags odbc -v ./internal/orm
```

## 테스트 커버리지

### ODBC 빌드 테스트 (`database_pool_odbc_test.go`)
- ✅ supportDriver 맵에서 Altibase 지원 확인
- ✅ 모든 드라이버 지원 상태 검증
- ✅ Altibase 특화 기능 테스트
- ✅ init() 함수 실행 효과 검증
- ✅ ODBC 통합 테스트
- ✅ 성능 벤치마크

### 비-ODBC 빌드 테스트 (`database_pool_no_odbc_test.go`)
- ✅ supportDriver 맵에서 Altibase 미지원 확인
- ✅ 부분 드라이버 지원 상태 검증
- ✅ Altibase 설정은 유지되지만 연결 실패 확인
- ✅ 다른 드라이버들의 정상 동작 확인
- ✅ 데이터베이스 풀 동작 테스트
- ✅ 성능 벤치마크

## 빌드 태그 동작 검증

### ODBC 빌드일 때 (`//go:build odbc`)
- `supportDriver[DriverAltibase] = true`
- ODBC 패키지 임포트
- Altibase 연결 시도 가능 (실제 ODBC 라이브러리 필요)

### 일반 빌드일 때 (`//go:build !odbc`)
- `supportDriver[DriverAltibase] = false`
- ODBC 패키지 미임포트
- Altibase 연결 시도 시 "not supported" 에러

## 성능 측정 결과

### 일반 빌드 성능 (Non-ODBC)
- GetDriverName: ~25ns/op
- GetDataSourceName: ~3ns/op
- SupportDriverLookup: ~11ns/op

### 기능별 테스트 결과
- ✅ 모든 구조체 테스트 통과
- ✅ 데이터 소스 생성 테스트 통과
- ✅ 에러 처리 테스트 통과
- ✅ 동시성 테스트 통과
- ✅ 통합 테스트 통과

## 주의사항

1. **ODBC 라이브러리 의존성**: ODBC 빌드 태그를 사용한 테스트는 실제 ODBC 라이브러리(`sql.h`)가 시스템에 설치되어 있어야 합니다.

2. **조건부 컴파일**: 빌드 태그에 따라 다른 코드가 컴파일되므로, 두 환경 모두에서 테스트해야 합니다.

3. **CI/CD 고려사항**: 
   - 일반 빌드는 대부분의 CI 환경에서 실행 가능
   - ODBC 빌드는 별도의 ODBC 라이브러리 설치가 필요한 환경에서만 실행

## 테스트 파일 관계도

```
database_pool.go (기본 supportDriver 정의)
├── database_pool_odbc.go (ODBC 빌드 시 Altibase 활성화)
│   └── database_pool_odbc_test.go (ODBC 테스트)
└── database_pool_no_odbc_test.go (비-ODBC 테스트)
```

이 구조를 통해 빌드 환경에 관계없이 적절한 테스트가 실행되며, 각 환경에서의 동작을 정확히 검증할 수 있습니다.
