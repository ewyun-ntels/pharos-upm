/**
 * Menu Registry System
 *
 * Central registry for menu items from core features and extensions.
 * Key format: "{extension}:{label}" (e.g. "catv:stb-control", "core:dashboard")
 *
 * Ordering and hierarchy are applied once via applyOrder(MENU_ORDER).
 * AppSidebar simply calls getSections() for rendering.
 */

import type { MenuItem } from '@pharos/shared/features/extension';
import type { MenuOrderGroup, MenuOrderItem } from '@pharos/meta/menu-order';
import React from 'react';

// ==========================================
// 타입
// ==========================================

export interface ResolvedMenuItem {
  key: string;
  displayName: string;      // 번역 전 (i18n key 또는 plain text) - 렌더링 시 t()로 번역
  path?: string;
  icon?: React.ReactNode;
  isRegistered: boolean;    // MenuRegistry에 등록된 메뉴 = 클릭 가능
  useChildren?: () => MenuItem[];  // 동적 children hook (각 extension이 자신의 store 캡슐화)
  flat?: boolean;           // useChildren 결과를 그룹 레벨에 바로 표현 (Collapsible 없음)
  permission?: MenuItem['permission']; // 권한 체크 함수 (MenuItem에서 전달)
  children: ResolvedMenuItem[];
}

export interface SidebarSection {
  separator?: boolean;      // 섹션 앞에 SidebarSeparator 표시
  label?: string;           // SidebarGroupLabel 텍스트
  items: ResolvedMenuItem[];
}

// ==========================================
// MenuRegistry
// ==========================================

class MenuRegistry {
  private items: Map<string, MenuItem> = new Map();
  private sections: SidebarSection[] = [];
  private orderApplied = false;

  private generateKey(item: MenuItem): string {
    return item.extension ? `${item.extension}:${item.label}` : item.label;
  }

  register(item: MenuItem): void {
    const key = this.generateKey(item);
    if (this.items.has(key)) {
      console.warn(`[MenuRegistry] "${key}" is already registered. Overwriting.`);
    }
    this.items.set(key, item);
  }

  registerAll(items: MenuItem[]): void {
    items.forEach(item => this.register(item));
  }

  registerAllAs(extension: string, items: MenuItem[]): void {
    const ownedItems = items.map(item => ({ ...item, extension }));
    this.registerAll(ownedItems);
  }

  getAllItems(): MenuItem[] {
    return Array.from(this.items.values());
  }

  // ==========================================
  // 정렬 / 계층 처리 (초기화 시 한 번만 호출)
  // ==========================================

  /**
   * MenuOrderItem → ResolvedMenuItem 변환
   */
  private resolveItem(orderItem: MenuOrderItem): ResolvedMenuItem | null {
    if (orderItem.hidden) return null;

    const key = `${orderItem.extension}:${orderItem.name}`;
    const menuItem = this.items.get(key);
    const displayName = orderItem.display_name ?? menuItem?.label ?? orderItem.name;

    const children = (orderItem.items ?? [])
      .map(child => this.resolveItem(child))
      .filter((c): c is ResolvedMenuItem => c !== null);

    return {
      key,
      displayName,
      path: menuItem?.path,
      icon: menuItem && typeof menuItem.icon !== 'string' ? menuItem.icon : undefined,
      isRegistered: !!menuItem,
      useChildren: menuItem?.useChildren,
      flat: orderItem.flat,
      permission: menuItem?.permission,
      children,
    };
  }

  /**
   * MENU_ORDER (MenuOrderGroup[]) 기반으로 SidebarSection[] 생성.
   * site-config.json의 그룹 구조가 SidebarSection과 1:1 대응.
   *
   * MENU_ORDER에 없는 항목은 맨 아래 default 그룹에 자동 추가.
   */
  applyOrder(menuOrder: MenuOrderGroup[]): void {
    const orderedKeys = new Set<string>();

    // hidden 그룹 포함 모든 키 수집 (remaining 계산 시 hidden 항목이 누락되지 않도록)
    const collectKeys = (orderItem: MenuOrderItem) => {
      orderedKeys.add(`${orderItem.extension}:${orderItem.name}`);
      (orderItem.items ?? []).forEach(collectKeys);
    };
    menuOrder.forEach(group => group.items.forEach(collectKeys));

    const sections: SidebarSection[] = menuOrder
      .filter(group => !group.hidden)
      .map(group => {
        const items = group.items
          .map(orderItem => this.resolveItem(orderItem))
          .filter((item): item is ResolvedMenuItem => item !== null);

        return {
          separator: group.separator,
          label: group.label,
          items,
        } satisfies SidebarSection;
      });

    // MENU_ORDER에 없는 항목 → 맨 아래 default 그룹에 자동 추가
    const remaining = Array.from(this.items.values())
      .filter(item => !orderedKeys.has(this.generateKey(item)))
      .map(item => ({
        key: this.generateKey(item),
        displayName: item.label,
        path: item.path,
        icon: typeof item.icon !== 'string' ? item.icon : undefined,
        isRegistered: true,
        useChildren: item.useChildren,
        permission: item.permission,
        children: [],
      } satisfies ResolvedMenuItem));

    if (remaining.length > 0) {
      const last = sections[sections.length - 1];
      if (last) {
        last.items.push(...remaining);
      } else {
        sections.push({ items: remaining });
      }
    }

    this.sections = sections;
    this.orderApplied = true;
  }

  /**
   * 정렬/계층이 적용된 섹션 목록 반환.
   * applyOrder() 호출 전이면 경고 출력.
   */
  getSections(): SidebarSection[] {
    if (!this.orderApplied) {
      console.warn('[MenuRegistry] getSections() called before applyOrder().');
    }
    return this.sections;
  }

  clear(): void {
    this.items.clear();
    this.sections = [];
    this.orderApplied = false;
  }
}

export const menuRegistry = new MenuRegistry();
export type { MenuItem };
