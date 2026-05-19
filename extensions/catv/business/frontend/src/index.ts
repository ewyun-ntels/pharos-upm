/**
 * CATV Extension for Pharos Frontend
 */

import {registerExtension} from '@pharos/core/extension-registry';
import type {Extension} from '@pharos/core/extension-registry';
import ControlPage from './features/control';
import React from 'react';
import {MonitorCheckIcon} from 'lucide-react';

// Import data provider
import {catvProvider, CATV_PROVIDER_NAME} from './providers';
import {CATV_PERMISSIONS} from './permissions';

// Extension metadata
const metadata: Extension = {
  name: 'catv',
  version: '1.0.0',
  displayName: 'CATV Management',
  description: 'CATV STB Control and Monitoring',

  // Data Providers
  dataProviders: {
    [CATV_PROVIDER_NAME]: catvProvider,
  },

  pages: [
    {
      path: '/extensions/catv/control',
      component: ControlPage,
      title: 'STB Control',
      requireAuth: true,
    },
  ],

  menuItems: [
    {
      label: 'stb-control',
      path: '/extensions/catv/control',
      icon: React.createElement(MonitorCheckIcon, {className: 'h-4 w-4'}),
      permission: (store) => store.hasPermission(CATV_PERMISSIONS.Read),
    },
  ],
};

// Register the extension
registerExtension(metadata);

// Export metadata
export {metadata};
export * from './features/control';
export * from './providers';
