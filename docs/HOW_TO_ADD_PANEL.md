# 패널 플러그인 추가하기

기존 패널(`core/frontend/src/features/dashboard/panels/plugins/pie` 등)을 참고하세요.

## 구조

```
src/features/dashboard/panels/plugins/myPanel/
├── MyPanelCard.tsx     # 패널 컴포넌트
├── options.tsx         # 편집 옵션 UI
├── setParam.tsx        # toPanelData / toFormData 변환 로직
└── panel.plugin.ts     # 등록 (이 파일만 있으면 됨)
```

## panel.plugin.ts

```typescript
import React from 'react';
import { PanelPlugin, panelPluginRegistry } from '@pharos/core/panel-registry';
import { Options } from './options';
import { setParamToChart, setDefaultParam as toFormData } from './setParam';
import { MyPanelCard } from './MyPanelCard';

export const myPanelPlugin: PanelPlugin = {
  info: {
    id: 'myPanel',           // DB의 renderType과 일치해야 함
    label: 'My Panel',
    description: '...',
    category: 'timeseries',  // timeseries | gauge | distribution | table | stat
  },
  component: MyPanelCard,
  editor: async () => ({
    OptionsComponent: Options,
    toPanelData: setParamToChart,
    toFormData,
  }),
};

panelPluginRegistry.register(myPanelPlugin);
```

## corePlugins.ts에 등록

`src/features/dashboard/panels/registry/corePlugins.ts`에 import 추가:

```typescript
import '../plugins/myPanel/panel.plugin';
```
