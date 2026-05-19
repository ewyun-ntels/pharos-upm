/**
 * Settings Feature Menu
 */

import React from 'react';
import {Settings, ShieldCheck, UserCog} from 'lucide-react';
import {MenuItem} from '@pharos/shared/features/extension';
import {PermissionKeysSchema} from '@pharos/shared/types/role';
import {ROLE_PROVIDER_NAME} from '@providers/role-provider';
import {USER_PROVIDER_NAME} from '@providers/user-provider';
import {menuRegistry} from '@features/menu/registry';

export const settingsMenuItems: MenuItem[] = [
  {
    label: 'settings',
    icon: React.createElement(Settings, {className: 'h-4 w-4'}),
    // path 없음 → Collapsible 그룹으로만 작동 (클릭 불가)
    permission: (store) => store.hasPermission(PermissionKeysSchema.enum['role:super_admin']),
  },
  {
    label: 'roles',
    path: '/settings/roles',
    icon: React.createElement(ShieldCheck, {className: 'h-4 w-4'}),
    dataProviderName: ROLE_PROVIDER_NAME,
    permission: (store) => store.hasPermission(PermissionKeysSchema.enum['role:super_admin']),
  },
  {
    label: 'userFields',
    path: '/settings/user-fields',
    icon: React.createElement(UserCog, {className: 'h-4 w-4'}),
    dataProviderName: USER_PROVIDER_NAME,
    permission: (store) => store.hasPermission(PermissionKeysSchema.enum['role:super_admin']),
  },
];

menuRegistry.registerAllAs('core', settingsMenuItems);
