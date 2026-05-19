# My Awesome Plugin

외부 플러그인 예제입니다. 이 플러그인은 간단한 데이터 시각화를 제공합니다.

## 구조

```
my-awesome-plugin/
├── README.md              # 이 파일
├── package.json           # 플러그인 메타데이터
├── plugin.config.ts       # 플러그인 설정
├── src/
│   ├── MyAwesomePanel.tsx # 메인 패널 컴포넌트
│   ├── Options.tsx        # 옵션 UI 컴포넌트
│   ├── setParam.ts        # 데이터 변환 함수
│   └── types.ts           # 타입 정의
└── dist/                  # 빌드 결과물 (자동 생성)
```

## 설치 방법

### 1. 개발 모드 (로컬)

```bash
# 프로젝트 루트에서
cd plugins/panel/my-awesome-plugin
npm install
npm run build

# 또는 pnpm 사용
pnpm install
pnpm build
```

### 2. 플러그인 등록

`core/frontend/src/features/panels/registry/externalPlugins.ts`에 추가:

```typescript
import {panelPluginRegistry} from './PanelPluginRegistry';

// 개발 모드에서 로컬 플러그인 로드
if (process.env.NODE_ENV === 'development') {
  import('../../../../../plugins/panel/my-awesome-plugin/plugin.config')
    .then((module) => {
      panelPluginRegistry.register(module.myAwesomePlugin);
      console.log('✅ My Awesome Plugin registered');
    })
    .catch((error) => {
      console.error('❌ Failed to load My Awesome Plugin:', error);
    });
}
```

### 3. 프로덕션 배포

```bash
# 플러그인 빌드
npm run build

# dist 폴더를 CDN에 업로드
# 예: https://cdn.example.com/plugins/my-awesome-plugin/

# 런타임에 로드
const plugin = await loadPluginFromUrl(
  'https://cdn.example.com/plugins/my-awesome-plugin/index.js'
);
panelPluginRegistry.register(plugin);
```

## 사용 방법

1. Dashboard 편집 화면으로 이동
2. "Add Panel" 클릭
3. Panel 타입에서 "My Awesome Panel" 선택
4. 옵션 설정 후 저장

## 개발 가이드

### 필수 구현 항목

1. **plugin.config.ts** - 플러그인 메타데이터 및 entry point
2. **MyAwesomePanel.tsx** - 메인 패널 컴포넌트 (데이터 시각화)
3. **Options.tsx** - 옵션 UI (Panel Editor에서 사용)
4. **setParam.ts** - Form 데이터 ↔ Panel 데이터 변환

### 타입 정의

```typescript
export interface MyAwesomePanelProps {
  args?: ChartQueryArgs;
  query?: ChartQuery;
  resource?: string;
  dataProviderName?: string;
  chartOptions?: MyAwesomePanelOptions;
  // ... 기타 props
}

export interface MyAwesomePanelOptions {
  color?: string;
  size?: 'small' | 'medium' | 'large';
  showLabels?: boolean;
  // ... 커스텀 옵션
}
```

## 배포 옵션

### A. NPM 패키지
```bash
npm publish @your-org/pharos-plugin-my-awesome
```

### B. CDN
- 빌드 결과물을 CDN에 업로드
- URL로 로드

### C. 로컬 개발
- 이 저장소에 포함 (현재 방식)
- Git submodule로 관리

## 라이센스

MIT
