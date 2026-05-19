# 메뉴 추가 가이드

## Core 메뉴

`core/frontend/src/features/{feature}/menu.ts` 파일을 만들고 `menuRegistry`에 등록합니다.

```typescript
// src/features/my-feature/menu.ts
import React from 'react';
import { MyIcon } from 'lucide-react';
import { MenuItem } from '@pharos/shared/features/extension';
import { menuRegistry } from '@features/menu/registry';

export const myMenuItems: MenuItem[] = [
  {
    label: 'my-feature',            // i18n key
    path: '/my-feature',
    icon: React.createElement(MyIcon, { className: 'h-4 w-4' }),
    dataProviderName: 'myProvider',
    canDelete: false,
  },
];

menuRegistry.registerAllAs('core', myMenuItems);
```

이 파일을 `_refine_context.tsx` 또는 관련 진입점에서 import하면 자동 등록됩니다.

## Extension 메뉴

Extension의 `index.ts`에서 `registerExtension()` 호출 시 `menuItems`에 포함하거나, 직접 `menuRegistry`를 사용합니다.

```typescript
import { menuRegistry } from '@pharos/core/menu-registry';

menuRegistry.registerAll([
  { label: 'my-page', path: '/my-page', icon: ... },
]);
```

## 메뉴 순서 (order)

`meta/site-config.json`의 `menu-order` 필드에서 사이트별 메뉴 순서를 정의합니다.
