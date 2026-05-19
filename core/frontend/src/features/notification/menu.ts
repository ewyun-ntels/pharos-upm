/**
 * Notification Feature Menu
 */

import React from 'react';
import { BellRing } from 'lucide-react';
import { MenuItem } from '@pharos/shared/features/extension';
import { NOTIFICATION_PROVIDER_NAME } from '@providers/notification-provider/types';
import { PermissionKeysSchema } from '@pharos/shared/types/role';
import { menuRegistry } from '@features/menu/registry';
export const notificationMenuItems: MenuItem[] = [
  {
    label: 'notification',
    path: '/notification',
    icon: React.createElement(BellRing, { className: 'h-4 w-4' }),
    dataProviderName: NOTIFICATION_PROVIDER_NAME,
    canDelete: true,
    permission: (store) => store.hasPermission(PermissionKeysSchema.enum['role:notification_read']),
  },
];

menuRegistry.registerAllAs('core', notificationMenuItems);
