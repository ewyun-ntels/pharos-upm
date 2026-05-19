# 프로젝트 개요

## 목적

**Pharos**는 SK Broadband CATV 인프라를 위한 STB(셋톱박스) 품질 모니터링 및 원격 제어 플랫폼입니다.  
Go 백엔드와 React 프론트엔드를 단일 바이너리로 배포하며, `SITE_MODE` 환경변수를 통해 여러 배포 형태를 하나의 코드베이스에서 관리합니다.

## 주요 기능

| 기능              | 설명                                                             |
| ----------------- | ---------------------------------------------------------------- |
| STB 원격 제어     | Kubernetes Job을 통한 STB 대규모 원격 명령 배포                  |
| 실시간 품질 수집  | UDP(gnet)로 STB 5종 메시지 수신 → ClickHouse 저장                |
| 기상 데이터 연계  | 기상청 FTP 자동 수집 → ClickHouse (CATV 신호 품질 상관관계 분석) |
| 대시보드 시스템   | Grafana 유사 대시보드 (패널 플러그인 시스템, 변수, 즐겨찾기)     |
| ETL 파이프라인    | NATS JetStream → Elasticsearch 인덱싱                            |
| Prometheus 메트릭 | STB 상태·UDP 수신·제어 작업 커스텀 메트릭 노출                   |
| 사용자·권한 관리  | OAuth2(fosite) + JWT RS256 + Casbin RBAC                         |

## 기술 스택 요약

### 백엔드

| 분류            | 기술                                                                   |
| --------------- | ---------------------------------------------------------------------- |
| 언어            | Go 1.25+                                                               |
| HTTP 프레임워크 | Gin                                                                    |
| 인증            | OAuth2 (ory/fosite), JWT RS256                                         |
| 권한            | Casbin v3 (RBAC)                                                       |
| 메시징          | NATS JetStream (내장 서버 + 클라이언트)                                |
| WebSocket       | Centrifuge                                                             |
| UDP 서버        | gnet v2 (고성능 네트워크 프레임워크)                                   |
| DB              | ClickHouse, SQLite, PostgreSQL, Altibase(ODBC), Vertica, Elasticsearch |
| DB 마이그레이션 | goose v3                                                               |
| 스케줄러        | robfig/cron v3                                                         |
| 플러그인        | HashiCorp go-plugin                                                    |
| Kubernetes      | k8s.io/client-go (In-Cluster)                                          |

### 프론트엔드

| 분류        | 기술                         |
| ----------- | ---------------------------- |
| 언어        | TypeScript 5                 |
| 프레임워크  | React 19                     |
| 번들러      | Vite 8                       |
| CSS         | Tailwind CSS v4              |
| 라우팅      | React Router v7              |
| 서버 상태   | TanStack Query v5            |
| 테이블      | TanStack Table v8            |
| 폼          | React Hook Form + Zod        |
| 차트        | Recharts + 커스텀 TimeSeries |
| 에디터      | Monaco Editor, Ace Editor    |
| i18n        | i18next (한국어/영어)        |
| 패키지 관리 | pnpm Workspaces              |

## SITE_MODE 별 구성

`meta/site-config.json`에서 각 모드의 Extension 조합을 정의합니다.

| SITE_MODE    | 포함 Extension                                              | 비고                        |
| ------------ | ----------------------------------------------------------- | --------------------------- |
| `catv`       | `catv/business`, `dashboard/datasources/clickhouse`, `demo` | SK Broadband CATV 운영 환경 |
| `default`    | `dashboard/datasources/clickhouse`, `demo`                  | 기본 Pharos 플랫폼          |
| `demo`       | `demo`, `dashboard/panels/panel-example`                    | 컴포넌트 데모               |
| `example`    | `example`, `dashboard/panels/panel-example`                 | Extension 개발 예제         |
| `clickhouse` | `dashboard/datasources/clickhouse`                          | ClickHouse 단독             |
| `all`        | 모든 Extension                                              | 전체 통합                   |

## 모노레포 구조 요약

```
pharos/
├── core/                   # Go 핵심 모듈 + React 앱 셸
├── extensions/
│   ├── catv/business/      # CATV 도메인 비즈니스 로직 (Go + React)
│   ├── catv/login/         # CATV 브랜딩 로그인 페이지
│   ├── dashboard/datasources/clickhouse/  # ClickHouse 데이터소스 플러그인
│   ├── dashboard/panels/   # 대시보드 패널 플러그인
│   ├── demo/               # 컴포넌트 데모
│   └── example/            # Extension 개발 템플릿
├── shared/                 # 공유 Go 타입 + React 컴포넌트 라이브러리
├── meta/                   # 빌드 메타데이터 및 코드 생성 스크립트
├── migrations/catv_quality/ # DB 마이그레이션 SQL, 대시보드 JSON, 다이어그램
├── plugins/                # 외부 데이터소스 플러그인 빌드 (별도 바이너리)
├── third_party/            # 포크된 서드파티 라이브러리 (goflow, GoVisual)
└── tools/                  # 운영 도구 (migration runner, simulator)
```

## 빌드 산출물

| 파일                        | 설명                                  |
| --------------------------- | ------------------------------------- |
| `bin/pharos`                | 메인 바이너리 (프론트엔드 embed 포함) |
| `bin/plugins/altibase`      | Altibase ODBC 플러그인 (CGO 빌드)     |
| `bin/plugins/clickhouse`    | ClickHouse 플러그인                   |
| `bin/plugins/postgresql`    | PostgreSQL 플러그인                   |
| `bin/plugins/http-receiver` | HTTP Receiver 플러그인                |
