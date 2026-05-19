/**
 * Demo Extension for Pharos Frontend
 * 
 * Provides demo pages for table components.
 */

import { registerExtension } from '@pharos/core/extension-registry';
import type { Extension } from '@pharos/core/extension-registry';
import { TableTabsDemo } from './routes/table-tabs-demo';
import { SingleTableDemo } from './routes/single-table-demo';
import { TableSelectDemo } from './routes/table-select-demo';
import { RealtimeMenuDemo } from './routes/realtime-menu-demo';
import { SelectComponentsDemo } from './routes/select-components-demo';
import { DataProviderDemo } from './routes/data-provider-demo';

// Data Provider 추가
import { demoDataProvider, DEMO_PROVIDER_NAME } from './providers/demoDataProvider';
import React from 'react';
import {PresentationIcon} from 'lucide-react';

// Extension metadata
const metadata: Extension = {
  name: 'demo',
  version: '1.0.0',
  displayName: 'Demo Pages',
  description: 'Demo pages for table components',

  // Data Provider 등록
  dataProviders: {
    [DEMO_PROVIDER_NAME]: demoDataProvider,
  },

  pages: [
    {
      path: '/extensions/demo/table-tabs',
      component: TableTabsDemo,
      title: 'Table Tabs Demo',
      requireAuth: true,
    },
    {
      path: '/extensions/demo/single-table',
      component: SingleTableDemo,
      title: 'Single Table Demo',
      requireAuth: true,
    },
    {
      path: '/extensions/demo/table-select',
      component: TableSelectDemo,
      title: 'Table Select Demo',
      requireAuth: true,
    },
    {
      path: '/extensions/demo/realtime-menu',
      component: RealtimeMenuDemo,
      title: 'Realtime Menu Demo',
      requireAuth: true,
    },
    {
      path: '/extensions/demo/select-components',
      component: SelectComponentsDemo,
      title: 'Select Components Demo',
      requireAuth: true,
    },
    {
      path: '/extensions/demo/data-provider',
      component: DataProviderDemo,
      title: 'Data Provider Demo',
      requireAuth: true,
    },
  ],

  menuItems: [
    {
      label: 'demo',
      icon: React.createElement(PresentationIcon, {className: 'h-4 w-4'}),
    },
    {
      label: 'table-tabs',
      path: '/extensions/demo/table-tabs',
      icon: 'Layers',
    },
    {
      label: 'single-table',
      path: '/extensions/demo/single-table',
    },
    {
      label: 'table-select',
      path: '/extensions/demo/table-select',
    },
    {
      label: 'realtime-menu',
      path: '/extensions/demo/realtime-menu',
    },
    {
      label: 'select-components',
      path: '/extensions/demo/select-components',
    },
    {
      label: 'data-provider',
      path: '/extensions/demo/data-provider',
      icon: 'Database',
    },
  ],
};

// Register the extension
registerExtension(metadata);
export * from './routes/realtime-menu-demo';

// Export metadata
export { metadata };
export * from './routes/table-tabs-demo';
export * from './routes/single-table-demo';
export * from './routes/table-select-demo';
export * from './routes/select-components-demo';
