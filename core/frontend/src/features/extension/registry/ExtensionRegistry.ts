/**
 * Extension Registry System
 *
 * Manages frontend extensions with runtime registration.
 * Extensions can contribute pages, widgets, menu items, and more.
 */

import * as React from 'react';
import { menuRegistry } from '@/features/menu/registry';
import type { MenuItem } from '@pharos/shared/features/extension';

// ===== Extension Types =====

export interface ExtensionMetadata {
  /** Unique identifier for the extension */
  name: string;
  /** Semantic version */
  version: string;
  /** Display name shown in UI */
  displayName: string;
  /** Brief description */
  description?: string;
}

export interface Branding {
  /** Logo (ReactNode - image, SVG, or custom component) */
  logo?: React.ReactNode;
  /** Title to display next to the logo or instead of it */
  title?: string;
  /** Copyright text */
  copyright?: string;
}

export interface ExtensionPage {
  /** Route path */
  path: string;
  /** React component for the page */
  component: React.ComponentType<any>;
  /** Page title */
  title: string;
  /** Whether authentication is required */
  requireAuth?: boolean;
}

export interface ExtensionWidget {
  /** Unique widget ID */
  id: string;
  /** React component for the widget */
  component: React.ComponentType<any>;
  /** Display name */
  displayName: string;
  /** Widget category */
  category?: string;
}

export interface ExtensionPanelPlugin {
  info: {
    label: string;
    value: string;
    icon?: React.ReactNode;
    description?: string;
    category?: string;
  };
  component: () => Promise<React.ComponentType<any>>;
  editor?: () => Promise<any>;
}

export interface Extension extends ExtensionMetadata {
  /** Pages provided by this extension */
  pages?: ExtensionPage[];
  /** Widgets that can be used in dashboards */
  widgets?: ExtensionWidget[];
  /** Navigation menu items */
  menuItems?: MenuItem[];
  /** Branding configuration */
  branding?: Branding;
  /** Panel plugins for dashboards */
  panels?: ExtensionPanelPlugin[];
  /** Data providers for backend integration */
  dataProviders?: Record<string, any>;
  /** Initialization function */
  initialize?: () => void | Promise<void>;
  /** Cleanup function */
  cleanup?: () => void | Promise<void>;
}

// ===== Extension Registry =====

class ExtensionRegistryClass {
  private extensions: Map<string, Extension> = new Map();
  private initialized: boolean = false;

  /**
   * Register an extension
   */
  register(extension: Extension): void {
    if (this.extensions.has(extension.name)) {
      console.warn(
        `Extension "${extension.name}" is already registered. Overwriting.`
      );
    }

    this.extensions.set(extension.name, extension);
    console.log(`[Extension] Registered: ${extension.name} v${extension.version}`);

    // Auto-register menu items to MenuRegistry with extension name injected
    if (extension.menuItems && extension.menuItems.length > 0) {
      const ownedItems = extension.menuItems.map(item => ({
        ...item,
        extension: extension.name,
      }));
      menuRegistry.registerAll(ownedItems);
      console.log(`[Extension] Registered ${extension.menuItems.length} menu items from ${extension.name}`);
    }
  }

  /**
   * Get a specific extension by name
   */
  getExtension(name: string): Extension | undefined {
    return this.extensions.get(name);
  }

  /**
   * Get all registered extensions
   */
  getAllExtensions(): Extension[] {
    return Array.from(this.extensions.values());
  }

  /**
   * Get all pages from all extensions
   */
  getAllPages(): ExtensionPage[] {
    const pages: ExtensionPage[] = [];

    for (const ext of this.extensions.values()) {
      if (ext.pages) {
        pages.push(...ext.pages);
      }
    }

    return pages;
  }

  /**
   * Get all widgets from all extensions
   */
  getAllWidgets(): ExtensionWidget[] {
    const widgets: ExtensionWidget[] = [];

    for (const ext of this.extensions.values()) {
      if (ext.widgets) {
        widgets.push(...ext.widgets);
      }
    }

    return widgets;
  }

