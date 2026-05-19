# Extensions

Pharos 프로젝트의 확장 기능(Extension)을 관리하는 디렉토리입니다.

## Extension 추가 방법

### 1. 디렉토리 생성

```bash
# 기본 구조
mkdir -p extensions/{category}/{extension-name}/frontend/src

# 예시
mkdir -p extensions/monitoring/metrics/frontend/src
mkdir -p extensions/catv/reports/frontend/src
```

**디렉토리 명명 규칙**:
- 소문자, 하이픈(`-`) 사용 권장
- 슬래시(`/`)로 카테고리 구분 가능
- 디렉토리 경로가 **패키지 이름의 기준**이 됩니다

### 2. package.json 생성

`extensions/{your-path}/frontend/package.json`:

```json
{
  "name": "@pharos/extension-{your-path-with-dashes}",
  "version": "1.0.0",
  "private": true,
  "main": "src/index.ts",
  "types": "src/index.ts",
  "dependencies": {
    "@pharos/core": "workspace:*",
    "@pharos/shared": "workspace:*",
    "react": "^19.2.3",
    "react-dom": "^19.2.3"
  },
  "devDependencies": {
    "@types/react": "^19.2.7",
    "@types/react-dom": "^19.2.3",
    "typescript": "^5.9.3"
  }
}
```

**중요**: `name` 필드가 규칙과 맞지 않아도 괜찮습니다.
메타데이터 생성 시 **자동으로 수정**됩니다.

**자동 변환 규칙**:
```
extensions/monitoring/metrics → @pharos/extension-monitoring-metrics
extensions/catv/business      → @pharos/extension-catv-business
extensions/demo               → @pharos/extension-demo
```

### 3. 진입점 파일 생성

`extensions/{your-path}/frontend/src/index.ts`:

```typescript
import { registerExtension } from '@pharos/core/extension-registry';

registerExtension({
  name: 'my-extension',
  version: '1.0.0',
  displayName: 'My Extension',
  description: 'Description',
  pages: [
    {
      path: '/extensions/my-extension/page',
      component: () => import('./pages/MyPage'),
      title: 'My Page',
      requireAuth: true,
    },
  ],
  menuItems: [
    {
      label: 'my-menu',
      path: '/extensions/my-extension/page',
    },
  ],
});
```

### 4. configs/{mode}.json에 등록

`meta/configs/{mode}.json`의 원하는 모드에 extension을 추가합니다:

```json
{
  "extensions": [
    "your-category/your-extension"
  ],
  "menu-order": [
    {
      "label": "My Section",
      "items": [
        {
          "extension": "your-extension",
          "name": "your-menu-id",
          "display_name": "Your Menu"
        }
      ]
    }
  ]
}
```

### 5. 메타데이터 생성

```bash
SITE_MODE=mymode pnpm exec tsx meta/scripts/generate-metadata.ts
```

메타데이터 생성 시 자동으로:
- ✅ package.json의 name이 규칙에 맞게 수정됨
- ✅ `meta/generated/frontend/vite-extensions.json` 업데이트 (alias 자동 반영)
- ✅ `meta/generated/frontend/extension-loader.ts`에 import 문 추가

### 6. TypeScript 설정 (선택사항)

Extension 내부에서 TypeScript를 사용하려면:

`extensions/{your-path}/frontend/tsconfig.json`:

```json
{
  "extends": "../../../tsconfig.base.json",
  "compilerOptions": {
    "outDir": "./dist",
    "rootDir": "./src"
  },
  "include": ["src"]
}
```

---

## 로고 등록

로고는 JSON 설정이 아닌 **코드로 등록**합니다. Login extension의 `index.ts`에서:

```typescript
import { registerSidebarLogo } from '@pharos/core/site-registry';
import { Logo } from './components/Logo';

registerSidebarLogo(Logo);
```

`Logo` 컴포넌트는 SVG를 직접 import하여 Vite가 번들링하도록 합니다:

```typescript
import logoSvg from '../../../../images/logo.svg';
import logoDarkSvg from '../../../../images/logo-dark.svg';
```

---

## 디렉토리 구조 예시

