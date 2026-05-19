# Panel Example Extension

커스텀 패널 개발을 위한 예시 Extension입니다.

## 포함된 패널

### 1. Simple Panel (`simple-panel`)

가장 기본적인 커스텀 패널 예시입니다.

**특징:**
- 데이터 소스 없이 동작하는 정적 패널
- 옵션을 통한 UI 커스터마이징
- 텍스트, 색상, 아이콘 등 기본 설정

**사용 사례:**
- 대시보드에 정적 안내 메시지 표시
- 로고나 브랜딩 요소 추가
- 간단한 상태 표시

### 2. Chart Panel Example (`chart-panel-example`)

데이터 소스를 사용하는 차트 패널 예시입니다.

**특징:**
- `useChartData` hook을 통한 시계열 데이터 로딩
- 로딩, 에러, 빈 데이터 상태 처리
- 데이터 통계 및 메트릭 표시
- 범례 및 데이터 포인트 옵션

**사용 사례:**
- 실시간 모니터링 데이터 시각화
- 커스텀 차트 구현
- 데이터 기반 위젯

## 패널 개발 가이드

### 1. 패널 컴포넌트 작성

```typescript
import type { PanelProps } from '@pharos/core/panel-registry';

export interface MyPanelOptions {
  title?: string;
  color?: string;
}

export function MyPanel(props: PanelProps<MyPanelOptions>) {
  const { title = 'My Panel', color = '#3b82f6' } = props.options || {};
  
  return (
    <div style={{ padding: '16px' }}>
      <h3 style={{ color }}>{props.title || title}</h3>
    </div>
  );
}
```

### 2. 옵션 에디터 작성

```typescript
import type { PanelEditorOptionsProps } from '@pharos/core/panel-registry';

export function MyPanelOptionsEditor(
  props: PanelEditorOptionsProps<MyPanelOptions>
) {
  const { value, onChange } = props;
  
  return (
    <div>
      <input
        value={value?.title || ''}
        onChange={(e) => onChange({ ...value, title: e.target.value })}
      />
    </div>
  );
}
```

### 3. 패널 플러그인 등록

```typescript
import { panelPluginRegistry } from '@pharos/core/panel-registry';
import type { PanelPlugin } from '@pharos/core/panel-registry';

export const myPanelPlugin: PanelPlugin = {
  info: {
    id: 'my-panel',
    label: 'My Panel',
    description: 'My custom panel',
    category: 'Custom',
  },
  component: MyPanel,
  editor: async () => ({
    OptionsComponent: MyPanelOptionsEditor,
    toPanelData: (formData) => ({ ...formData, type: 'my-panel' }),
    toFormData: (panelData) => panelData,
    getDefaults: () => ({
      chartOptions: {
        title: 'My Panel',
        color: '#3b82f6',
      },
    }),
  }),
};

panelPluginRegistry.register(myPanelPlugin);
```

### 4. Extension 등록

```typescript
import { registerExtension } from '@pharos/shared/extension-registry';

registerExtension({
  name: 'my-extension',
  version: '1.0.0',
  displayName: 'My Extension',
  description: 'My custom extension',
});
```

## 데이터 로딩 (useChartData)

```typescript
// pluginContext에서 useChartData hook 가져오기
const dataProvider = Array.isArray(props.dataProvider)
  ? props.dataProvider[0]
  : props.dataProvider;

const { loading, chartData } = props.pluginContext?.hooks?.useChartData?.({
  queries: dataProvider?.chartQuery || [],
  args: props.args || new Map(),
  resource: dataProvider?.resource || 'metrics',
  dataProviderName: dataProvider?.dataProviderName || 'dashboard',
  refetchInterval: props.refetchInterval,
  dashboardId: props.dashboardId,
  id: props.id,
  location: props.location,
}) || { loading: false, chartData: undefined };
```

## Props 설명

### PanelProps

- `id`: 패널 고유 ID
- `dashboardId`: 대시보드 ID
- `title`: 패널 제목
- `options`: 패널 옵션 (타입 파라미터로 지정)
- `dataProvider`: 데이터 소스 설정
- `args`: 쿼리 인자 (시간 범위, 필터 등)
- `pluginContext`: Core hooks 및 유틸리티 접근
- `location`: 패널 위치 정보
- `refetchInterval`: 데이터 갱신 주기

### PanelEditorOptionsProps

- `value`: 현재 옵션 값
- `onChange`: 옵션 변경 콜백
- `panel`: 전체 패널 데이터

## 파일 구조

```
extensions/panel-example/
├── frontend/
│   ├── src/
│   │   ├── index.ts                    # Extension 등록
│   │   └── panels/
│   │       ├── SimplePanel.tsx          # Simple Panel 컴포넌트
│   │       ├── SimplePanelOptionsEditor.tsx
│   │       ├── ChartPanel.tsx           # Chart Panel 컴포넌트
│   │       └── ChartPanelOptionsEditor.tsx
│   └── package.json
└── README.md
```

## 사용 방법

1. Extension이 자동으로 로드됨
2. 대시보드 편집 모드에서 "Add Panel" 클릭
3. "Examples" 카테고리에서 원하는 패널 선택
4. 패널 설정에서 옵션 커스터마이징
5. 데이터 소스 탭에서 쿼리 설정 (Chart Panel의 경우)

## 참고 자료

- [Extension 개발 가이드](../../docs/EXTENSION_REGISTRATION_GUIDE.md)
- [Panel 플러그인 시스템](../../docs/EXTENSIONS.md)
- [Core Hooks 사용법](../../docs/EXTENSION_CORE_USAGE.md)
