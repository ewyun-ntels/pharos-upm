import React from 'react';
import type { AuthState } from '../auth/types';

export interface MenuItem {
  label: string;
  path?: string;
  icon?: React.ReactNode;
  dataProviderName?: string;
  canDelete?: boolean;
  /**
   * Permission check function - determines if the menu item should be visible
   * @param store - AuthStore instance with hasPermission, hasAnyPermission, hasAllPermissions methods
   * @returns true if menu should be visible, false otherwise
   *
   * @example
   * // Single permission check
   * permission: (store) => store.hasPermission('role:user_read')
   *
   * @example
   * // OR condition (any permission)
   * permission: (store) => store.hasAnyPermission(['role:admin', 'role:super_admin'])
   *
   * @example
   * // AND condition (all permissions)
   * permission: (store) => store.hasAllPermissions(['role:a', 'role:b'])
   *
   * @example
   * // Complex logic
   * permission: (store) => {
   *   return store.hasPermission('role:advanced') &&
   *          store.user?.group === 'premium';
   * }
   */
  permission?: (store: AuthState) => boolean;
  extension?: string;
  useChildren?: () => MenuItem[];
  children?: MenuItem[];
}
