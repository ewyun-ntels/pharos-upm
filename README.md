# Pharos

Pharos는 React(Vite) UI와 Go 백엔드로 구성된 관제/데이터 파이프라인 솔루션입니다. `SITE_MODE`에 따라 포함되는 Extension과 마이그레이션 아티팩트가 달라집니다.

## 주요 특징

- **SITE_MODE 기반 멀티 테넌트**: `meta/site-config.json`에서 사이트별 Extension 구성 관리
- **Extension 시스템**: Go Workspaces + pnpm Workspaces를 통한 모듈화 구조
- **플러그인 구조**: Altibase/ClickHouse/PostgreSQL, HTTP Receiver 등 데이터소스 플러그인
- **단일 바이너리 실행**: Go embed를 통해 Frontend가 바이너리에 포함됨
- **NATS 기반 분산 통신**: Master-Agent 간 실시간 통신

## 디렉토리 구조

```
/
├── core/                   # 핵심 Go 애플리케이션 모듈
│   ├── cmd/pharos/         # 실행 엔트리포인트 (main.go)
│   ├── frontend/           # React(Vite) UI 소스
│   ├── pkg/                # 재사용 가능한 Go 패키지
│   ├── internal/           # 내부 전용 Go 패키지
│   ├── config/             # 설정 및 공유 라이브러리 (ODBC 등)
│   └── bin/                # 실행 스크립트 (run.sh)
├── extensions/             # 확장 모듈
│   ├── catv/               # CATV 관련 Extension
│   ├── dashboard/
│   │   ├── panels/         # 대시보드 패널 Extension
│   │   └── datasources/    # 데이터소스 Extension (ClickHouse 등)
│   ├── demo/               # 데모 Extension
│   ├── example/            # Extension 개발 예제
│   ├── header/             # 헤더 UI Extension
│   └── login/              # 로그인 UI Extension
├── shared/                 # 공유 라이브러리 (Go + Frontend)
├── migrations/             # 사이트별 DB 마이그레이션/설정
│   └── [SITE_MODE]/
│       ├── config/         # 사이트별 설정 파일
│       ├── migration/      # DB 마이그레이션 스크립트
│       └── artifact/       # 배포 아티팩트
├── meta/                   # 메타데이터 및 사이트 설정
│   └── site-config.json    # SITE_MODE별 Extension 구성 정의
├── plugins/                # 데이터소스 플러그인
├── tools/                  # 유틸리티 도구 (migration, simulator 등)
├── third_party/            # 서드파티 라이브러리
├── go.work                 # Go Workspaces 설정
├── pnpm-workspace.yaml     # pnpm Workspaces 설정
├── Makefile                # 빌드 스크립트
└── Dockerfile              # 컨테이너 이미지 빌드
```

## 사전 준비

