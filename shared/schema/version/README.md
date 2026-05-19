# Version Schema Documentation

Version 시스템의 JSON Schema 정의와 사용법을 설명합니다.

## 📋 Schema Files

### 1. VersionResponse.schema  
Version 정보 조회 응답을 위한 스키마입니다.

**위치**: `shared/schema/version/VersionResponse.schema`  
**용도**: GET 응답에서 사용  
**생성 타입**: `shared/types/versionresponse.Types`

## 🏗️ 아키텍처 패턴

```
Frontend ←→ shared/types ←→ Adapter ←→ Internal Types ←→ Backend
            ↓ 자동 생성     ↓ 변환      ↓ 빌드 정보 통합
        JSON Schema     ToAPI      Response + DB Row
```

## 📖 사용 예제

### GET Version Info (조회)
```bash
curl /api/version
```

**응답 예시**:
```json
{
  "module": "core",
  "version": "1.2.3",
  "commit": "abc123def456",
  "build_time": "2025-09-27T08:00:00Z",
  "updated_at": "2025-09-27T10:30:00Z"
}
```

## 🔄 Adapter 패턴

### Handler에서의 사용
```go
// Internal Response 준비
internalResp := adapters.Response{
    Module:    "core",
    Version:   Version,        // 빌드 시 임베드
    Commit:    Commit,         // 빌드 시 임베드
    BuildTime: parsedBuildTime,
}

// DB에서 추가 정보 조회 (fallback 방식)
if row, found, _ := m.GetCoreVersionRow(); found {
    if row.Module != "" {
        internalResp.Module = row.Module
    }
    // ... 다른 필드들도 DB 값으로 덮어쓰기
}

// API 응답으로 변환
adapter := adapters.NewVersionAdapter()
apiResp := adapter.ToAPIVersionResponse(internalResp)
c.JSON(http.StatusOK, apiResp)
```

### 타입 변환 흐름
1. **Build Info**: `내장 상수 → internal Response`
2. **DB Info**: `DB Row → internal Response (덮어쓰기)`  
3. **API Response**: `internal Response → ToAPIVersionResponse() → versionresponse.Types → JSON`

## 🏗️ 빌드 정보 통합

### 빌드 시 임베드 정보
```go
// 컴파일 타임에 주입되는 정보
var (
    Version   = "dev"
    Commit    = "unknown"  
    BuildTime = "unknown"
)
```

### DB 정보와의 통합
```sql
-- core_app_version 테이블
SELECT module, version, commit, build_time, updated_at 
FROM core_app_version 
WHERE module = 'core';
```

### Fallback 메커니즘
1. **1차**: 빌드 시 임베드된 기본 정보 사용
2. **2차**: DB에서 조회한 정보가 있으면 덮어쓰기
3. **결과**: 가장 최신 정보로 응답 구성

## ⏰ 시간 처리

### 다양한 시간 형식 지원
```go
func parseTimeBestEffort(s string) (time.Time, bool) {
    layouts := []string{
        time.RFC3339,     // "2025-09-27T10:30:00Z"
        time.RFC3339Nano, // "2025-09-27T10:30:00.123Z"  
        time.DateTime,    // "2025-09-27 10:30:00"
    }
    // ... UTC로 정규화
}
```

### 시간 정규화
- 모든 시간은 **UTC**로 변환
- **초 단위**로 절사하여 일관성 확보

## 🔒 Read-Only 설계

### 특징
- **조회 전용**: GET 엔드포인트만 제공
- **FromAPI* 없음**: 입력 스키마 불필요
- **ToAPI*만 구현**: 응답 변환만 필요

### 장점
1. **단순성**: 복잡한 입력 검증 불필요
2. **안정성**: 시스템 정보 변경 방지
3. **최적화**: 조회 성능에만 집중

## 📊 제공 정보

### Module
- **의미**: 애플리케이션 모듈 이름
- **기본값**: "core"
- **소스**: DB > 빌드 임베드

### Version  
- **의미**: 애플리케이션 버전
- **형식**: SemVer (예: "1.2.3")
- **소스**: DB > 빌드 임베드

### Commit
- **의미**: Git 커밋 해시
- **형식**: 짧은 해시 (예: "abc123def")
- **소스**: DB > 빌드 임베드

### Build Time
- **의미**: 빌드 수행 시각
- **형식**: RFC3339 (UTC)
- **소스**: DB > 빌드 임베드

### Updated At
- **의미**: DB 레코드 최종 수정 시각
- **형식**: RFC3339 (UTC)
- **소스**: DB 전용 (빌드 정보에는 없음)

## ✅ 검증된 기능

- ✅ Read-Only API에 최적화된 스키마 설계
- ✅ 빌드 정보와 DB 정보의 안정적 통합  
- ✅ 다양한 시간 형식 지원 및 UTC 정규화
- ✅ Fallback 메커니즘으로 항상 응답 보장
- ✅ Frontend와 Backend 타입 동기화  
- ✅ 13개 테스트 케이스 통과

## 🔗 관련 파일

- **Handler**: `core/pkg/version/handler.go`
- **Adapter**: `core/pkg/version/adapters/version_adapter.go`
- **Model**: `core/pkg/version/model.go`
- **Tests**: `core/pkg/version/adapters/version_adapter_test.go`
- **Types**: `shared/types/versionresponse/`

## 🎯 장점

1. **정보 통합**: 빌드 정보 + DB 정보를 지능적으로 통합
2. **Fallback 안정성**: DB 오류 시에도 빌드 정보로 응답 보장
3. **시간 정규화**: 다양한 입력 형식을 UTC로 일관성 있게 처리
4. **Read-Only 최적화**: 조회 전용으로 단순하고 안전한 설계  
5. **타입 안전성**: JSON Schema로 응답 형식 보장