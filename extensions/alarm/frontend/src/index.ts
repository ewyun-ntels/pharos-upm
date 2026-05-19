/**
 * Alarm Extension for Pharos Frontend
 *
 * Provides an Alerta alert dashboard by proxying the Alerta REST API
 * through the PHAROS backend (/alarm/alerts).
 */

import {registerExtension} from '@pharos/core/extension-registry';
import type {Extension} from '@pharos/core/extension-registry';
import React from 'react';
import {BellIcon} from 'lucide-react';
import {AlertaDashboard} from './routes/alerta-dashboard';

const metadata: Extension = {
  name: 'alarm',
  version: '1.0.0',
  displayName: 'Alarm',
  description: 'Alerta alert dashboard',

  pages: [
    {
      path: '/extensions/alarm/alerta',
      component: AlertaDashboard,
      title: 'Alerta',
      requireAuth: true,
    },
  ],

  menuItems: [
    {
      label: 'alarm',
      icon: React.createElement(BellIcon, {className: 'h-4 w-4'}),
    },
    {
      label: 'alerta',
      path: '/extensions/alarm/alerta',
    },
  ],
};

registerExtension(metadata);

export {metadata};
export * from './routes/alerta-dashboard';