- **Go 1.25+**: [go.dev](https://go.dev)
- **Node.js 24+**: nvm으로 설치 권장
- **pnpm**: `npm install -g pnpm`
- Altibase 플러그인 빌드 시: unixODBC 헤더/라이브러리 필요 (Linux/WSL2)

### nvm 설치

**macOS / Linux**

```bash
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash
# 터미널 재시작 후
nvm install 24
nvm use 24
npm install -g pnpm
```

**Windows** — [nvm-windows](https://github.com/coreybutler/nvm-windows/releases) 설치 후:

```powershell
nvm install 24
nvm use 24
npm install -g pnpm
```

## 기존 환경 업그레이드

Node.js 버전을 올린 경우 네이티브 바이너리 재설치가 필요합니다.

```bash
nvm install 24
nvm use 24
pnpm install
```

## 개발 환경 설정

### 1. 의존성 설치

```bash
# Go workspace 동기화
go work sync

# Frontend 의존성 설치 (루트에서 실행, 모든 workspace 포함)
pnpm install
```

### 2. 개발 서버 실행

프론트엔드와 백엔드를 별도 터미널에서 각각 실행합니다.

**백엔드** (터미널 1)

```bash
# 먼저 바이너리 빌드 후 실행 (port 8080)
SITE_MODE=demo make build_pharos
./bin/pharos server --config ./migrations/[SITE_MODE]/config/config.toml
```

**프론트엔드** (터미널 2)

```bash
# API_SERVER_URL로 백엔드 주소를 지정 (기본값은 vite.config.ts 내 하드코딩 주소)
API_SERVER_URL=http://localhost:8080 SITE_MODE=demo pnpm dev
```

프론트 개발 서버는 port 3000에서 실행되며, API 요청은 `API_SERVER_URL`로 프록시됩니다.

사용 가능한 SITE_MODE: `demo`, `catv`, `example`, `clickhouse`, `all`

각 SITE_MODE에 포함되는 Extension 목록은 `meta/site-config.json`에서 확인하세요.

### 3. 타입 자동 생성

스키마(`shared/schema/`) 변경 후 타입 재생성:

```bash
# Frontend TypeScript 타입 생성
pnpm generate:types

# Backend Go 타입 생성
cd shared && go run scripts/generate-types/main.go
```

## Windows 개발 환경

Go와 Node.js는 Windows에서 네이티브로 동작합니다. 단, `make` 명령과 Altibase 플러그인(CGO)은 WSL2가 필요합니다.

### Frontend + Go 개발 (PowerShell)

```powershell
go work sync
pnpm install
```

**백엔드** (터미널 1)

```powershell
$env:SITE_MODE="demo"
go generate ./core/frontend
go run meta/scripts/generate-metadata.go
go build -o bin/pharos.exe ./core/cmd/pharos
.\bin\pharos.exe server --config .\migrations\[SITE_MODE]\config\config.toml
```

**프론트엔드** (터미널 2)

```powershell
$env:API_SERVER_URL="http://localhost:8080"; $env:SITE_MODE="demo"; pnpm dev
```

### 빌드 (PowerShell — make 없이)

```powershell
# 1. Frontend 빌드
$env:SITE_MODE="demo"
go generate ./core/frontend

# 2. 메타데이터 생성 및 Go 빌드
go run meta/scripts/generate-metadata.go
go build -ldflags "-X 'ntels.com/pharos/core/pkg/version.Version=dev'" -o bin/pharos.exe ./core/cmd/pharos

# 3. 플러그인 빌드 (Altibase 제외, CGO 불필요)
go build -o bin/plugins/clickhouse.exe ./core/pkg/plugins/plugin/datasource/clickhouse/main/main.go
go build -o bin/plugins/postgresql.exe ./core/pkg/plugins/plugin/datasource/postgresql/main/main.go
go build -o bin/plugins/http-receiver.exe ./core/pkg/plugins/plugin/datasource/http-receiver/main/main.go
```

### Altibase 플러그인 / make 사용

WSL2 환경에서 일반 Linux 빌드 방법과 동일하게 진행합니다.

### 실행 (PowerShell)

```powershell
.\bin\pharos.exe server --config .\migrations\[SITE_MODE]\config\config.toml
```

## 빌드

### 전체 빌드 (Linux/macOS)

```bash
SITE_MODE=demo make build
```

빌드 흐름:
1. `generate_pharos` → Frontend(Vite) 빌드 + Go 바이너리에 embed
2. `build_prepare` → Go workspace 동기화, Extension import 파일 자동 생성
3. `build_pharos` → Go 바이너리 빌드
4. `build_plugin` → 데이터소스 플러그인 빌드

### 플러그인 개별 빌드

```bash
make build_plugin_altibase      # Altibase (CGO + unixODBC 필요)
make build_plugin_clickhouse
make build_plugin_postgresql
make build_plugin_http-receiver
```

### Docker 빌드

```bash
SITE_MODE=catv make docker
```

## 실행

```bash
./core/bin/run.sh        # 설정 파일 기반 실행
./bin/pharos server      # 직접 실행
```

## Extension 시스템

`meta/site-config.json`에서 각 SITE_MODE에 포함할 Extension을 정의합니다.

```json
{
  "siteModes": {
    "catv": {
      "extensions": ["catv", "dashboard/datasources/clickhouse", "demo"],
      "login-extension": "login/modern-card",
      "header-extension": "header/default"
    }
  }
}
```

자세한 내용은 [docs/EXTENSIONS.md](docs/EXTENSIONS.md)를 참고하세요.

## 버전 정책

태그 규칙: `vX.Y.Z-[SITE_MODE]` 또는 `vX.Y.Z-[SITE_MODE]-beta.N`

자세한 내용은 [docs/VERSIONING.md](docs/VERSIONING.md) 참고.

## 문서

- [docs/EXTENSIONS.md](docs/EXTENSIONS.md) - Extension 개발 가이드
- [docs/HOW_TO_ADD_PANEL.md](docs/HOW_TO_ADD_PANEL.md) - Panel 추가 방법
- [docs/NATS.md](docs/NATS.md) - NATS 통신 시스템
- [docs/VERSIONING.md](docs/VERSIONING.md) - 버전 관리 정책

## 문제 해결

- **Altibase 플러그인 빌드 오류**: unixODBC 개발 패키지 필요. `core/config/shared-library/unixODBC/` 경로 확인
- **Frontend 빌드 실패**: `SITE_MODE` 환경변수 설정 여부 확인
- **Extension 미로드**: `meta/site-config.json`에 Extension이 등록되어 있는지 확인 후 `go work sync` 실행
