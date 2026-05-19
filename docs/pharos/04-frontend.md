# 프론트엔드 상세

## 패키지 구성 (pnpm Workspaces)

| 패키지명                        | 위치                                                    | 역할                               |
| ------------------------------- | ------------------------------------------------------- | ---------------------------------- |
| `@pharos/core`                  | `core/frontend/`                                        | 앱 셸, 핵심 페이지, 레지스트리     |
| `@pharos/shared`                | `shared/frontend/`                                      | 공유 컴포넌트 라이브러리           |
| `@pharos/catv-business`         | `extensions/catv/business/frontend/`                    | CATV STB 제어 UI                   |
| `@pharos/catv-login`            | `extensions/catv/login/frontend/`                       | CATV 로그인 페이지                 |
| `@pharos/clickhouse-datasource` | `extensions/dashboard/datasources/clickhouse/frontend/` | ClickHouse 데이터소스 설정         |
| `@pharos/panel-example`         | `extensions/dashboard/panels/panel-example/frontend/`   | 패널 플러그인 예제                 |
| `@pharos/demo`                  | `extensions/demo/frontend/`                             | 컴포넌트 데모                      |
| `@pharos/example`               | `extensions/example/frontend/`                          | Extension 개발 템플릿              |
| `@pharos/meta/extension-loader` | `meta/generated/frontend/`                              | 빌드 시 자동 생성 Extension loader |

## `@pharos/core` 구조

### 진입점 및 라우팅

```
core/frontend/src/main.tsx
  └─ App.tsx
       ├── @pharos/meta/extension-loader  (가장 먼저 import — Extension 등록)
       ├── QueryClientProvider (TanStack Query)
       ├── BrowserRouter (basename="/ui")
       └── RefineContext
           └── Routes
               ├── /login         → LoginPage
               ├── /home          → HomePage
               ├── /dashboards/*  → DashboardsPage
               ├── /alert/*       → AlertPage
               ├── /notification/* → NotificationPage
               ├── /users/*       → UsersPage
               ├── /settings/*    → Settings (Roles, UserFields)
               └── (Extension 동적 라우트)
                   예) /catv/stb-control → CATV ControlPage
```

### Features 디렉토리

```
src/features/
├── auth/          # 인증 상태, 로그인/로그아웃 액션
├── dashboard/     # 대시보드 뷰어, 편집기, 패널 레지스트리
│   ├── panels/    # 내장 패널 플러그인 (BarChart, TimeSeries, Stat, BarGauge 등)
│   ├── filters/   # 대시보드 변수(Variable) 필터 레지스트리
│   └── ...
├── extension/     # Extension 레지스트리 (페이지·메뉴·Provider 등록)
├── login/         # 로그인 Extension 레지스트리
├── menu/          # 메뉴 레지스트리 및 렌더링
├── site/          # 사이트 메타(로고, 제목, 아이콘) 레지스트리
├── user/          # 사용자 CRUD UI
├── alert/         # 알림 규칙 UI
└── notification/  # 알림 수신 UI
```

### 레지스트리 패턴

Extension은 빌드 시 자동 생성된 `extension-loader.ts`를 통해 다음 레지스트리에 등록합니다:

```typescript
// Extension 등록 예시 (catv/business/frontend/src/index.ts)
extensionRegistry.register({
  menus: [...],           // 사이드바 메뉴 항목
  pages: [...],           // React 라우트 + 컴포넌트
  providers: [...],       // 앱 레벨 Context Provider
})

panelRegistry.register({...})    // 대시보드 패널 플러그인
filterRegistry.register({...})   // 대시보드 Variable 필터
loginRegistry.register({...})    // 로그인 페이지 교체
siteRegistry.register({...})     // 로고/제목/아이콘 교체
```

### Providers

```
src/providers/
├── auth-provider/       # 인증 상태 초기화 및 토큰 갱신
├── alert-provider/      # 알림 규칙 훅 (useAlertRuleMutation 등)
├── notification-provider/ # 알림 수신 상태
├── dashboard-provider/  # 대시보드 컨텍스트
├── default-provider/    # 기본 데이터 Provider
├── plugin-provider/     # 데이터소스 플러그인 목록 캐시
├── role-provider/       # 권한 정보 로드
├── user-provider/       # 현재 사용자 정보
├── ui-config-provider/  # 서버 UI 설정(테마 등)
└── theme-provider/      # 다크/라이트 테마
```

### 컴포넌트

```
src/components/
├── controlBar/         # 대시보드 상단 제어바 (시간 범위, 새로고침)
├── editor/             # Monaco 기반 YAML/JSON 에디터 + K8s 리소스 템플릿
│   └── templates/      # K8s Deployment/Job/Service 등 YAML 템플릿
├── header/             # 헤더 (시계, 테마 전환, 사용자 정보)
├── breadcrumb/         # 브레드크럼 내비게이션
├── logView/            # XTerm.js 터미널 로그 뷰어
├── layout/             # 사이드바 + 콘텐츠 레이아웃
├── cell/               # 테이블 셀 렌더러 (age, label 등)
├── badge/              # 알림 뱃지
└── rjsf/               # React JSON Schema Form 커스텀 위젯
    (shadcn/ui 기반 BaseInputTemplate, CheckboxWidget, SelectWidget 등)
```

