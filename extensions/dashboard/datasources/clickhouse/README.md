# ClickHouse Datasource Extension

ClickHouse 데이터소스를 Extension 시스템으로 마이그레이션한 예제입니다.

## 개요

기존 built-in ClickHouse 플러그인을 독립적인 Extension으로 이관하여:
- 모듈화된 구조 제공
- 독립적인 개발/배포 가능
- 확장 가능한 아키텍처 구현

## 프로젝트 구조

```
extensions/clickhouse-datasource/
├── README.md                  # 이 파일
├── go.mod                     # Go 모듈 정의
├── go.sum                     # 의존성 체크섬
├── clickhouse.go              # Extension 메인 구현
├── extension.config.json      # Extension 메타데이터 (미래 확장용)
├── schema/                    # JSON Schema 파일들
│   ├── json.json              # 데이터소스 설정 스키마
│   ├── ui.json                # UI 스키마
│   └── sample-form-data.json  # 샘플 설정 데이터
└── frontend/                  # 프론트엔드 컴포넌트
    ├── package.json           # 프론트엔드 의존성
    └── src/
        ├── index.ts           # 엔트리 포인트
        ├── types.ts           # TypeScript 타입 정의
        └── components/
            ├── ConfigEditor.tsx   # 설정 에디터
            ├── QueryBuilder.tsx   # 쿼리 빌더
            └── styles.css         # 스타일시트
```

## 빌드 방법

### 백엔드 빌드

1. **Extension 활성화**
   ```bash
   # core/cmd/pharos/import_extensions.go에 import 추가
   _ "ntels.com/pharos/extensions/clickhouse-datasource"
   ```

2. **Pharos 빌드**
   ```bash
   cd core/
   go build -o ../pharos ./cmd/pharos
   ```

3. **서버 실행**
   ```bash
   ./pharos serve --config config.toml
   ```

### 프론트엔드 빌드 (선택사항)

```bash
cd frontend/
npm install
npm run build
```

## Extension 등록 과정

1. **자동 등록**: 패키지가 import될 때 `init()` 함수가 자동 실행
2. **Extension 생성**: `NewClickHouseExtension()` 호출
3. **레지스트리 등록**: `plugins.RegisterExtension()` 호출
4. **플러그인 로딩**: Extension에서 기존 ClickHouse 플러그인 재사용

## API 사용법

### 데이터소스 목록 조회
```bash
curl http://localhost:8080/plugins/datasources
```

### 쿼리 실행
```bash
curl -X POST http://localhost:8080/plugins/ds/query \
  -H "Content-Type: application/json" \
  -d '{
    "queries": [
      {
        "id": "test-query",
        "datasourceName": "clickhouse-01",
        "sql": "SELECT count() FROM table_name",
        "timeout": 30
      }
    ]
  }'
```

## Extension 인터페이스

```go
type Extension interface {
    GetPlugin() (*model.Plugin, error)     // 플러그인 설정 반환
    RegisterRoutes(router *gin.RouterGroup) // 커스텀 API 엔드포인트 등록
    GetName() string                       // Extension 이름
    GetVersion() string                    // Extension 버전
    GetType() string                       // Extension 타입
}
```

## 설정 예시

`config.toml`에서 데이터소스 설정:

```toml
[[provisioning.plugins.datasources]]
type = "clickhouse"
name = "clickhouse-01"
[provisioning.plugins.datasources.data]
host = "192.168.15.105"
port = 9000
database = "default"
username = "default"
password = "default"
max_open_connection = 10
max_lifetime = 14400
```

## 개발 가이드

### 새로운 Extension 만들기

1. **디렉토리 생성**
   ```bash
   mkdir extensions/my-datasource
   cd extensions/my-datasource
   ```

2. **go.mod 초기화**
   ```bash
   go mod init ntels.com/pharos/extensions/my-datasource
   ```

3. **Extension 구현**
   ```go
   package mydatasource

   import (
       "github.com/gin-gonic/gin"
       "ntels.com/pharos/core/pkg/common"
       "ntels.com/pharos/core/pkg/plugins"
       "ntels.com/pharos/core/pkg/plugins/model"
   )

   func init() {
       extension := NewMyExtension(common.Config{})
       plugins.RegisterExtension(extension)
   }

   type MyExtension struct {
       config common.Config
   }

   func (e *MyExtension) GetPlugin() (*model.Plugin, error) { /* 구현 */ }
   func (e *MyExtension) RegisterRoutes(router *gin.RouterGroup) { /* 구현 */ }
   func (e *MyExtension) GetName() string { return "my-datasource" }
   func (e *MyExtension) GetVersion() string { return "1.0.0" }
   func (e *MyExtension) GetType() string { return "datasource" }
   ```

4. **Extension 등록**
   ```bash
   # import_extensions.go에 추가
   _ "ntels.com/pharos/extensions/my-datasource"
   ```

### 커스텀 API 추가

```go
func (c *ClickHouseExtension) RegisterRoutes(router *gin.RouterGroup) {
    router.GET("/schema", c.getSchemaInfo)
    router.POST("/preview", c.previewQuery)
    router.GET("/tables", c.getTableList)
}

// 접근 URL: /plugins/ext/clickhouse-datasource/schema
func (c *ClickHouseExtension) getSchemaInfo(ctx *gin.Context) {
    // 스키마 정보 반환
}
```

## 마이그레이션 체크리스트

- [x] Extension 구조 생성
- [x] Go 모듈 설정
- [x] Extension 인터페이스 구현
- [x] 자동 등록 시스템
- [x] 기존 플러그인 재사용
- [x] JSON Schema 복사
- [x] 프론트엔드 컴포넌트 준비
- [x] API 테스트 완료
- [x] 문서 작성

## 테스트 결과

✅ **성공**: 2,729,040개 데이터 정상 조회
✅ **성능**: 쿼리 실행 시간 1.77초
✅ **호환성**: 기존 API 완전 호환

## 다음 단계

1. **다른 데이터소스 이관**: PostgreSQL, Elasticsearch 등
2. **프론트엔드 동적 로딩**: React 컴포넌트 런타임 로드
3. **Extension 관리 UI**: 활성화/비활성화, 설정 관리
4. **패키징 시스템**: Extension 독립 배포