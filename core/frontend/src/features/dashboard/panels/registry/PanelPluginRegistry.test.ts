import {panelPluginRegistry, PanelPlugin} from './PanelPluginRegistry';
import React from 'react';

describe('PanelPluginRegistry', () => {
  beforeEach(() => {
    panelPluginRegistry.clear();
  });

  const TestComponent: React.FC = () => React.createElement('div', null, 'Test');

  const mockPlugin: PanelPlugin<unknown> = {
    info: {
      id: 'testPanel',
      label: 'Test Panel',
    },
    component: () => Promise.resolve(TestComponent),
  };

  it('should register and retrieve plugins', () => {
    panelPluginRegistry.register(mockPlugin);
    expect(panelPluginRegistry.exists('testPanel')).toBe(true);
    expect(panelPluginRegistry.get('testPanel')).toBe(mockPlugin);
  });

  it('should unregister plugins', () => {
    panelPluginRegistry.register(mockPlugin);
    expect(panelPluginRegistry.unregister('testPanel')).toBe(true);
    expect(panelPluginRegistry.exists('testPanel')).toBe(false);
  });

  it('should get all plugins', () => {
    panelPluginRegistry.register(mockPlugin);
    expect(panelPluginRegistry.getAll()).toContain(mockPlugin);
    expect(panelPluginRegistry.getAllInfo()).toContainEqual(mockPlugin.info);
  });

  it('should load components', async () => {
    panelPluginRegistry.register(mockPlugin);
    const component = await panelPluginRegistry.loadComponent('testPanel');
    expect(component).toBe(TestComponent);
  });
});
