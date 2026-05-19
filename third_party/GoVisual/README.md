

##  Gin Web Framework Request 이력 관리
Gin Web Framework를 사용하는 Application의 Request 이력 관리 기능을 제공

GoVisual HTTP 요청 디버깅 도구에서 제공하는 기능을 참고하여 구현되었습니다.

### 추가 기능
``` shell
1. ClickHouse Storage 기능 제공
2. websocket Request 이력 관리 기능 제공
```

## 사용법
/examples/history/chat/main.go
```
import "ntels.com/pharos/pkg/history"

r := gin.New()
r.Use(gin.Logger(), gin.Recovery())
r.GET("/connection/websocket", gin.WrapH(authMiddleware(websocketHandler)))
r.GET("/metrics", gin.WrapH(promhttp.Handler()))
r.StaticFile("/", "./index.html")

cfg, _ := config_loader.Load("/tarzan/config/config.toml")
handler := history.Wrap(r,
		history.WithClickHouseStorage(&config),
		history.WithRequestBodyLogging(false),  // request body logging 여부(true/false)
		history.WithResponseBodyLogging(false), // response body logging 여부(true/false)
	)

server := &http.Server{
    Addr:         "127.0.0.1:" + strconv.Itoa(*port),
    Handler:      handler,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
}
```

##  Gin Web Framework Request 이력 관리
Gin Web Framework를 사용하는 Application의 Request 이력 관리 기능을 제공

GoVisual HTTP 요청 디버깅 도구에서 제공하는 기능을 참고하여 구현되었습니다.

### 추가 기능
``` shell
1. ClickHouse Storage 기능 제공
2. websocket Request 이력 관리 기능 제공
```

## 변경 이력 (Changelog)
- 2025-08-28 SQLite/PostgreSQL 스토리지 안정성 및 보안 개선 (15:15)
  - SQLite/PG 테이블 식별자 안전 처리: 테이블명 및 인덱스명에 대해 안전한 인용 처리 추가
    - SQLite: quoteSQLiteIdent 도입, 모든 DDL/DML에서 식별자에 적용 (CREATE TABLE, CREATE INDEX, SELECT/DELETE 등)
    - PostgreSQL: quotePGQualifiedIdentifier 도입, 스키마.테이블 형태 지원 및 모든 쿼리에 적용
  - 인덱스 생성 시 인덱스명도 안전하게 인용 처리하여 특수문자/충돌 이슈 방지
  - Clear/cleanup/Get/GetAll/GetLatest 등 전체 쿼리 경로에서 안전 인용 적용으로 식별자 주입 가능성 차단
  - SQLite 리소스 관리 개선: rows.Close를 안전한 defer 형태로 변경하여 오류 전파 방지
  - JSON 언마샬링 에러 무시 방식을 명시적으로(_ =) 적용해 로그 파싱 실패로 인한 흐름 중단 방지
  - SQLite 불필요한 드라이버 등록 코드 제거 및 strings 패키지로 교체, 의존성 단순화
- 2025-08-27 보안 개선 작업
  - OTLP gRPC 연결 보안 강화: insecure.NewCredentials() 제거, TLS(credentials.NewTLS(&tls.Config{}))로 교체하여 암호화된 연결 사용 (third_party/history/internal/telemetry/telemetry.go).
  - 외부 명령 실행 하드닝: exec.Command → exec.CommandContext 로 변경하고, 명령/인자에 대한 정규식 기반 유효성 검증 추가하여 코드 주입 위험 완화 (third_party/goflow/operator.go).
  - 플러그인 바이너리 실행 안전성 강화: 플러그인 파일명 검증(isSafePluginBinaryName) 추가 및 filepath.Join 사용. 검증 실패 시 로드/실행을 건너뜀 (pkg/plugins/load.go).
  - 경로 처리 보안 개선: path.Clean을 경로 검증/살균 용도로 사용하지 않도록 수정. "/" 강제 부여 + strings.Trim + path.Clean + filepath.FromSlash 조합으로 정규화 (third_party/history/internal/dashboard/handler.go).
  - 미완성 대시보드 경로 처리 제거: Wrap에서 대시보드 경로 특수 처리로 인해 빈 응답이 발생하던 문제를 제거하고 모든 요청을 애플리케이션 핸들러로 위임하도록 단순화 (third_party/history/wrap.go).
  - 옵션 정리: DashboardPath 필드 및 해당 경로 자동 무시 로직 제거. 이제 IgnorePaths 설정만으로 무시 경로를 제어 (third_party/history/options.go).
  - SQL 보안 강화:
    - PostgreSQL 저장소(postgres.go): 테이블명(식별자) 주입 방지를 위해 엄격한 유효성 검증(영숫자/언더스코어 및 선택적 스키마.테이블만 허용)을 추가했습니다. 값 파라미터는 기존대로 $1, $2… 플레이스홀더로 바인딩합니다.
    - SQLite 백업 작업(pkg/workflow/types/task/sqlite/backup.go): `VACUUM main INTO ?` 형태로 플레이스홀더를 사용하도록 수정하여 파일 경로 문자열을 직접 SQL에 연결하지 않도록 했습니다.