  /**
   * Get all menu items from all extensions
   */
  getAllMenuItems(): MenuItem[] {
    const items: MenuItem[] = [];

    for (const ext of this.extensions.values()) {
      if (ext.menuItems) {
        items.push(...ext.menuItems);
      }
    }

    return items;
  }

  /**
   * Get all panel plugins from all extensions
   */
  getAllPanels(): ExtensionPanelPlugin[] {
    const panels: ExtensionPanelPlugin[] = [];

    for (const ext of this.extensions.values()) {
      if (ext.panels) {
        panels.push(...ext.panels);
      }
    }

    return panels;
  }

  /**
   * Get all data providers from all extensions
   */
  getAllDataProviders(): Record<string, any> {
    const dataProviders: Record<string, any> = {};

    for (const ext of this.extensions.values()) {
      if (ext.dataProviders) {
        Object.assign(dataProviders, ext.dataProviders);
      }
    }

    return dataProviders;
  }

  /**
   * Initialize all extensions
   */
  async initializeAll(): Promise<void> {
    if (this.initialized) {
      console.warn('[Extension] Already initialized');
      return;
    }

    console.log('[Extension] Initializing all extensions...');

    for (const ext of this.extensions.values()) {
      if (ext.initialize) {
        try {
          await ext.initialize();
          console.log(`[Extension] Initialized: ${ext.name}`);
        } catch (error) {
          console.error(`[Extension] Failed to initialize ${ext.name}:`, error);
        }
      }
    }

    this.initialized = true;
    console.log('[Extension] All extensions initialized');
  }

  /**
   * Cleanup all extensions
   */
  async cleanupAll(): Promise<void> {
    console.log('[Extension] Cleaning up all extensions...');

    for (const ext of this.extensions.values()) {
      if (ext.cleanup) {
        try {
          await ext.cleanup();
          console.log(`[Extension] Cleaned up: ${ext.name}`);
        } catch (error) {
          console.error(`[Extension] Failed to cleanup ${ext.name}:`, error);
        }
      }
    }

    this.initialized = false;
  }

  /**
   * Check if extensions are initialized
   */
  isInitialized(): boolean {
    return this.initialized;
  }

  /**
   * Get branding information from extensions
   * Returns the first extension that provides branding info
   */
  getBranding(): Branding | undefined {
    for (const ext of this.extensions.values()) {
      if (ext.branding) {
        return ext.branding;
      }
    }
    return undefined;
  }

  /**
   * Get extension count
   */
  getExtensionCount(): number {
    return this.extensions.size;
  }
}

// Global singleton instance
export const ExtensionRegistry = new ExtensionRegistryClass();

// ===== Helper Functions =====

/**
 * Get branding info
 */
export function getBranding(): Branding | undefined {
  return ExtensionRegistry.getBranding();
}

/**
 * Register an extension (convenience function)
 */
export function registerExtension(extension: Extension): void {
  ExtensionRegistry.register(extension);
}

/**
 * Get all registered extensions
 */
export function getExtensions(): Extension[] {
  return ExtensionRegistry.getAllExtensions();
}

/**
 * Get extension pages for routing
 */
export function getExtensionPages(): ExtensionPage[] {
  return ExtensionRegistry.getAllPages();
}

/**
 * Get extension widgets
 */
export function getExtensionWidgets(): ExtensionWidget[] {
  return ExtensionRegistry.getAllWidgets();
}

/**
 * Get all menu items from registered extensions
 * Sorted by order field
 */
export function getMenuItems(): MenuItem[] {
  return ExtensionRegistry.getAllMenuItems();
}

/**
 * Get extension panel plugins
 */
export function getExtensionPanels(): ExtensionPanelPlugin[] {
  return ExtensionRegistry.getAllPanels();
}

/**
 * Get all data providers from registered extensions
 */
export function getExtensionDataProviders(): Record<string, any> {
  return ExtensionRegistry.getAllDataProviders();
}

