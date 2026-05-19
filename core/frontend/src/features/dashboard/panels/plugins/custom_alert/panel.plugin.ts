import {PanelPlugin, panelPluginRegistry} from '@pharos/core/panel-registry';
import type {CustomAlertPanelOptions} from './types';

export const alertPlugin: PanelPlugin<CustomAlertPanelOptions> = {
  info: {
    id: 'custom_alert',
    label: 'Alert',
    description: 'Alert table display',
    category: 'table',
  },
  component: () =>
    import('./alertTable').then((module) => module.default),
};

// 🔥 직접 등록
panelPluginRegistry.register(alertPlugin);
