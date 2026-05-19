# Demo Extension

Table 컴포넌트를 위한 데모 페이지를 제공하는 extension입니다.

## 구조

```
extensions/demo/
├── extension.go          # Go extension 등록
├── go.mod                # Go module 설정
└── frontend/
    ├── package.json      # NPM 패키지 설정
    ├── tsconfig.json     # TypeScript 설정
    └── src/
        ├── index.ts      # Extension 메타데이터 및 등록
        └── pages/
            ├── table-tabs-demo.tsx    # 탭 테이블 데모
            ├── single-table-demo.tsx  # 단일 테이블 데모
            └── table-select-demo.tsx  # 선택 테이블 데모
```

## 개발 환경 실행

```bash
# Demo extension과 함께 빌드
cd core/frontend
SITE_MODE=demo pnpm generate:extensions
SITE_MODE=demo pnpm dev
```

## 프로덕션 빌드

```bash
# Demo extension 포함
cd core/frontend
SITE_MODE=demo pnpm build

# 또는 Go에서
cd ../../
SITE_MODE=demo make build
```

## 메뉴 구조

- Demo (최상위 메뉴)
  - Table Tabs (`/extensions/demo/table-tabs`)
  - Single Table (`/extensions/demo/single-table`)
  - Table Select (`/extensions/demo/table-select`)

## 특징

- ✅ Path 기반 라우팅 (query string 불필요)
- ✅ React Router v6 네이티브 지원
- ✅ Shared 라이브러리 참조
- ✅ Core에서 extension 참조 방지 (격리된 구조)
- ✅ SITE_MODE로 선택적 빌드

## 의존성

- `@pharos/shared` - Shared components 및 utilities
- React Router v6 - 라우팅