## 사용법
/examples/history/chat/main.go
```
import "ntels.com/pharos/pkg/history"

r := gin.New()
r.Use(gin.Logger(), gin.Recovery())
r.GET("/connection/websocket", gin.WrapH(authMiddleware(websocketHandler)))
r.GET("/metrics", gin.WrapH(promhttp.Handler()))
r.StaticFile("/", "./index.html")

cfg, _ := config_loader.Load("/tarzan/config/config.toml")
handler := history.Wrap(r,
		history.WithClickHouseStorage(&config),
		history.WithRequestBodyLogging(false),  // request body logging 여부(true/false)
		history.WithResponseBodyLogging(false), // response body logging 여부(true/false)
	)

server := &http.Server{
    Addr:         "127.0.0.1:" + strconv.Itoa(*port),
    Handler:      handler,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
}
```

---

## 라이선스 고지 (Third-party Notice)
본 모듈은 GoVisual(https://github.com/doganarif/GoVisual.git) 프로젝트의 아이디어/기능을 참고하여 개발되었습니다. GoVisual은 MIT 라이선스 하에 배포됩니다. 본 저장소는 GoVisual의 라이선스 의무를 준수하기 위해 다음과 같은 고지를 포함합니다. 원저작물의 정확한 저작권자와 연도는 해당 프로젝트의 LICENSE 파일을 참고하십시오.

아래는 MIT 라이선스의 표준 허가 고지문입니다 (원저작물의 저작권 표기는 해당 원문을 따릅니다):

```
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

참고: GoVisual의 원문 라이선스 전문과 상세 정보는 해당 프로젝트 저장소의 LICENSE 문서를 확인해 주세요.

r.GET("/metrics", gin.WrapH(promhttp.Handler()))
r.StaticFile("/", "./index.html")

cfg, _ := config_loader.Load("/tarzan/config/config.toml")
handler := history.Wrap(r,
		history.WithClickHouseStorage(&config),
		history.WithRequestBodyLogging(false),  // request body logging 여부(true/false)
		history.WithResponseBodyLogging(false), // response body logging 여부(true/false)
	)

server := &http.Server{
    Addr:         "127.0.0.1:" + strconv.Itoa(*port),
    Handler:      handler,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
}
```

##  Gin Web Framework Request 이력 관리
Gin Web Framework를 사용하는 Application의 Request 이력 관리 기능을 제공

GoVisual HTTP 요청 디버깅 도구에서 제공하는 기능을 참고하여 구현되었습니다.

### 추가 기능
``` shell
1. ClickHouse Storage 기능 제공
2. websocket Request 이력 관리 기능 제공
```

## 변경 이력 (Changelog)
- 2025-08-28 SQLite/PostgreSQL 스토리지 안정성 및 보안 개선 (15:15)
  - SQLite/PG 테이블 식별자 안전 처리: 테이블명 및 인덱스명에 대해 안전한 인용 처리 추가
    - SQLite: quoteSQLiteIdent 도입, 모든 DDL/DML에서 식별자에 적용 (CREATE TABLE, CREATE INDEX, SELECT/DELETE 등)
    - PostgreSQL: quotePGQualifiedIdentifier 도입, 스키마.테이블 형태 지원 및 모든 쿼리에 적용
  - 인덱스 생성 시 인덱스명도 안전하게 인용 처리하여 특수문자/충돌 이슈 방지
  - Clear/cleanup/Get/GetAll/GetLatest 등 전체 쿼리 경로에서 안전 인용 적용으로 식별자 주입 가능성 차단
  - SQLite 리소스 관리 개선: rows.Close를 안전한 defer 형태로 변경하여 오류 전파 방지
  - JSON 언마샬링 에러 무시 방식을 명시적으로(_ =) 적용해 로그 파싱 실패로 인한 흐름 중단 방지
  - SQLite 불필요한 드라이버 등록 코드 제거 및 strings 패키지로 교체, 의존성 단순화
- 2025-08-27 보안 개선 작업
  - OTLP gRPC 연결 보안 강화: insecure.NewCredentials() 제거, TLS(credentials.NewTLS(&tls.Config{}))로 교체하여 암호화된 연결 사용 (third_party/history/internal/telemetry/telemetry.go).
  - 외부 명령 실행 하드닝: exec.Command → exec.CommandContext 로 변경하고, 명령/인자에 대한 정규식 기반 유효성 검증 추가하여 코드 주입 위험 완화 (third_party/goflow/operator.go).
  - 플러그인 바이너리 실행 안전성 강화: 플러그인 파일명 검증(isSafePluginBinaryName) 추가 및 filepath.Join 사용. 검증 실패 시 로드/실행을 건너뜀 (pkg/plugins/load.go).
  - 경로 처리 보안 개선: path.Clean을 경로 검증/살균 용도로 사용하지 않도록 수정. "/" 강제 부여 + strings.Trim + path.Clean + filepath.FromSlash 조합으로 정규화 (third_party/history/internal/dashboard/handler.go).
  - 미완성 대시보드 경로 처리 제거: Wrap에서 대시보드 경로 특수 처리로 인해 빈 응답이 발생하던 문제를 제거하고 모든 요청을 애플리케이션 핸들러로 위임하도록 단순화 (third_party/history/wrap.go).
  - 옵션 정리: DashboardPath 필드 및 해당 경로 자동 무시 로직 제거. 이제 IgnorePaths 설정만으로 무시 경로를 제어 (third_party/history/options.go).
  - SQL 보안 강화:
    - PostgreSQL 저장소(postgres.go): 테이블명(식별자) 주입 방지를 위해 엄격한 유효성 검증(영숫자/언더스코어 및 선택적 스키마.테이블만 허용)을 추가했습니다. 값 파라미터는 기존대로 $1, $2… 플레이스홀더로 바인딩합니다.
    - SQLite 백업 작업(pkg/workflow/types/task/sqlite/backup.go): `VACUUM main INTO ?` 형태로 플레이스홀더를 사용하도록 수정하여 파일 경로 문자열을 직접 SQL에 연결하지 않도록 했습니다.

## 사용법
/examples/history/chat/main.go
```
import "ntels.com/pharos/pkg/history"

r := gin.New()
r.Use(gin.Logger(), gin.Recovery())
r.GET("/connection/websocket", gin.WrapH(authMiddleware(websocketHandler)))
r.GET("/metrics", gin.WrapH(promhttp.Handler()))
r.StaticFile("/", "./index.html")

cfg, _ := config_loader.Load("/tarzan/config/config.toml")
handler := history.Wrap(r,
		history.WithClickHouseStorage(&config),
		history.WithRequestBodyLogging(false),  // request body logging 여부(true/false)
		history.WithResponseBodyLogging(false), // response body logging 여부(true/false)
	)

server := &http.Server{
    Addr:         "127.0.0.1:" + strconv.Itoa(*port),
    Handler:      handler,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
}
```

---

## 라이선스 고지 (Third-party Notice)
본 모듈은 GoVisual(https://github.com/doganarif/GoVisual.git) 프로젝트의 아이디어/기능을 참고하여 개발되었습니다. GoVisual은 MIT 라이선스 하에 배포됩니다. 본 저장소는 GoVisual의 라이선스 의무를 준수하기 위해 다음과 같은 고지를 포함합니다. 원저작물의 정확한 저작권자와 연도는 해당 프로젝트의 LICENSE 파일을 참고하십시오.

아래는 MIT 라이선스의 표준 허가 고지문입니다 (원저작물의 저작권 표기는 해당 원문을 따릅니다):

```
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

참고: GoVisual의 원문 라이선스 전문과 상세 정보는 해당 프로젝트 저장소의 LICENSE 문서를 확인해 주세요.



# 2025-09-11 업데이트: Statistics DB 연동 및 Request History 통합
- third_party/history 모듈이 애플리케이션의 Statistics Database 설정과 연동되도록 개선되었습니다.
- 저장소 자동 선택: Wrap에 history.WithClickHouseStorage(&config)를 전달하면 config.Statistics.Database.driver 값에 따라 자동으로 적절한 스토리지 백엔드가 선택됩니다.
  - sqlite → 내부 SQLiteStore 사용
  - postgresql → 내부 PostgresStore 사용
  - clickhouse → 내부 ClickHouseStore 사용 (sqlx/ORM 경유)
- ClickHouse HTTP 의존성 제거: clickhouse 전용 HTTP 인터페이스를 제거하고, 내부 ORM(StatisticsHandler) + sqlx 경유로 동일한 코드 경로에서 INSERT/SELECT 하도록 변경했습니다.
- 공통 스키마 통합: 모든 DB에서 동일한 테이블 이름과 컬럼을 사용합니다: history_requests
  - 컬럼
    - id TEXT/String
    - timestamp DATETIME/TIMESTAMP
    - user TEXT/String (JWT Authorization 토큰의 subject를 추출하여 저장; 없으면 빈 문자열)
    - method TEXT/String
    - path TEXT/String
    - query TEXT/String
    - request_headers TEXT/String (JSON 문자열)
    - response_headers TEXT/String (JSON 문자열)
    - status_code INTEGER/Int32
    - duration INTEGER/Int64 (나노/마이크로 단위 가능; 애플리케이션 값 그대로 저장)
    - request_body TEXT/String
    - response_body TEXT/String
    - error TEXT/String
    - middleware_trace TEXT/String (JSON 문자열)
    - route_trace TEXT/String (JSON 문자열)
    - created_at DATETIME/TIMESTAMP (기본값: 현재 시간; ClickHouse는 DEFAULT now())
- 마이그레이션 제공:
  - SQLite: pkg/migration/databases/master/statistics/sqlite/20250910000005_request_history.sql
  - PostgreSQL: pkg/migration/databases/master/statistics/postgresql/20250910000005_request_history.sql
  - ClickHouse: pkg/migration/databases/master/statistics/clickhouse/20250819000005_request_history.sql
- JSON 보관 방식: SQLite/PostgreSQL/ClickHouse 모두 TEXT/String 컬럼에 JSON 문자열로 저장합니다.
- 용량 관리: SQLite/PostgreSQL은 capacity 초과 시 오래된 레코드를 삭제하는 cleanup 로직이 포함되어 있습니다. ClickHouse는 테이블 TTL/엔진 설정으로 관리하세요.

## 설정 및 사용 방법
1) Statistics DB 설정 (config.toml 예시)
```
[statistics.database]
# sqlite 예시
driver = "sqlite"
[statistics.database.sqlite]
path = "./statistics.db"

# postgresql 예시
# driver = "postgresql"
# [statistics.database.postgresql]
# host = "127.0.0.1"
# port = 5432
# username = "user"
# password = "pass"
# database = "pharos"

# clickhouse 예시
# driver = "clickhouse"
# [statistics.database.clickhouse]
# host = "127.0.0.1"
# port = 9000
# username = "default"
# password = ""
# database = "pharos"
```

2) 마이그레이션 실행
- 애플리케이션의 migration 러너가 statistics.database 설정에 맞춰 history_requests 테이블을 생성합니다.

3) 미들웨어 적용
```
handler := history.Wrap(r,
    // 통합 설정: statistics.database 기반으로 백엔드 자동 선택
    history.WithClickHouseStorage(&config),
    history.WithRequestBodyLogging(false),
    history.WithResponseBodyLogging(false),
)
```

참고
- 기본 테이블명은 history_requests 입니다. 커스텀 테이블명을 사용하려면 명시적 스토리지 옵션(예: WithSQLiteStorage/WithPostgresStorage)으로 테이블명을 지정할 수 있습니다.
- Authorization 헤더가 없거나 형식이 올바르지 않으면 user 컬럼은 빈 문자열로 저장됩니다.
