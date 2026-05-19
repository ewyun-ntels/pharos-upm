# Extension 개발 가이드

Extension은 `meta/site-config.json`에서 SITE_MODE별로 활성화됩니다. `extensions/example`을 참고 기준으로 쓰세요.

## 구조

```
extensions/my-extension/
├── frontend/
│   ├── package.json    # name: "@pharos/extension-my-extension"
│   └── src/
│       └── index.ts    # 패널/메뉴 등록 진입점
├── go.mod
└── extension.go        # Backend 진입점 (init()에서 등록)
```

## Frontend Extension

### 1. package.json

```json
{
  "name": "@pharos/extension-my-extension",
  "dependencies": {
    "@pharos/core": "workspace:*",
    "@pharos/shared": "workspace:*"
  }
}
```

### 2. src/index.ts

```typescript
import { panelPluginRegistry } from '@pharos/core/panel-registry';
import { registerExtension } from '@pharos/core/extension-registry';
import type { PanelPlugin } from '@pharos/core/panel-registry';
import { MyPanel } from './panels/MyPanel';
import { MyPanelOptions } from './panels/MyPanel';

const myPlugin: PanelPlugin = {
  info: {
    id: 'my-panel',
    label: 'My Panel',
    category: 'Charts',
  },
  component: MyPanel,
  editor: async () => ({
    OptionsComponent: MyPanelOptions,
    toPanelData: (formData) => formData,
    toFormData: (panelData) => panelData,
    getDefaults: () => ({ title: 'My Panel' }),
  }),
};

panelPluginRegistry.register(myPlugin);
registerExtension({ name: 'my-extension', version: '1.0.0', displayName: 'My Extension' });
```

### 3. 메뉴 추가 (선택)

```typescript
import { menuRegistry } from '@pharos/core/menu-registry';

menuRegistry.registerAll([
  { label: 'my-page', path: '/my-page', icon: ... },
]);
```

## Backend Extension

```go
// extension.go
package my_extension

import "ntels.com/pharos/core/pkg/some_registry"

func init() {
    some_registry.Register(...)
}
```

## SITE_MODE 등록

`meta/site-config.json`에 Extension 추가:

```json
{
  "siteModes": {
    "my-site": {
      "extensions": ["my-extension"],
      "login-extension": "login/default",
      "header-extension": "header/default"
    }
  }
}
```

`go.work`에 Go 모듈 추가:

```
use (
    ...
    ./extensions/my-extension
)
```

이후 `go work sync && pnpm install`.

## 개발 실행

```bash
SITE_MODE=my-site pnpm dev
```