## `@pharos/shared` 컴포넌트 라이브러리

### UI Extension (커스텀 확장)

```
shared/frontend/src/components/ui-extension/
├── data-grid/              # 고기능 데이터 그리드
│   ├── DataGrid.tsx        # 기본 그리드
│   ├── DataGridInfinity.tsx # 무한 스크롤 그리드
│   ├── DataGridFilters.tsx # 필터 패널
│   └── hooks/              # useDataGridCore, useInfiniteScroll, useTableHeight 등
├── datetime-range/         # 날짜·시간 범위 선택기 (커스텀 파싱)
├── multiselect/            # 다중 선택 드롭다운
├── select/                 # 단일 선택 드롭다운 (shadcn Command 기반)
├── tag/                    # 태그 입력 (자동완성 포함)
├── data-table.tsx          # TanStack Table v8 래퍼
├── data-pagination.tsx     # 페이지네이션 UI
├── monaco-editor.tsx       # Monaco Editor 래퍼
├── variable-input.tsx      # Pharos 변수 문법 ($var) 인식 입력
├── scroll-table.tsx        # 고정 헤더 스크롤 테이블
└── loading-indicator.tsx   # 전체화면 로딩 인디케이터
```

### shadcn/ui 기본 컴포넌트

`shared/frontend/src/components/ui/` — Radix UI 기반 shadcn 컴포넌트 전체 세트  
(accordion, alert, avatar, badge, button, calendar, card, chart, checkbox, command, dialog, drawer, dropdown-menu, form, input, select, sheet, sidebar, table, tabs, toast, tooltip 등)

### Template 컴포넌트

```
shared/frontend/src/components/template/
├── login-default/          # 기본 로그인 레이아웃
├── login-modern-card/      # 카드형 로그인 레이아웃
├── table-tabs/             # 탭 전환 테이블 레이아웃 (CATV 제어 페이지에서 사용)
│   └── pagination/         # TablePagination 컴포넌트
├── table-select/           # 테이블 + 선택 레이아웃
└── editor-select/          # 에디터 + 선택 레이아웃
```

### Hooks

```
shared/frontend/src/hooks/
├── table-columns/          # useTableColumns — 컬럼 정의 + 내보내기 + 전역 필터 통합 훅
│   └── utils/              # columnsUtils, exportUtils, globalFilterUtils, selectionUtils
├── use-close-menu.ts       # 메뉴 닫기 훅
└── use-outsideclick-registry.ts # 바깥 클릭 감지 전역 레지스트리
```

### 유틸리티 (`shared/frontend/src/lib/`)

| 파일                  | 역할                                               |
| --------------------- | -------------------------------------------------- |
| `timestamp.ts`        | Unix timestamp ↔ luxon DateTime 변환               |
| `timeDuration.ts`     | `parse-duration` 기반 시간 단위 파싱 ("1h", "30m") |
| `mustache.ts`         | Mustache 템플릿 렌더링 (대시보드 변수 치환)        |
| `unitUtils.ts`        | 숫자 단위 변환 (bytes, bps 등)                     |
| `sheetType.ts`        | Sheet(슬라이드 패널) 타입 상수                     |
| `extension-loader.ts` | 런타임 Extension 동적 로딩                         |

## 인증 플로우 (프론트엔드)

```
1. 로그인 페이지 → POST /api/auth/token (ROPC)
   → { access_token, refresh_token, expires_in }

2. axios interceptor (src/lib/axios/index.ts)
   → 모든 API 요청에 Authorization: Bearer {access_token} 헤더 추가
   → 401 응답 시 refresh_token으로 자동 갱신
   → 갱신 실패 시 /login 리다이렉트

3. auth-store.ts (Zustand)
   → 전역 인증 상태 관리
   → useHasPermission(permissionKey) 훅으로 권한별 버튼/페이지 표시 제어
```

## 빌드 시 코드 생성

```bash
# 1. Frontend 메타데이터 생성 (SITE_MODE 기반 extension-loader.ts 생성)
tsx meta/scripts/generate-metadata.ts

# 2. Vite 빌드 (산출물: core/frontend/out/)
vite build

# 3. Go embed 처리
# core/frontend/embed.go:
#   //go:generate sh -c "..."  → pnpm build 실행
#   //go:embed out
#   var FS embed.FS
```

## 국제화 (i18n)

- `src/locales/ko/translation.json` — 한국어
- `src/locales/en/translation.json` — 영어
- `i18next-browser-languagedetector`로 브라우저 언어 자동 감지
- `react-i18next`의 `useTranslation()` 훅으로 컴포넌트에서 사용