```
extensions/
├── README.md
├── demo/                              # 단일 레벨 extension
│   └── frontend/
│       ├── package.json              → @pharos/extension-demo
│       └── src/
│           └── index.ts
├── catv/                              # 카테고리별 그룹
│   ├── business/
│   │   └── frontend/
│   │       ├── package.json          → @pharos/extension-catv-business
│   │       └── src/
│   ├── login/
│   │   ├── images/                   # 파비콘, 로고 이미지
│   │   │   ├── favicon.svg
│   │   │   ├── logo.svg
│   │   │   └── logo-dark.svg
│   │   └── frontend/
│   │       ├── package.json          → @pharos/extension-catv-login
│   │       └── src/
│   │           ├── index.ts          # registerSidebarLogo 호출
│   │           └── components/
│   │               └── Logo.tsx
└── dashboard/
    └── datasources/
        └── clickhouse/
            └── frontend/
                ├── package.json      → @pharos/extension-dashboard-datasources-clickhouse
                └── src/
```

---

## Backend (Go) Extension go.mod 설정

Go 코드가 포함된 Backend Extension을 추가할 때는 `go.mod` 모듈 설정이 필요합니다.

### go.work 등록

`go.work` 파일의 `use` 블록에 새 모듈을 추가합니다:

```
use (
    ...
    ./extensions/your-category/your-extension
)
```

### go.mod 작성 규칙

```
module ntels.com/pharos/extensions/your-category/your-extension

go 1.26.0

require (
    ntels.com/pharos/core v0.0.0
    ntels.com/pharos/shared v0.0.0
)

replace (
    ntels.com/pharos/core => ../../core
    ntels.com/pharos/shared => ../../shared
)
```

**`replace` 블록이 필수인 이유**: `go.work`의 `use`에 등록된 워크스페이스 모듈은 `go.work`의 `replace`로 재지정할 수 없습니다. 로컬 모듈 간의 `replace`는 반드시 각 `go.mod` 파일 안에 있어야 합니다.

**버전은 `v0.0.0` 사용**: `go mod tidy`가 자동 생성하는 null 슈도버전 대신 명시적으로 `v0.0.0`을 사용합니다.

### catv/business ↔ core 순환 모듈 의존성

`catv/business` extension은 `core`와 **상호 모듈 의존성**이 있습니다:

- `catv/business/go.mod` → `core`를 require
- `core/go.mod` → `catv/business`를 require (`SITE_MODE=catv`일 때 blank import로 추가)

`replace` 블록이 있어야 프록시 조회 없이 로컬 경로로 해결됩니다.

---

## 주의사항

### ✅ 해야 할 것
1. **디렉토리 이름을 신중하게** 지으세요 (나중에 바꾸기 어렵습니다)
2. `meta/configs/{mode}.json`에 extension 경로 등록
3. 메타데이터 생성 실행

### ❌ 하지 않아도 되는 것
1. package.json의 `name` 필드를 수동으로 맞추기 (자동 수정됨)
2. vite.config.ts 수동 편집 (vite-extensions.json으로 자동 관리됨)
3. extension-loader.ts 수동 편집 (자동 생성됨)

---

## 문제 해결

### Q: extension이 로드되지 않아요
**A**: 다음을 확인하세요:
1. `meta/configs/{mode}.json`의 `extensions` 배열에 경로가 있는지
2. 메타데이터 생성을 실행했는지
3. `meta/generated/frontend/extension-loader.ts`에 import가 있는지

### Q: package.json name을 바꿨는데 다시 원래대로 돌아가요
**A**: 의도된 동작입니다. 패키지 이름을 바꾸려면 **디렉토리 이름을 변경**하세요.

### Q: TypeScript에서 extension을 import할 수 없어요
**A**: 메타데이터 생성 후 개발 서버를 재시작하세요.

---

## 참고

- Meta Generation: `meta/scripts/generate-metadata.ts`
- Extension Registry: `shared/frontend/src/extension-registry.ts`
- Menu Registry: `core/frontend/src/features/menu/registry/index.ts`
- Panel Registry: `core/frontend/src/features/dashboard/panels/registry/index.ts`
