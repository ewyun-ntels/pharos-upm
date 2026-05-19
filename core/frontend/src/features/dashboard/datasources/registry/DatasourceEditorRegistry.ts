import React from 'react';

/**
 * Datasource Editor Component Props
 */
export interface DatasourceEditorComponentProps {
  /** Panel ID for unique editor instance */
  panelId: string;
  /** Datasource name selected by user */
  datasourceName: string;
  /** Current query string */
  query: string;
  /** Query change handler */
  onQueryChange: (query: string) => void;
  /** Run query handler */
  onRunQuery: () => void;
  /** Loading state */
  isLoading?: boolean;
}

/**
 * Datasource Editor Plugin Interface
 */
export interface DatasourceEditorPlugin {
  /** Datasource type this editor handles */
  datasourceType: string;
  /** Display name for the editor */
  name: string;
  /** Editor component */
  component: React.ComponentType<DatasourceEditorComponentProps>;
  /** Editor description */
  description?: string;
  /** Priority for plugin selection (higher = preferred) */
  priority?: number;
}

/**
 * Registry for managing datasource editor plugins
 */
class DatasourceEditorRegistry {
  private editors = new Map<string, DatasourceEditorPlugin>();

  /**
   * Register a datasource editor plugin
   */
  register(plugin: DatasourceEditorPlugin): void {
    this.editors.set(plugin.datasourceType, plugin);
    console.log(`[DatasourceEditorRegistry] Registered editor for type: ${plugin.datasourceType}`);
  }

  /**
   * Unregister a datasource editor plugin
   */
  unregister(datasourceType: string): boolean {
    const removed = this.editors.delete(datasourceType);
    if (removed) {
      console.log(`[DatasourceEditorRegistry] Unregistered editor for type: ${datasourceType}`);
    }
    return removed;
  }

  /**
   * Get editor plugin for datasource type
   */
  get(datasourceType: string): DatasourceEditorPlugin | undefined {
    return this.editors.get(datasourceType);
  }

  /**
   * Get all registered editor plugins
   */
  getAll(): DatasourceEditorPlugin[] {
    return Array.from(this.editors.values());
  }

  /**
   * Check if editor exists for datasource type
   */
  exists(datasourceType: string): boolean {
    return this.editors.has(datasourceType);
  }

  /**
   * Get available datasource types
   */
  getSupportedTypes(): string[] {
    return Array.from(this.editors.keys());
  }

  /**
   * Clear all registered editors
   */
  clear(): void {
    this.editors.clear();
  }

  /**
   * Get editor count
   */
  getEditorCount(): number {
    return this.editors.size;
  }
}

export const datasourceEditorRegistry = new DatasourceEditorRegistry();
