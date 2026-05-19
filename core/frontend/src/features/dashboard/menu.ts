/**
 * Dashboard Feature Menu
 */

import React from 'react';
import { Folder, LayoutGrid } from 'lucide-react';
import { MenuItem } from '@pharos/shared/features/extension';
import { USER_PROVIDER_NAME } from '@providers/user-provider';
import { PermissionKeysSchema } from '@pharos/shared/types/role';
import { menuRegistry } from '@features/menu/registry';
import { useFavoritesStore } from './hooks/use-favorites-store';
import { useDashboardNavStore } from './hooks/use-dashboard-nav-store';

export const dashboardMenuItems: MenuItem[] = [
  {
    label: 'dashboard',
    path: '/dashboards',
    icon: React.createElement(LayoutGrid, { className: 'h-4 w-4' }),
    dataProviderName: USER_PROVIDER_NAME,
    canDelete: true,
    permission: (store) => store.hasPermission(PermissionKeysSchema.enum['role:dashboard_read']),
  },
  /**
   * starred: 즐겨찾기 단순 목록 (flat)
   * site-config에서 { "extension": "core", "name": "starred" } 으로 사용
   */
  {
    label: 'starred',
    useChildren: () => {
      const favorites = useFavoritesStore((state) => state.favorites);
      return favorites.map((fav) => ({
        label: fav.displayName,
        path: `/dashboards/${fav.id}`,
      }));
    },
  },
  /**
   * starred-tree: 즐겨찾기 폴더 계층 구조
   * site-config에서 { "extension": "core", "name": "starred-tree" } 으로 사용
   */
  {
    label: 'starred-tree',
    useChildren: () => {
      const favorites = useFavoritesStore((state) => state.favorites);
      const dashboards = useDashboardNavStore((state) => state.dashboards);
      const folders = useDashboardNavStore((state) => state.folders);

      const folderMap = new Map<string, { name: string; items: typeof favorites }>();
      const rootItems: typeof favorites = [];

      favorites.forEach((fav) => {
        const dashData = dashboards.find((d) => d.id === fav.id);
        const folderId = dashData?.folderId;
        if (folderId) {
          const folder = folders.find((f) => f.id === folderId);
          if (folder) {
            if (!folderMap.has(folderId)) {
              folderMap.set(folderId, { name: folder.name, items: [] });
            }
            folderMap.get(folderId)!.items.push(fav);
            return;
          }
        }
        rootItems.push(fav);
      });

      return [
        ...Array.from(folderMap.values()).map(({ name, items }) => ({
          label: name,
          icon: React.createElement(Folder, { className: 'w-4 h-4' }),
          children: items.map((item) => ({
            label: item.displayName,
            path: `/dashboards/${item.id}`,
          })),
        })),
        ...rootItems.map((item) => ({
          label: item.displayName,
          path: `/dashboards/${item.id}`,
        })),
      ];
    },
  },
];

menuRegistry.registerAllAs('core', dashboardMenuItems);
