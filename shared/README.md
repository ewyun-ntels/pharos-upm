# @pharos/shared

Pharos 프로젝트의 공유 타입 및 스키마 관리 패키지입니다.

## 개요

이 패키지는 JSON Schema를 단일 진실 공급원(Single Source of Truth)으로 사용하여 프론트엔드(TypeScript Zod)와 백엔드(Go)의 타입을 자동 생성합니다.

## 프로젝트 구조

```
shared/
├── README.md                          # 이 문서
├── run.sh                             # Go 타입 생성 스크립트
├── schema/                            # JSON Schema 정의 (원본)
│   ├── alert/                         # Alert 도메인
│   ├── badge/                         # Badge 도메인
│   ├── dashboard/                     # Dashboard 도메인
│   ├── notification/                  # Notification 도메인
│   ├── role/                          # Role 도메인
│   └── version/                       # Version 도메인
├── types/                             # Go 타입 (자동 생성)
│   ├── alert/types.go
│   ├── badge/types.go
│   ├── dashboard/types.go
│   ├── notification/types.go
│   ├── role/types.go
│   └── version/types.go
└── frontend/                          # TypeScript 워크스페이스 패키지
    ├── package.json                   # @pharos/shared 정의
    ├── tsconfig.json                  # TypeScript 설정
    ├── .gitignore                     # dist/ 제외
    ├── scripts/
    │   └── generate-types.js          # TypeScript 타입 생성
    └── types/                         # TypeScript 타입 (자동 생성)
        ├── alert/index.ts
        ├── badge/index.ts
        ├── dashboard/index.ts
        ├── notification/index.ts
        ├── role/index.ts
        └── version/index.ts
```

## 타입 생성

### TypeScript 타입 생성

```bash
# 프로젝트 루트에서
pnpm generate:types

# 또는 shared/frontend에서 직접
cd shared/frontend
pnpm generate:types
```

이 명령은 `shared/schema/{domain}/*.schema` → `shared/frontend/types/{domain}/index.ts` 변환을 수행합니다.

### Go 타입 생성

```bash
cd shared
./run.sh
```

이 스크립트는 `shared/schema/{domain}/*.schema` → `shared/types/{domain}/types.go` 변환을 수행합니다.

## 사용 방법

### TypeScript에서 사용

```typescript
// Dashboard 타입 import
import { 
  DashboardConfig, 
  Card, 
  ChartOptions,
  DataProvider 
} from '@pharos/shared/types/dashboard';

// Alert 타입 import
import { 
  AlertRule, 
  AlertEvent, 
  AlertSeverity 
} from '@pharos/shared/types/alert';

// Zod 스키마 사용
import { DashboardConfigSchema } from '@pharos/shared/types/dashboard';

// 사용 예시
const dashboard: DashboardConfig = {
  cards: [],
  description: "Example",
  displayName: "Dashboard",
  favorite: false,
  headerL: [],
  headerR: [],
  left: [],
  rangeStepOptions: {
    "1day": ["5m", "15m"],
    "1month": ["1h", "6h"],
    "1year": ["1d", "1w"]
  },
  title: "My Dashboard",
  type: "dashboard"
};

// Zod 검증
const validated = DashboardConfigSchema.parse(dashboard);
```

### 사용 가능한 도메인

각 도메인별로 `shared/frontend/types/{domain}/index.ts`에 타입이 자동 생성됩니다:

- **alert**: Alert 관련 타입 (AlertRule, AlertEvent, AlertSeverity 등)
- **badge**: Badge 관련 타입
- **dashboard**: Dashboard 관련 타입 (DashboardConfig, Card, ChartOptions 등)
- **notification**: Notification 관련 타입
- **role**: Role 관련 타입
- **version**: Version 관련 타입

모든 타입에는 Zod 스키마가 함께 제공되어 런타임 검증이 가능합니다.

## 새로운 타입 추가

1. `shared/schema/{domain}/` 디렉토리에 `.schema` 파일 추가
2. TypeScript 타입 생성: `pnpm generate:types`
3. Go 타입 생성: `cd shared && ./run.sh`
4. 생성된 타입 자동 인식됨

## 개발 워크플로우

1. **의존성 설치** (workspace 루트):
   ```bash
   pnpm install
   ```

2. **타입 생성** (스키마 변경 시):
   ```bash
   pnpm generate:types  # TypeScript
   cd shared && ./run.sh  # Go
   ```

3. **개발 모드** (core/frontend):
   ```bash
   pnpm dev
   ```

4. **빌드** (core/frontend):
   ```bash
   pnpm build
   ```

## 설정 정보

### Path Mapping

`core/frontend/tsconfig.json`:
```json
{
  "compilerOptions": {
    "paths": {
      "@pharos/shared": ["../../shared/frontend"],
      "@shared/*": ["../../shared/*"]
    }
  }
}
```

### Workspace 설정

`pnpm-workspace.yaml`:
```yaml
packages:
  - 'core/frontend'
  - 'shared/frontend'
```

`core/frontend/package.json`:
```json
{
  "dependencies": {
    "@pharos/shared": "workspace:*"
  }
}
```
