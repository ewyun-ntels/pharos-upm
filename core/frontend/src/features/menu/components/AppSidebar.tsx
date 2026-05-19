/**
 * AppSidebar - 렌더링 전담 컴포넌트
 *
 * 메뉴 순서/계층/필터링은 MenuRegistry.applyOrder()에서 처리.
 * 이 컴포넌트는 menuRegistry.getSections()를 받아 렌더링만 담당.
 */

import * as React from 'react';
import { Link } from 'react-router-dom';
import {
  Sidebar,
  SidebarHeader,
  SidebarContent,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuAction,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  SidebarSeparator,
  SidebarFooter,
  SidebarTrigger,
} from '@pharos/shared/components/ui';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@pharos/shared/components/ui';
import { ChevronDown, ChevronsUpDown } from 'lucide-react';
import { menuRegistry, type ResolvedMenuItem, type SidebarSection } from '@features/menu/registry';
import type { MenuItem } from '@pharos/shared/features/extension';
import { useTranslation } from 'react-i18next';
import { useLocation } from 'react-router-dom';
import { getSidebarLogo } from '@/features/site';
import Clock from '@components/header/clock';
import {
  Avatar,
  AvatarFallback,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@pharos/shared/components/ui';
import { useGetIdentity, useLogout } from '@/lib/data-provider';
import { useUser, useAuthStore } from '@pharos/shared/features/auth';
import { UserPasswordDialog } from '@components/header/_components/userPasswordDialog';
import { MyInfoDialog } from '@components/header/_components/myInfoDialog';
import { useTheme } from '@providers/theme-provider';

type IdentityUser = {
  id: string;
  name: string;
  email: string;
  refreshToken?: string;
};



// ==========================================
// DynamicSubItems - useChildren hook 호출 컴포넌트
// hook 규칙상 조건부 호출 불가 → 별도 컴포넌트로 분리
// ==========================================

interface DynamicSubItemsProps {
  useChildren: () => MenuItem[];
  currentPath: string;
}

const DynamicSubItems = ({ useChildren, currentPath }: DynamicSubItemsProps) => {
  const children = useChildren();
  return (
    <>
      {children.map(child => (
        <SidebarMenuSubItem key={child.path ?? child.label}>
          <SidebarMenuSubButton
            asChild
            isActive={!!child.path && currentPath.startsWith(child.path)}
          >
            <Link to={child.path ?? '#'}>
              {typeof child.icon !== 'string' ? child.icon : null}
              <span>{child.label}</span>
            </Link>
          </SidebarMenuSubButton>
        </SidebarMenuSubItem>
      ))}
    </>
  );
};

// ==========================================
// DynamicMenuItems - flat 렌더링 (그룹 레벨)
// ==========================================

interface DynamicMenuItemsProps {
  useChildren: () => MenuItem[];
  currentPath: string;
}

const DynamicMenuItems = ({ useChildren, currentPath }: DynamicMenuItemsProps) => {
  const children = useChildren();
  return (
    <>
      {children.map(child => {
        if (child.children && child.children.length > 0) {
          const isChildActive = child.children.some(
            (sub) => !!sub.path && currentPath.startsWith(sub.path)
          );
          return (
            <Collapsible
              key={child.label}
              className="group/collapsible"
              defaultOpen={isChildActive}
            >
              <SidebarMenuItem>
                <CollapsibleTrigger asChild>
                  <SidebarMenuButton className="cursor-pointer group-data-[state=open]/collapsible:hover:bg-sidebar-accent">
                    {typeof child.icon !== 'string' ? child.icon : null}
                    <span>{child.label}</span>
                    <ChevronDown className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-180" />
                  </SidebarMenuButton>
                </CollapsibleTrigger>
                <CollapsibleContent>
                  <SidebarMenuSub>
                    {child.children.map((sub) => (
                      <SidebarMenuSubItem key={sub.path ?? sub.label}>
                        <SidebarMenuSubButton
                          asChild
                          isActive={!!sub.path && currentPath.startsWith(sub.path)}
                        >
                          <Link to={sub.path ?? '#'}>
                            {typeof sub.icon !== 'string' ? sub.icon : null}
                            <span>{sub.label}</span>
                          </Link>
                        </SidebarMenuSubButton>
                      </SidebarMenuSubItem>
                    ))}
                  </SidebarMenuSub>
                </CollapsibleContent>
              </SidebarMenuItem>
            </Collapsible>
          );
        }
        return (
          <SidebarMenuItem key={child.path ?? child.label}>
            <SidebarMenuButton
              asChild
              isActive={!!child.path && currentPath.startsWith(child.path)}
            >
              <Link to={child.path ?? '#'}>
                {typeof child.icon !== 'string' ? child.icon : null}
                <span>{child.label}</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        );
      })}
    </>
  );
};

// ==========================================
// Submenu Children Renderer
// ==========================================

interface SubmenuChildrenProps {
  item: ResolvedMenuItem;
  currentPath: string;
}

const SubmenuChildren = ({ item, currentPath }: SubmenuChildrenProps) => {
  const { t } = useTranslation();

  return (
    <SidebarMenuSub>
      {item.children.filter(child => child.isRegistered || child.useChildren).map(child => {
        if (child.useChildren) {
          return (
            <DynamicSubItems
              key={child.key}
              useChildren={child.useChildren}
              currentPath={currentPath}
            />
          );
        }
        const childLabel = t(child.displayName) || child.displayName;
        return (
          <SidebarMenuSubItem key={child.key}>
            <SidebarMenuSubButton
              asChild
              isActive={!!child.path && currentPath.startsWith(child.path)}
            >
              <Link to={child.path ?? '#'}>
                {child.icon}
                <span>{childLabel}</span>
              </Link>
            </SidebarMenuSubButton>
          </SidebarMenuSubItem>
        );
      })}
      {item.useChildren && (
        <DynamicSubItems useChildren={item.useChildren} currentPath={currentPath} />
      )}
    </SidebarMenuSub>
  );
};

// ==========================================
// Collapsible Menu (path 없음 - 토글 전용)
// ==========================================

interface CollapsibleToggleMenuProps {
  item: ResolvedMenuItem;
  label: string;
  isChildActive: boolean;
  currentPath: string;
}

const CollapsibleToggleMenu = ({ item, label, isChildActive, currentPath }: CollapsibleToggleMenuProps) => {
  return (
    <Collapsible
      asChild
      defaultOpen={isChildActive}
      className="group/collapsible"
    >
      <SidebarMenuItem>
        <CollapsibleTrigger asChild>
          <SidebarMenuButton className="cursor-pointer group-data-[state=open]/collapsible:hover:bg-sidebar-accent">
            {item.icon}
            <span>{label}</span>
            <ChevronDown className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-180" />
          </SidebarMenuButton>
        </CollapsibleTrigger>
        <CollapsibleContent>
          <SubmenuChildren item={item} currentPath={currentPath} />
        </CollapsibleContent>
      </SidebarMenuItem>
    </Collapsible>
  );
};

// ==========================================
// Collapsible Menu (path 있음 - 링크 + 토글)
// ==========================================

interface CollapsibleLinkMenuProps {
  item: ResolvedMenuItem;
  label: string;
  isActive: boolean;
  isChildActive: boolean;
  currentPath: string;
}

const CollapsibleLinkMenu = ({ item, label, isActive, isChildActive, currentPath }: CollapsibleLinkMenuProps) => {
  return (
    <Collapsible
      asChild
      defaultOpen={isActive || isChildActive}
      className="group/collapsible"
    >
      <SidebarMenuItem>
        <SidebarMenuButton asChild isActive={isActive && !isChildActive}>
          <Link to={item.path!}>
            {item.icon}
            <span>{label}</span>
          </Link>
        </SidebarMenuButton>
        <CollapsibleTrigger asChild>
          <SidebarMenuAction className="data-[state=open]:rotate-180">
            <ChevronDown />
          </SidebarMenuAction>
        </CollapsibleTrigger>
        <CollapsibleContent>
          <SubmenuChildren item={item} currentPath={currentPath} />
        </CollapsibleContent>
      </SidebarMenuItem>
    </Collapsible>
  );
};

// ==========================================
// Simple Link Menu
// ==========================================

interface SimpleLinkMenuProps {
  item: ResolvedMenuItem;
  label: string;
  isActive: boolean;
}

const SimpleLinkMenu = ({ item, label, isActive }: SimpleLinkMenuProps) => {
  return (
    <SidebarMenuItem>
      <SidebarMenuButton asChild isActive={isActive}>
        <Link to={item.path ?? '#'}>
          {item.icon}
          <span>{label}</span>
        </Link>
      </SidebarMenuButton>
    </SidebarMenuItem>
  );
};

// ==========================================
// MenuItemNode - 라우터
// ==========================================

interface MenuItemNodeProps {
  item: ResolvedMenuItem;
  currentPath: string;
}

const MenuItemNode = ({ item, currentPath }: MenuItemNodeProps) => {
  const { t } = useTranslation();

  const hasChildren = item.children.length > 0 || !!item.useChildren;

  // 등록되지 않은 항목: 자식이 없으면 숨김 (익스텐션 미로드)
  // 자식이 있으면 컨테이너(그룹)로 표시
  if (!item.isRegistered && !hasChildren) return null;

  const label = t(item.displayName) || item.displayName;
  const isActive = !!item.path && currentPath.startsWith(item.path);
  const isChildActive = item.children.some(c => !!c.path && currentPath.startsWith(c.path));

  // flat: useChildren 결과를 그룹 레벨에 바로 표현 (Collapsible 없음)
  if (item.flat && item.useChildren) {
    return <DynamicMenuItems useChildren={item.useChildren} currentPath={currentPath} />;
  }

  // Case 1: Collapsible 메뉴 (children 있음)
  if (hasChildren) {
    // path 없음 → 토글 전용
    if (!item.path) {
      return <CollapsibleToggleMenu item={item} label={label} isChildActive={isChildActive} currentPath={currentPath} />;
    }
    // path 있음 → 링크 + 토글
    return <CollapsibleLinkMenu item={item} label={label} isActive={isActive} isChildActive={isChildActive} currentPath={currentPath} />;
  }

  // Case 2: 단순 링크
  return <SimpleLinkMenu item={item} label={label} isActive={isActive} />;
};

// ==========================================
// DynamicFlatSection
// flat+useChildren 아이템만으로 구성된 섹션에서
// children이 비어있을 때 섹션 전체(라벨 포함)를 숨김
// ==========================================

interface DynamicFlatSectionProps {
  section: SidebarSection;
  useChildrenFn: () => MenuItem[];
  currentPath: string;
}

const DynamicFlatSection = ({ section, useChildrenFn, currentPath }: DynamicFlatSectionProps) => {
  const children = useChildrenFn();
  if (children.length === 0) return null;

  return (
    <>
      {section.separator && <SidebarSeparator className="w-auto!" />}
      <SidebarGroup>
        {section.label && (
          <SidebarGroupLabel>{section.label}</SidebarGroupLabel>
        )}
        <SidebarMenu>
          {children.map(child => {
            if (child.children && child.children.length > 0) {
              const isChildActive = child.children.some(
                (sub) => !!sub.path && currentPath.startsWith(sub.path)
              );
              return (
                <Collapsible
                  key={child.label}
                  className="group/collapsible"
                  defaultOpen={isChildActive}
                >
                  <SidebarMenuItem>
                    <CollapsibleTrigger asChild>
                      <SidebarMenuButton className="cursor-pointer group-data-[state=open]/collapsible:hover:bg-sidebar-accent">
                        {typeof child.icon !== 'string' ? child.icon : null}
                        <span>{child.label}</span>
                        <ChevronDown className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-180" />
                      </SidebarMenuButton>
                    </CollapsibleTrigger>
                    <CollapsibleContent>
                      <SidebarMenuSub>
                        {child.children.map((sub) => (
                          <SidebarMenuSubItem key={sub.path ?? sub.label}>
                            <SidebarMenuSubButton
                              asChild
                              isActive={!!sub.path && currentPath.startsWith(sub.path)}
                            >
                              <Link to={sub.path ?? '#'}>
                                {typeof sub.icon !== 'string' ? sub.icon : null}
                                <span>{sub.label}</span>
                              </Link>
                            </SidebarMenuSubButton>
                          </SidebarMenuSubItem>
                        ))}
                      </SidebarMenuSub>
                    </CollapsibleContent>
                  </SidebarMenuItem>
                </Collapsible>
              );
            }
            return (
              <SidebarMenuItem key={child.path ?? child.label}>
                <SidebarMenuButton
                  asChild
                  isActive={!!child.path && currentPath.startsWith(child.path)}
                >
                  <Link to={child.path ?? '#'}>
                    {typeof child.icon !== 'string' ? child.icon : null}
                    <span>{child.label}</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            );
          })}
        </SidebarMenu>
      </SidebarGroup>
    </>
  );
};

// ==========================================
// AppSidebar
// ==========================================

export const AppSidebar = () => {
  const { pathname } = useLocation();
  const ref = React.useRef<HTMLDivElement>(null);

  const peKey = 'pwExp-dialog';
  const [dropOpenState, setDropOpenState] = React.useState(false);
  const [cpDialogOpen, setCpDialogOpen] = React.useState(false);
  const [miDialogOpen, setMiDialogOpen] = React.useState(false);
  const { mutate: logout } = useLogout();
  const { data: identity } = useGetIdentity<IdentityUser>();
  const authUser = useUser();
  const { setTheme, resolvedTheme } = useTheme();

  // 권한 기반 메뉴 필터링 (authUser 변경 시 재계산)
  const sections = React.useMemo(() => {
    const store = useAuthStore.getState();
    return menuRegistry.getSections()
      .map(section => ({
        ...section,
        items: section.items.filter(item => {
          if (!item.permission) return true;
          return item.permission(store);
        }),
      }))
      .filter(section => section.items.length > 0);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [authUser]);

  const handleLogout = () => {
    logout();
  };

  React.useEffect(() => {
    const pDialog = localStorage.getItem(peKey);
    if (authUser && authUser.password_expired_at && !pDialog) {
      const currentDate = new Date().getTime();
      const expirationDate = authUser.password_expired_at.getTime();
      const daysUntilExpiration = Math.ceil(
        (expirationDate - currentDate) / (1000 * 60 * 60 * 24),
      );
      if (daysUntilExpiration <= 5) {
        setDropOpenState(false);
        setCpDialogOpen(true);
      }
    }
  }, [authUser]);

  const handleCloseDialog = (type: string) => {
    if (type === 'cp') {
      setCpDialogOpen(false);
    } else if (type === 'mi') {
      setMiDialogOpen(false);
    }
    localStorage.setItem(peKey, 'true');
  };

  return (
    <>
    <Sidebar collapsible="offcanvas" className='border-r'>
      <SidebarHeader>
        <div ref={ref} className="flex items-center justify-between gap-2">
          <Link to="/home" className="hidden item-center gap-3 md:flex">
            {(() => {
              const LogoComponent = getSidebarLogo();
              return LogoComponent ? <LogoComponent className="h-7 w-auto" /> : null;
            })()}
          </Link>
          <SidebarTrigger className="shrink-0" />
        </div>
      </SidebarHeader>
      <SidebarContent>
        {sections.map((section, i) => {
          // flat+useChildren 아이템만으로 구성된 섹션:
          // children이 비어있으면 라벨을 포함한 섹션 전체를 숨김
          const flatDynamicItems = section.items.filter(item => item.flat && item.useChildren);
          if (flatDynamicItems.length > 0 && flatDynamicItems.length === section.items.length) {
            return (
              <DynamicFlatSection
                key={i}
                section={section}
                useChildrenFn={flatDynamicItems[0].useChildren!}
                currentPath={pathname}
              />
            );
          }
          return (
            <React.Fragment key={i}>
              {section.separator && <SidebarSeparator className="w-auto!" />}
              <SidebarGroup>
                {section.label && (
                  <SidebarGroupLabel>{section.label}</SidebarGroupLabel>
                )}
                <SidebarMenu>
                  {section.items.map(item => (
                    <MenuItemNode key={item.key} item={item} currentPath={pathname} />
                  ))}
                </SidebarMenu>
              </SidebarGroup>
            </React.Fragment>
          );
        })}
      </SidebarContent>
      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu open={dropOpenState} onOpenChange={(open) => setDropOpenState(open)}>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton
                  size="lg"
                  className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                >
                  <Avatar className="h-8 w-8 rounded-lg">
                    <AvatarFallback className="rounded-lg text-xs font-semibold">
                      {(identity?.name || identity?.email || '?').charAt(0).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                  <div className="grid flex-1 text-left text-sm leading-tight">
                    <span className="truncate font-semibold">
                      {identity?.name || identity?.email}
                    </span>
                    <Clock />
                  </div>
                  <ChevronsUpDown className="ml-auto size-4 text-muted-foreground" />
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent side="top" align="start" className="min-w-56 rounded-lg">
                <DropdownMenuItem
                  className="cursor-pointer"
                  onClick={() => { setDropOpenState(false); setMiDialogOpen(true); }}
                >
                  My Info
                </DropdownMenuItem>
                <DropdownMenuItem
                  className="cursor-pointer"
                  onClick={() => { setDropOpenState(false); setCpDialogOpen(true); }}
                >
                  Change password
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  className="cursor-pointer"
                  onClick={() => setTheme(resolvedTheme === 'dark' ? 'light' : 'dark')}
                >
                  <span className="flex-1">{resolvedTheme === 'dark' ? 'Light' : 'Dark'}</span>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem className="cursor-pointer" onClick={handleLogout}>
                  Sign out
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
    <UserPasswordDialog
      open={cpDialogOpen}
      onOpenChange={() => handleCloseDialog('cp')}
      dialogTitle="Change password"
      passwordExpiryDate={
        authUser?.password_expired_at
          ? Math.floor(
            new Date(authUser.password_expired_at).getTime() / 1000
          )
          : 0
      }
    />
    <MyInfoDialog
      open={miDialogOpen}
      onOpenChange={() => handleCloseDialog('mi')}
      dialogTitle="My Info"
    />
    </>
  );
};
