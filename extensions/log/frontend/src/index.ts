import {registerExtension} from '@pharos/core/extension-registry';
import type {Extension} from '@pharos/core/extension-registry';
import React from 'react';
import {ScrollTextIcon} from 'lucide-react';
import {LogDashboard} from './routes/log-dashboard';

const metadata: Extension = {
  name: 'log',
  version: '1.0.0',
  displayName: 'Logs',
  description: 'OpenSearch log viewer',

  pages: [
    {
      path: '/extensions/log/logs',
      component: LogDashboard,
      title: 'Logs',
      requireAuth: true,
    },
  ],

  menuItems: [
    {
      label: 'log',
      icon: React.createElement(ScrollTextIcon, {className: 'h-4 w-4'}),
    },
    {
      label: 'logs',
      path: '/extensions/log/logs',
      icon: React.createElement(ScrollTextIcon, {className: 'h-4 w-4'}),
    },
  ],
};

registerExtension(metadata);

export {metadata};
export * from './routes/log-dashboard';
