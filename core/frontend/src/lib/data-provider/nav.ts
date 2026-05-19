import { useLocation, useParams } from 'react-router-dom';
import { useMemo } from 'react';
import { menuRegistry } from '@features/menu';
import type { SidebarSection, ResolvedMenuItem } from '@features/menu/registry';
import type { Identifier } from '@pharos/shared/lib/data-provider/types';
import { useDashboardNavStore } from '@features/dashboard/hooks/use-dashboard-nav-store';

// ============================================================
// useBreadcrumb
// ============================================================

interface Breadcrumb {
  label: string;
  href?: string;
  isFolder?: boolean;
}

export function useBreadcrumb(currentLabel?: string) {
  const { pathname } = useLocation();
  const dashboards = useDashboardNavStore((state) => state.dashboards);
  const folders = useDashboardNavStore((state) => state.folders);

  const breadcrumbs = useMemo((): Breadcrumb[] => {
    // UUID → 대시보드 이름 매핑
    const dashboardNameMap = new Map(
      dashboards.map(d => [d.id, d.config.displayName || d.config.title])
    );
    // UUID → 대시보드의 folderId 매핑
    const dashboardFolderMap = new Map(
      dashboards.map(d => [d.id, d.folderId])
    );
    // folderId → 폴더 이름 매핑
    const folderNameMap = new Map(
      folders.map(f => [f.id, f.name])
    );

    const flattenItems = (items: ResolvedMenuItem[]): ResolvedMenuItem[] =>
      items.flatMap(item => [item, ...flattenItems(item.children)]);

    const allItems = menuRegistry.getSections().flatMap(section => flattenItems(section.items));

    let bestMatch: { label: string; path: string } | null = null;
    let bestMatchLength = 0;

    for (const item of allItems) {
      if (!item.path) continue;
      if (pathname === item.path || pathname.startsWith(item.path + '/')) {
        if (item.path.length > bestMatchLength) {
          bestMatchLength = item.path.length;
          bestMatch = { label: item.displayName, path: item.path };
        }
      }
    }

    if (!bestMatch) return [];

    const crumbs: Breadcrumb[] = [];

    if (pathname !== bestMatch.path) {
      crumbs.push({ label: bestMatch.label, href: bestMatch.path });
    } else {
      crumbs.push({ label: bestMatch.label });
    }

    const subPath = pathname.slice(bestMatch.path.length).replace(/^\//, '');
    if (subPath) {
      const segments = subPath.split('/');
      segments.forEach((segment, i) => {
        const isLast = i === segments.length - 1;
        
        if (dashboardFolderMap.has(segment)) {
          const folderId = dashboardFolderMap.get(segment);
          if (folderId && folderNameMap.has(folderId)) {
            crumbs.push({ label: folderNameMap.get(folderId)!, isFolder: true });
          }
        }

        const label = (isLast && currentLabel)
          ? currentLabel
          : dashboardNameMap.get(segment) ?? segment.charAt(0).toUpperCase() + segment.slice(1);
        
        let href = isLast
          ? undefined
          : bestMatch!.path + '/' + segments.slice(0, i + 1).join('/');
          
        // 단독 페이지가 아닌 부모 Settings 페이지의 하위 탭으로 편입된 메뉴들을 위한 라우트 우회 맵
        const tabRedirectMap: Record<string, string> = {
          variables: 'settings#variables',
        };

        if (!isLast && tabRedirectMap[segment]) {
          href = bestMatch!.path + '/' + segments.slice(0, i).join('/') + '/' + tabRedirectMap[segment];
        }

        crumbs.push({ label, href });
      });
    } else if (currentLabel) {
      crumbs[crumbs.length - 1] = { label: currentLabel };
    }

    return crumbs;
  }, [pathname, currentLabel, dashboards, folders]);

  return { breadcrumbs };
}

// ============================================================
// useMenu
// ============================================================

interface MenuItemMeta {
  label?: string;
  groupName?: string;
  displayName?: string;
}

interface MenuItemResult {
  key: string;
  route?: string;
  label: string;
  meta?: MenuItemMeta;
  children: MenuItemResult[];
}

export function useMenu() {
  const menuItems = useMemo((): MenuItemResult[] => {
    const sections = menuRegistry.getSections();

    const convertItem = (item: ResolvedMenuItem): MenuItemResult => ({
      key: item.key,
      route: item.path,
      label: item.displayName,
      meta: { label: item.displayName },
      children: item.children.map(convertItem),
    });

    return sections.flatMap((section: SidebarSection) =>
      section.items.map((item: ResolvedMenuItem) => ({
        key: item.key,
        route: item.path,
        label: item.displayName,
        meta: {
          label: item.displayName,
          groupName: section.label,
        },
        children: item.children.map(convertItem),
      })),
    );
  }, []);

  return { menuItems };
}

// ============================================================
// useResourceParams
// ============================================================

interface ResourceMeta {
  showLabel?: string;
  editLabel?: string;
  createLabel?: string;
  label?: string;
}

interface ResourceParams {
  name: string;
  meta?: ResourceMeta;
}

export function useResourceParams() {
  const { pathname } = useLocation();
  const params = useParams<{ id?: string }>();

  const resource = useMemo((): ResourceParams | undefined => {
    const allItems = menuRegistry.getAllItems();
    for (const item of allItems) {
      if (!item.path) continue;
      if (pathname === item.path || pathname.startsWith(item.path + '/')) {
        return { name: item.label, meta: {} };
      }
    }
    return undefined;
  }, [pathname]);

  const id: Identifier | undefined = params.id ?? undefined;

  return { resource, id };
}
