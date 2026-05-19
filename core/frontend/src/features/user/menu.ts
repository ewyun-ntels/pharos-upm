/**
 * User Feature Menu
 */

import React from 'react';
import { Users } from 'lucide-react';
import { MenuItem } from '@pharos/shared/features/extension';
import { USER_PROVIDER_NAME } from '@providers/user-provider';
import { PermissionKeysSchema } from '@pharos/shared/types/role';
import { menuRegistry } from '@features/menu/registry';

export const userMenuItems: MenuItem[] = [
  {
    label: 'users',
    path: '/users',
    icon: React.createElement(Users, { className: 'h-4 w-4' }),
    dataProviderName: USER_PROVIDER_NAME,
    canDelete: true,
    permission: (store) => store.hasPermission(PermissionKeysSchema.enum['role:user_read']),
  },
];

menuRegistry.registerAllAs('core', userMenuItems);
