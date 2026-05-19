# CI/CD 파이프라인 및 배포

## GitHub Actions 워크플로우

### `ci.yml` — 코드 품질 검증

**트리거**: `main` 브랜치 push, PR  
**러너**: self-hosted (`pharos`, `ci` 레이블)

```yaml
steps:
  1. Checkout
  2. Set up Go (go.work의 Go 버전 기준)
  3. make build_prepare       # go work sync + 메타데이터 생성
  4. staticcheck 설치 + 실행  # core/... 및 모든 extensions 정적 분석
  5. go build ./core/...      # 빌드 검증
  6. 모든 extensions go build  # 각 모듈별 빌드 검증
```

**캐시**: staticcheck 결과 캐시 (`~/.cache/staticcheck`)

---

### `deploy-catv.yml` — CATV 빌드·배포

**트리거**: `main`, `test-cicd` 브랜치 push  
**병렬화 방지**: `concurrency: cancel-in-progress: false` (진행 중 배포 취소 차단)  
**타임아웃**: 120분

```yaml
steps: 1. Checkout
  2. Docker Buildx 설정
  3. Harbor 레지스트리 로그인 (ntels.harbor.core)
  4. SITE_MODE=catv make docker
  → Dockerfile 멀티 스테이지 빌드
  → Harbor push
  5. kubectl rollout restart statefulset catv-master -n catv
  6. kubectl rollout status (완료 대기)
  7. kubeconfig 파일 정리 (항상 실행)
```

---

### `release.yml` — Github Release 생성

**트리거**: 태그 push (`v*`)

태그 규칙: `vX.Y.Z-catv` 또는 `vX.Y.Z-catv-beta.N`

---

### `trivy-daily.yml` — 보안 취약점 스캔

**트리거**: 매일 스케줄  
**도구**: Aqua Security Trivy

컨테이너 이미지 및 Go 의존성에 대한 CVE 취약점을 검사합니다.

---

## Docker 빌드 (`Dockerfile`)

멀티 스테이지 빌드로 최종 이미지 크기를 최소화합니다.

### Stage 1: Frontend 빌드 (`builder_pharos_frontend`)

```
Base: node:24-alpine
1. pnpm install --frozen-lockfile (캐시 마운트)
2. SITE_MODE=${SITE_MODE} pnpm build
   → meta/scripts/generate-metadata.ts 실행 (extension-loader.ts 생성)
   → Vite 빌드 → core/frontend/out/ 생성
```

### Stage 2: Go 백엔드 빌드 (`builder_pharos_backend`)

```
Base: golang:1.26.1-trixie
1. apt-get: upx-ucl unixodbc-dev
2. COPY --from=deps /app/core/frontend/out core/frontend/out
   (Stage 1 빌드 결과 복사)
3. make build_prepare build_pharos
   → go work sync
   → generate-metadata.go (import_extensions.go 생성)
   → go generate ./core/ (go:embed 처리)
   → go build -ldflags "..." -o /APP/bin/pharos ./core/cmd/pharos
```

### Stage 3: 런타임 이미지

```
Base: debian:stable-slim
패키지: curl tzdata unixodbc-dev

설치 경로:
  /opt/pharos/bin/     → 실행 파일 (pharos, run.sh, plugins/)
  /opt/pharos/config/  → 공유 라이브러리 (.so 파일들), TLS 인증서

환경 변수:
  MASTER_STATISTICS_ALTIBASE_DRIVER  → altibase 7.3.0 .so 경로
  SHARED-LIBRARY_ALTIBASE_6_DRIVER   → altibase 6.5.1.10.3 .so 경로
  PLUGINS_PATH                       → bin/plugins/ 경로

ENTRYPOINT: /opt/pharos/bin/run.sh
```

---

## 빌드 시 버전 정보 주입

```makefile
GIT_COMMIT  := $(shell git rev-parse --short=10 HEAD)
GIT_TAG     := $(shell git describe --tags --abbrev=0)
BUILD_TIME  := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS += -X 'ntels.com/pharos/core/pkg/version.Version=$(GIT_TAG)'
LDFLAGS += -X 'ntels.com/pharos/core/pkg/version.Commit=$(GIT_COMMIT)'
LDFLAGS += -X 'ntels.com/pharos/core/pkg/version.BuildTime=$(BUILD_TIME)'
```

---

## Makefile 주요 타겟

| 타겟                              | 설명                                         |
| --------------------------------- | -------------------------------------------- |
| `make build`                      | 전체 빌드 (프론트 + 바이너리 + 플러그인)     |
| `make build_without_plugin`       | 플러그인 제외 빌드                           |
| `make generate_pharos`            | 프론트엔드 빌드 (SITE_MODE 기반 Vite 빌드)   |
| `make build_prepare`              | go work sync + 메타데이터 생성 + go generate |
| `make build_pharos`               | Go 바이너리 빌드                             |
| `make build_plugin`               | 모든 플러그인 빌드                           |
| `make build_plugin_altibase`      | Altibase 플러그인 (CGO + unixODBC)           |
| `make build_plugin_clickhouse`    | ClickHouse 플러그인                          |
| `make build_plugin_postgresql`    | PostgreSQL 플러그인                          |
| `make build_plugin_http-receiver` | HTTP Receiver 플러그인                       |
| `make docker`                     | Docker 이미지 빌드 + Harbor push             |
| `make pack_build_pharos`          | UPX 압축 (최적화)                            |

---

## 로컬 개발 환경

### 전제 조건

```bash
# Go 1.25+
go version

# Node.js 24+ (nvm 권장)
nvm install 24 && nvm use 24

# pnpm
npm install -g pnpm
```

### 의존성 설치

```bash
# Go workspace 동기화
go work sync

# Frontend 의존성
pnpm install
```

### 개발 서버 실행

**터미널 1 (백엔드)**:

```bash
SITE_MODE=catv make build_pharos
./bin/pharos server --config ./migrations/catv_quality/config/config.toml
```

**터미널 2 (프론트엔드 개발 서버)**:

```bash
API_SERVER_URL=http://localhost:8080 SITE_MODE=catv pnpm dev
# → http://localhost:3000/ui/ 접속
```

### 타입 생성

```bash
# Frontend TypeScript 타입 (shared/schema/ 변경 후)
pnpm generate:types

# Backend Go 타입
cd shared && go run scripts/generate-types/main.go
```

---

## Kubernetes 배포 구성

- **워크로드**: `StatefulSet` (`catv-master`) — 단일 마스터 서버
- **네임스페이스**: `catv`
- **롤링 업데이트**: `kubectl rollout restart statefulset catv-master -n catv`
- **STB 제어 워커**: `Job` 타입 (`catv-control-settopbox`) — 마스터 서버가 동적으로 생성
  - ServiceAccount: `config.Catv.Control.K8sJob.ServiceAccountName`
  - 이미지: `config.Catv.Control.K8sJob.Image` (동일 pharos 이미지)
  - 허용 환경변수: `config.Catv.Control.K8sJob.AllowedEnvVars` (설정 파일에서 화이트리스트 지정)
