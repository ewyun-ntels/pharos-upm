import type { ComponentType } from 'react';

export interface SidebarLogoProps {
  className?: string;
}

let _logo: ComponentType<SidebarLogoProps> | null = null;

export function registerSidebarLogo(component: ComponentType<SidebarLogoProps>): void {
  _logo = component;
}

export function getSidebarLogo(): ComponentType<SidebarLogoProps> | null {
  return _logo;
}
