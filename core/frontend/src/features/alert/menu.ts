/**
 * Alert Feature Menu
 */

import React from 'react';
import { Bell } from 'lucide-react';
import { MenuItem } from '@pharos/shared/features/extension';
import { ALERT_PROVIDER_NAME } from '@providers/alert-provider/types';
import { PermissionKeysSchema } from '@pharos/shared/types/role';
import { menuRegistry } from '@features/menu/registry';
export const alertMenuItems: MenuItem[] = [
  {
    label: 'alert',
    path: '/alert',
    icon: React.createElement(Bell, { className: 'h-4 w-4' }),
    dataProviderName: ALERT_PROVIDER_NAME,
    canDelete: true,
    permission: (store) => store.hasPermission(PermissionKeysSchema.enum['role:alert_read']),
  },
];

menuRegistry.registerAllAs('core', alertMenuItems);
