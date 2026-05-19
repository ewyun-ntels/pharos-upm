import {renderHook, act, waitFor} from '@testing-library/react';
import { useDashboardStore, useDashboardData } from './use-dashboard-store';
import { dashboardProvider } from '@providers/dashboard-provider';

// Mock dependencies first
jest.mock('@providers/dashboard-provider', () => ({
  dashboardProvider: {
    getOne: jest.fn(),
    create: jest.fn(),
    update: jest.fn(),
    deleteOne: jest.fn(),
  },
  DASHBOARD_RESOURCES: {
    DASHBOARD: 'dashboard',
  }
}));

describe('useDashboardStore - Integrated Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks();

    // Reset store state
    const { result } = renderHook(() => useDashboardStore());
    act(() => {
      result.current.reset();
    });
  });

  describe('Core Store Functionality', () => {
    it('should initialize with default state', () => {
      const { result } = renderHook(() => useDashboardStore());
      
      expect(result.current.id).toBeUndefined();
      expect(result.current.dashboard).toBeUndefined();
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toBeNull();
    });

    it('should load dashboard directly', () => {
      const { result } = renderHook(() => useDashboardStore());
      const sampleDashboard = {
        title: 'Test Dashboard',
        displayName: 'Test Dashboard',
        type: 'dashboard',
        panels: []
      };

      act(() => {
        // Initialize dashboard state directly using Zustand's setState
        (useDashboardStore.setState as any)({
          dashboard: sampleDashboard,
          hasUnsavedChanges: false
        });
      });

      expect(result.current.dashboard?.title).toBe('Test Dashboard');
      expect(result.current.dashboard?.displayName).toBe('Test Dashboard');
    });

    it('should update dashboard state', () => {
      const { result } = renderHook(() => useDashboardStore());

      // First set some initial state directly
      act(() => {
        // Initialize dashboard state using Zustand's setState
        (useDashboardStore.setState as any)({
          dashboard: {
            title: 'Initial Dashboard',
            displayName: 'Initial Dashboard',
            type: 'dashboard',
            panels: []
          },
          hasUnsavedChanges: false
        });
      });

      const updates = {
        title: 'Updated Dashboard',
        displayName: 'Updated Dashboard'
      };

      act(() => {
        result.current.updateDashboard(updates);
      });

      expect(result.current.dashboard?.title).toBe('Updated Dashboard');
      expect(result.current.dashboard?.displayName).toBe('Updated Dashboard');
      expect(result.current.dashboard?.type).toBe('dashboard'); // Should remain unchanged
    });

    it('should preserve panel collection when updateDashboard does not include panels', () => {
      const { result } = renderHook(() => useDashboardStore());
      const panel = { id: 'panel-1', title: 'Panel 1', layout: { type: 'card', x: 0, y: 0, w: 6, h: 4 } } as any;

      act(() => {
        (useDashboardStore.setState as any)({
          dashboard: {
            title: 'Initial Dashboard',
            displayName: 'Initial Dashboard',
            type: 'dashboard',
            panels: [panel],
          },
          panelOrder: ['panel-1'],
          panelMap: { 'panel-1': panel },
          hasUnsavedChanges: false,
        });
      });

      act(() => {
        result.current.updateDashboard({ title: 'Renamed Dashboard' });
      });

      expect((result.current as any).panelOrder).toEqual(['panel-1']);
      expect(result.current.getPanel('panel-1')?.title).toBe('Panel 1');
      expect(result.current.dashboard?.title).toBe('Renamed Dashboard');
    });

    it('should replace panel collection when updateDashboard includes panels', () => {
      const { result } = renderHook(() => useDashboardStore());
      const oldPanel = { id: 'panel-old', title: 'Old', layout: { type: 'card', x: 0, y: 0, w: 6, h: 4 } } as any;
      const newPanelA = { id: 'panel-a', title: 'A', layout: { type: 'card', x: 6, y: 0, w: 6, h: 4 } } as any;
      const newPanelB = { id: 'panel-b', title: 'B', layout: { type: 'card', x: 12, y: 0, w: 6, h: 4 } } as any;

      act(() => {
        (useDashboardStore.setState as any)({
          dashboard: {
            title: 'Initial Dashboard',
            displayName: 'Initial Dashboard',
            type: 'dashboard',
            panels: [oldPanel],
          },
          panelOrder: ['panel-old'],
          panelMap: { 'panel-old': oldPanel },
          hasUnsavedChanges: false,
        });
      });

      act(() => {
        result.current.updateDashboard({
          panels: [newPanelA, newPanelB],
          title: 'Updated Dashboard',
        } as any);
      });

      expect((result.current as any).panelOrder).toEqual(['panel-a', 'panel-b']);
      expect(result.current.getPanel('panel-old')).toBeUndefined();
      expect(result.current.getPanel('panel-a')?.title).toBe('A');
      expect(result.current.getPanel('panel-b')?.title).toBe('B');
    });

    it('should access dashboard data', () => {
      const { result } = renderHook(() => useDashboardStore());
      const sampleDashboard = {
        title: 'Test Dashboard',
        displayName: 'Test Dashboard',
        type: 'dashboard',
        panels: []
      };

      act(() => {
        // Initialize dashboard state directly
        (result.current as any).dashboard = sampleDashboard;
        (result.current as any).hasUnsavedChanges = false;
      });

      const dashboardData = result.current.dashboard;
      expect(dashboardData).toEqual(sampleDashboard);
    });

    it('should handle reset function', () => {
      const { result } = renderHook(() => useDashboardStore());
      const initialDashboard = {
        title: 'Test Dashboard',
        displayName: 'Test Dashboard',
        type: 'dashboard',
        panels: []
      };

      // First set some data directly (simulate loaded state)
      act(() => {
        (result.current as any).dashboard = initialDashboard;
        (result.current as any).hasUnsavedChanges = false;
      });

      // Then make a change
      act(() => {
        result.current.updateDashboard({ title: 'Modified Dashboard' });
        // hasUnsavedChanges는 updateDashboard에서 자동으로 true가 됨
      });

      expect(result.current.hasUnsavedChanges).toBe(true);

      // Then reset (로그아웃 시나리오)
      act(() => {
        result.current.reset();
      });

      expect(result.current.id).toBeUndefined();
      expect(result.current.dashboard).toBeUndefined();
      expect(result.current.permission).toBe('viewer');
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toBeNull();
      expect(result.current.hasUnsavedChanges).toBe(false);
    });

    it('should have all required methods', () => {
      const { result } = renderHook(() => useDashboardStore());
      
      // Core methods
      expect(typeof result.current.reset).toBe('function');
      expect(typeof result.current.updateDashboard).toBe('function');
      
      // Panel methods
      expect(typeof result.current.addPanel).toBe('function');
      expect(typeof result.current.updatePanel).toBe('function');
      expect(typeof result.current.deletePanel).toBe('function');
      expect(typeof result.current.getPanel).toBe('function');
      
      // API methods
      expect(typeof result.current._loadDashboard).toBe('function');
      expect(typeof result.current.createDashboard).toBe('function');
      expect(typeof result.current.saveDashboard).toBe('function');
      expect(typeof result.current.deleteDashboard).toBe('function');
    });
  });

  describe('Panel Management', () => {
    it('should create panel with id-based panelOrder and panelMap single source', () => {
      const { result } = renderHook(() => useDashboardStore());

      act(() => {
        (useDashboardStore.setState as any)({
          id: 'dashboard-1',
          panelOrder: [],
          panelMap: {},
          hasUnsavedChanges: false,
        });
      });

      let panelId = '';
      act(() => {
        panelId = result.current.createPanel('timeSeries');
      });

      const panelOrder = (result.current as any).panelOrder as string[];
      expect(typeof panelId).toBe('string');
      expect(panelOrder[0]).toBe(panelId);
      expect(result.current.getPanel(panelId)?.id).toBe(panelId);
      expect((result.current as any).panelMap[panelId]).toBeDefined();
    });

    it('should add a new panel', () => {
      const { result } = renderHook(() => useDashboardStore());
      const existingPanel = { id: 'existing-panel', title: 'Existing Panel' } as any;
      const dashboard = {
        title: 'Test Dashboard',
        displayName: 'Test Dashboard',
        type: 'dashboard',
        panels: [existingPanel]
      };

      act(() => {
        // Initialize dashboard state directly with panelOrder and panelMap
        (result.current as any).dashboard = dashboard;
        (result.current as any).panelOrder = ['existing-panel'];
        (result.current as any).panelMap = { 'existing-panel': existingPanel };
        (result.current as any).hasUnsavedChanges = false;
      });

      const newPanel = { 
        title: 'New Panel',
        type: 'text'
      };

      act(() => {
        result.current.addPanel(newPanel);
      });

      // panelOrder를 체크 (dashboard.panels는 save할 때만 동기화)
      const panelOrder = (result.current as any).panelOrder as string[];
      expect(panelOrder).toHaveLength(2);
      
      const newPanelId = panelOrder.find((id) => id !== 'existing-panel');
      const addedPanel = newPanelId ? result.current.getPanel(newPanelId) : undefined;
      expect(addedPanel?.title).toBe('New Panel');
      expect(addedPanel?.id).toBeDefined();
    });

    it('should update an existing panel', () => {
      const { result } = renderHook(() => useDashboardStore());
      const panel1 = { id: 'panel-1', title: 'Original Panel' } as any;
      const dashboard = {
        title: 'Test Dashboard',
        displayName: 'Test Dashboard',
        type: 'dashboard',
        panels: [panel1]
      };

      act(() => {
        // Initialize dashboard state directly (update panel test) with panelOrder and panelMap
        (result.current as any).dashboard = dashboard;
        (result.current as any).panelOrder = ['panel-1'];
        (result.current as any).panelMap = { 'panel-1': panel1 };
        (result.current as any).hasUnsavedChanges = false;
      });

      const updates = { title: 'Updated Panel' };

      act(() => {
        result.current.updatePanel('panel-1', updates);
      });

      const panel = result.current.getPanel('panel-1');
      expect(panel?.title).toBe('Updated Panel');
    });

    it('should merge panel options without changing panelOrder ids', () => {
      const { result } = renderHook(() => useDashboardStore());
      const panel = {
        id: 'panel-1',
        title: 'Panel 1',
        options: { a: 1, b: 2 },
        layout: { type: 'card', x: 0, y: 0, w: 6, h: 4 },
      } as any;

      act(() => {
        (result.current as any).dashboard = {
          title: 'Test Dashboard',
          displayName: 'Test Dashboard',
          type: 'dashboard',
          panels: [panel],
        };
        (result.current as any).panelOrder = ['panel-1'];
        (result.current as any).panelMap = { 'panel-1': panel };
        (result.current as any).hasUnsavedChanges = false;
      });

      act(() => {
        result.current.updatePanelOptions('panel-1', { b: 3, c: 4 } as any);
      });

      expect((result.current as any).panelOrder).toEqual(['panel-1']);
      expect(result.current.getPanel('panel-1')?.options).toEqual({ a: 1, b: 3, c: 4 });
    });

    it('should delete a panel', () => {
      const { result } = renderHook(() => useDashboardStore());
      const panel1 = { id: 'panel-1', title: 'Panel 1' } as any;
      const panel2 = { id: 'panel-2', title: 'Panel 2' } as any;
      const dashboard = {
        title: 'Test Dashboard',
        displayName: 'Test Dashboard',
        type: 'dashboard',
        panels: [panel1, panel2]
      };

      act(() => {
        // Initialize dashboard state directly (delete panel test) with panelOrder and panelMap
        (result.current as any).dashboard = dashboard;
        (result.current as any).panelOrder = ['panel-1', 'panel-2'];
        (result.current as any).panelMap = { 'panel-1': panel1, 'panel-2': panel2 };
        (result.current as any).hasUnsavedChanges = false;
      });

      act(() => {
        result.current.deletePanel('panel-1');
      });

      // panelOrder를 체크 (dashboard.panels는 save할 때만 동기화)
      const panelOrder = (result.current as any).panelOrder;
      expect(panelOrder).toHaveLength(1);
      expect(panelOrder[0]).toBe('panel-2');
      expect(result.current.getPanel('panel-1')).toBeUndefined();
    });

    it('should keep panelOrder ids while batch updating panelMap', () => {
      const { result } = renderHook(() => useDashboardStore());
      const panel1 = { id: 'panel-1', title: 'Panel 1', layout: { x: 0, y: 0, w: 6, h: 4 } } as any;
      const panel2 = { id: 'panel-2', title: 'Panel 2', layout: { x: 6, y: 0, w: 6, h: 4 } } as any;

      act(() => {
        (result.current as any).dashboard = {
          title: 'Test Dashboard',
          displayName: 'Test Dashboard',
          type: 'dashboard',
          panels: [panel1, panel2],
        };
        (result.current as any).panelOrder = ['panel-1', 'panel-2'];
        (result.current as any).panelMap = { 'panel-1': panel1, 'panel-2': panel2 };
        (result.current as any).hasUnsavedChanges = false;
      });

      act(() => {
        result.current.batchUpdatePanels([
          { panelId: 'panel-1', updates: { title: 'Panel 1 Updated' } },
          { panelId: 'panel-2', updates: { title: 'Panel 2 Updated' } },
        ]);
      });

      expect((result.current as any).panelOrder).toEqual(['panel-1', 'panel-2']);
      expect(result.current.getPanel('panel-1')?.title).toBe('Panel 1 Updated');
      expect(result.current.getPanel('panel-2')?.title).toBe('Panel 2 Updated');
    });

    it('should duplicate a panel and keep ids/entities synchronized', async () => {
      const { result } = renderHook(() => useDashboardStore());
      const panel1 = {
        id: 'panel-1',
        title: 'Original Panel',
        renderType: 'timeSeries',
        layout: { type: 'card', x: 0, y: 0, w: 6, h: 4 },
      } as any;

      act(() => {
        (result.current as any).dashboard = {
          title: 'Test Dashboard',
          displayName: 'Test Dashboard',
          type: 'dashboard',
          panels: [panel1],
        };
        (result.current as any).panelOrder = ['panel-1'];
        (result.current as any).panelMap = { 'panel-1': panel1 };
        (result.current as any).hasUnsavedChanges = false;
      });

      await act(async () => {
        await result.current.duplicatePanel('panel-1');
      });

      const panelOrder = (result.current as any).panelOrder as string[];
      const panelMap = (result.current as any).panelMap as Record<string, any>;

      expect(panelOrder).toHaveLength(2);
      expect(Object.keys(panelMap)).toHaveLength(2);

      const duplicatedId = panelOrder.find((id) => id !== 'panel-1');
      expect(duplicatedId).toBeDefined();

      const duplicatedPanel = duplicatedId ? panelMap[duplicatedId] : undefined;
      expect(duplicatedPanel?.id).not.toBe('panel-1');
      expect(duplicatedPanel?.title).toBe('Original Panel (Copy)');
      expect(panelMap['panel-1']?.title).toBe('Original Panel');
    });
  });

  describe('API Operations', () => {
    it('should fetch dashboard successfully', async () => {
      const mockDataProvider = dashboardProvider as jest.Mocked<typeof dashboardProvider>;
      mockDataProvider.getOne.mockResolvedValue({
        data: {
          config: {
            title: 'Fetched Dashboard',
            displayName: 'Fetched Dashboard',
            type: 'dashboard'
          },
          permission: 'owner'
        },
      });

      const { result } = renderHook(() => useDashboardStore());

      await act(() => result.current._loadDashboard('test-id'));

      expect(mockDataProvider.getOne).toHaveBeenCalledWith({
        resource: 'dashboard',
        id: 'test-id',
      });
      
      await waitFor(() => {
        expect(result.current.id).toBe('test-id');
        expect(result.current.dashboard?.title).toBe('Fetched Dashboard');
        expect(result.current.permission).toBe('owner');
      });
    });

    it('should convert loaded panels into panelOrder ids and panelMap entities', async () => {
      const mockDataProvider = dashboardProvider as jest.Mocked<typeof dashboardProvider>;
      const panelA = { id: 'panel-a', title: 'Panel A', layout: { type: 'card', x: 6, y: 0, w: 6, h: 4 } } as any;
      const panelB = { id: 'panel-b', title: 'Panel B', layout: { type: 'card', x: 0, y: 0, w: 6, h: 4 } } as any;

      mockDataProvider.getOne.mockResolvedValue({
        data: {
          config: {
            title: 'Fetched Dashboard',
            displayName: 'Fetched Dashboard',
            type: 'dashboard',
            panels: [panelA, panelB],
            filters: [],
          },
          permission: 'owner',
        },
      } as any);

      const { result } = renderHook(() => useDashboardStore());

      await act(() => result.current._loadDashboard('dashboard-with-panels'));

      await waitFor(() => {
        expect((result.current as any).panelOrder).toEqual(['panel-a', 'panel-b']);
        expect(result.current.getPanel('panel-a')?.title).toBe('Panel A');
        expect(result.current.getPanel('panel-b')?.title).toBe('Panel B');
      });
    });

    it('should handle fetch errors', async () => {
      const mockDataProvider = dashboardProvider as jest.Mocked<typeof dashboardProvider>;
      const mockError = new Error('Fetch failed');
      mockDataProvider.getOne.mockRejectedValue(mockError);

      const { result } = renderHook(() => useDashboardStore());

      await act(() => result.current._loadDashboard('test-id'));

      await waitFor(() => {
        expect(result.current.error).toBe(mockError);
        expect(result.current.loading).toBe(false);
        expect(result.current.id).toBeUndefined();
        expect(result.current.dashboard).toBeUndefined();
      });
    });

    it('should create dashboard successfully', async () => {
      const mockDataProvider = dashboardProvider as jest.Mocked<typeof dashboardProvider>;
      const newDashboard = { title: 'New Dashboard', type: 'dashboard' };
      const createdDashboard = { id: 'new-id', ...newDashboard };
      
      mockDataProvider.create.mockResolvedValue({
        data: createdDashboard,
      });

      const { result } = renderHook(() => useDashboardStore());

      let createdResult;
      await act(async () => {
        createdResult = await result.current.createDashboard(newDashboard);
      });

      expect(mockDataProvider.create).toHaveBeenCalledWith({
        resource: 'dashboard',
        variables: newDashboard,
      });
      expect(createdResult).toEqual(createdDashboard);
    });

    it('should duplicate dashboard with copy suffix', async () => {
      const mockDataProvider = dashboardProvider as jest.Mocked<typeof dashboardProvider>;
      mockDataProvider.getOne.mockResolvedValue({
        data: {
          config: {
            title: 'Original Dashboard',
            displayName: 'Original Dashboard',
            type: 'dashboard',
            panels: [],
            filters: [],
          },
        },
      } as any);
      mockDataProvider.create.mockResolvedValue({
        data: { id: 'duplicated-dashboard-id' },
      } as any);

      const { result } = renderHook(() => useDashboardStore());

      let duplicatedResult;
      await act(async () => {
        duplicatedResult = await result.current.duplicateDashboard('original-dashboard-id');
      });

      expect(mockDataProvider.getOne).toHaveBeenCalledWith({
        resource: 'dashboard',
        id: 'original-dashboard-id',
      });

      expect(mockDataProvider.create).toHaveBeenCalledWith({
        resource: 'dashboard',
        variables: expect.objectContaining({
          title: 'Original Dashboard (Copy)',
          displayName: 'Original Dashboard (Copy)',
        }),
      });

      expect(duplicatedResult).toEqual({ id: 'duplicated-dashboard-id' });
    });

    it('should set error and rethrow when duplicateDashboard fails', async () => {
      const mockDataProvider = dashboardProvider as jest.Mocked<typeof dashboardProvider>;
      const duplicateError = new Error('Duplicate failed');
      mockDataProvider.getOne.mockRejectedValue(duplicateError);

      const { result } = renderHook(() => useDashboardStore());

      await act(async () => {
        await expect(result.current.duplicateDashboard('dashboard-id')).rejects.toThrow('Duplicate failed');
      });

      expect(result.current.error).toBe(duplicateError);
      expect(result.current.loading).toBe(false);
    });

    it('should save dashboard changes', async () => {
      const mockDataProvider = dashboardProvider as jest.Mocked<typeof dashboardProvider>;
      const dashboard = {
        title: 'Test Dashboard',
        displayName: 'Test Dashboard',
        type: 'dashboard',
        panels: []
      };
      
      mockDataProvider.update.mockResolvedValue({
        data: { id: 'test-id', ...dashboard }
      });

      const { result } = renderHook(() => useDashboardStore());
      
      act(() => {
        // Initialize dashboard state using Zustand's setState
        (useDashboardStore.setState as any)({
          dashboard: dashboard,
          id: 'test-id',
          panelOrder: [],
          panelMap: {},
          filters: [],
          hasUnsavedChanges: false
        });
      });

      await act(async () => {
        await result.current.saveDashboard();
      });

      expect(mockDataProvider.update).toHaveBeenCalledWith({
        resource: 'dashboard',
        id: 'test-id',
        variables: {
          ...dashboard,
          filters: []
        },
      });
    });

    it('should serialize panels using panelOrder and panelMap order when saving', async () => {
      const mockDataProvider = dashboardProvider as jest.Mocked<typeof dashboardProvider>;
      const panel1 = { id: 'panel-1', title: 'Panel 1', layout: { type: 'card', x: 0, y: 0, w: 6, h: 4 } } as any;
      const panel2 = { id: 'panel-2', title: 'Panel 2', layout: { type: 'card', x: 6, y: 0, w: 6, h: 4 } } as any;

      mockDataProvider.update.mockResolvedValue({ data: { id: 'save-order-test' } } as any);

      const { result } = renderHook(() => useDashboardStore());

      act(() => {
        (useDashboardStore.setState as any)({
          dashboard: {
            title: 'Save Order Dashboard',
            displayName: 'Save Order Dashboard',
            type: 'dashboard',
            panels: [],
          },
          id: 'save-order-test',
          panelOrder: ['panel-2', 'panel-1'],
          panelMap: {
            'panel-1': panel1,
            'panel-2': panel2,
          },
          filters: [],
          hasUnsavedChanges: true,
        });
      });

      await act(async () => {
        await result.current.saveDashboard();
      });

      const updateCall = mockDataProvider.update.mock.calls.at(-1)?.[0] as any;
      const savedPanels = updateCall?.variables?.panels ?? [];

      expect(savedPanels.map((panel: any) => panel.id)).toEqual(['panel-2', 'panel-1']);
      expect(savedPanels[0].title).toBe('Panel 2');
      expect(savedPanels[1].title).toBe('Panel 1');
    });
  });

  describe('Derived Data', () => {
    it('should derive ordered panels from panelOrder ids and panelMap entities', () => {
      const panel1 = { id: 'panel-1', title: 'Panel 1', layout: { type: 'card', x: 0, y: 0, w: 6, h: 4 } } as any;
      const panel2 = { id: 'panel-2', title: 'Panel 2', layout: { type: 'card', x: 6, y: 0, w: 6, h: 4 } } as any;

      const { result } = renderHook(() => useDashboardData());

      act(() => {
        (useDashboardStore.setState as any)({
          panelOrder: ['panel-2', 'missing-panel', 'panel-1'],
          panelMap: {
            'panel-1': panel1,
            'panel-2': panel2,
          },
        });
      });

      expect(result.current.panels.map((panel) => panel.id)).toEqual(['panel-2', 'panel-1']);
    });
  });

  describe('Error Handling and Edge Cases', () => {
    it('should handle operations when no dashboard is loaded', () => {
      const { result } = renderHook(() => useDashboardStore());

      // Try to update when no dashboard exists
      act(() => {
        result.current.updateDashboard({ title: 'Should not work' });
      });

      expect(result.current.dashboard).toBeUndefined();

      // Try to add panel when no dashboard exists
      act(() => {
        result.current.addPanel({ title: 'Should not work' });
      });

      expect(result.current.dashboard).toBeUndefined();
    });

    it('should handle malformed dashboard data gracefully', async () => {
      const mockDataProvider = dashboardProvider as jest.Mocked<typeof dashboardProvider>;
      mockDataProvider.getOne.mockResolvedValue({
        data: {} as any, // Malformed data
      });

      const { result } = renderHook(() => useDashboardStore());

      await act(() => result.current._loadDashboard('test-id'));

      await waitFor(() => {
        expect(result.current.error).toBeTruthy();
        expect(result.current.dashboard).toBeUndefined();
      });
    });

    it('should handle network errors gracefully', async () => {
      const mockDataProvider = dashboardProvider as jest.Mocked<typeof dashboardProvider>;
      const networkError = new Error('Network error');
      mockDataProvider.getOne.mockRejectedValue(networkError);

      const { result } = renderHook(() => useDashboardStore());

      await act(() => result.current._loadDashboard('test-id'));

      await waitFor(() => {
        expect(result.current.error).toBe(networkError);
        expect(result.current.loading).toBe(false);
      });
    });
  });

  describe('Legacy Compatibility', () => {
    it('should maintain backward compatibility with existing API', () => {
      const { result } = renderHook(() => useDashboardStore());
      const existingPanel = { id: 'existing-panel', title: 'Existing Panel' } as any;
      const dashboard = {
        title: 'Legacy Dashboard',
        displayName: 'Legacy Dashboard',
        type: 'dashboard',
        panels: [existingPanel]
      };

      act(() => {
        // Initialize dashboard state directly with panelOrder and panelMap
        (result.current as any).dashboard = dashboard;
        (result.current as any).id = 'legacy-id';
        (result.current as any).user_role = 'editor';
        (result.current as any).panelOrder = ['existing-panel'];
        (result.current as any).panelMap = { 'existing-panel': existingPanel };
        (result.current as any).hasUnsavedChanges = false;
      });

      // Add panel (legacy way)
      act(() => {
        result.current.addPanel({ title: 'Legacy Panel' });
      });

      // panelOrder를 체크 (dashboard.panels는 save할 때만 동기화)
      let panelOrder = (result.current as any).panelOrder as string[];
      expect(panelOrder.length).toBe(2); // 1 existing + 1 new

      const newPanelId = panelOrder.find((id) => id !== 'existing-panel');
      const addedPanel = newPanelId ? result.current.getPanel(newPanelId) : undefined;
      expect(addedPanel).toBeDefined();
      expect(typeof addedPanel?.id).toBe('string');

      // Update panel
      const panelId = addedPanel?.id;
      act(() => {
        result.current.updatePanel(panelId!, { title: 'Updated Legacy Panel' });
      });

      const updatedPanel = result.current.getPanel(panelId!);
      expect(updatedPanel?.title).toBe('Updated Legacy Panel');

      // Delete panel
      act(() => {
        result.current.deletePanel(panelId!);
      });

      panelOrder = (result.current as any).panelOrder;
      expect(panelOrder.length).toBe(1);
      expect(panelOrder.find((id: string) => id === panelId)).toBeUndefined();
    });
  });

  describe('Annotations', () => {
    it('should have empty storeAnnotations on initial state', () => {
      const { result } = renderHook(() => useDashboardStore());
      expect(result.current.storeAnnotations).toEqual([]);
    });

    it('should have empty storeAnnotations after reset', () => {
      const { result } = renderHook(() => useDashboardStore());

      act(() => {
        (useDashboardStore.setState as any)({ storeAnnotations: [{ id: 'a1', time: 1000, text: 'test', scope: 'global' }] });
      });
      expect(result.current.storeAnnotations).toHaveLength(1);

      act(() => { result.current.reset(); });
      expect(result.current.storeAnnotations).toEqual([]);
    });

    it('should set annotations via setAnnotations', () => {
      const { result } = renderHook(() => useDashboardStore());
      const annotations = [
        { id: 'a1', time: 1700000000000, text: 'global event', scope: 'global' as const },
        { id: 'a2', time: 1700001000000, text: 'panel note', scope: 'panel' as const, panel_id: 'panel-1' },
      ];

      act(() => { result.current.setAnnotations(annotations); });

      expect(result.current.storeAnnotations).toHaveLength(2);
      expect(result.current.storeAnnotations[0].id).toBe('a1');
      expect(result.current.storeAnnotations[1].panel_id).toBe('panel-1');
    });

    it('should replace all annotations on setAnnotations (overwrite)', () => {
      const { result } = renderHook(() => useDashboardStore());

      act(() => {
        result.current.setAnnotations([
          { id: 'old-1', time: 1000, text: 'old', scope: 'global' as const },
          { id: 'old-2', time: 2000, text: 'old2', scope: 'global' as const },
        ]);
      });
      act(() => {
        result.current.setAnnotations([{ id: 'new-1', time: 3000, text: 'new', scope: 'global' as const }]);
      });

      expect(result.current.storeAnnotations).toHaveLength(1);
      expect(result.current.storeAnnotations[0].id).toBe('new-1');
    });

    it('should clear annotations by setting empty array', () => {
      const { result } = renderHook(() => useDashboardStore());

      act(() => {
        result.current.setAnnotations([{ id: 'a1', time: 1000, text: 'x', scope: 'global' as const }]);
      });
      act(() => {
        result.current.setAnnotations([]);
      });

      expect(result.current.storeAnnotations).toEqual([]);
    });

    it('should load annotations from API response on _loadDashboard', async () => {
      const mockDataProvider = dashboardProvider as jest.Mocked<typeof dashboardProvider>;
      mockDataProvider.getOne.mockResolvedValue({
        data: {
          config: { title: 'Dashboard', displayName: 'Dashboard', type: 'dashboard', panels: [], filters: [] },
          permission: 'owner',
          annotations: [
            { id: 'ann-1', time: 1700000000000, text: 'from server', scope: 'global' },
            { id: 'ann-2', time: 1700001000000, text: 'panel ann', scope: 'panel', panel_id: 'p1' },
          ],
        },
      } as any);

      const { result } = renderHook(() => useDashboardStore());
      await act(() => result.current._loadDashboard('dash-with-annotations'));

      await waitFor(() => {
        expect(result.current.storeAnnotations).toHaveLength(2);
        expect(result.current.storeAnnotations[0].id).toBe('ann-1');
        expect(result.current.storeAnnotations[1].panel_id).toBe('p1');
      });
    });

    it('should default to empty array when API response has no annotations', async () => {
      const mockDataProvider = dashboardProvider as jest.Mocked<typeof dashboardProvider>;
      mockDataProvider.getOne.mockResolvedValue({
        data: {
          config: { title: 'Dashboard', displayName: 'Dashboard', type: 'dashboard', panels: [], filters: [] },
          permission: 'owner',
          // annotations 필드 없음
        },
      } as any);

      const { result } = renderHook(() => useDashboardStore());
      await act(() => result.current._loadDashboard('dash-no-annotations'));

      await waitFor(() => {
        expect(result.current.storeAnnotations).toEqual([]);
      });
    });

    it('should expose setAnnotations as a function', () => {
      const { result } = renderHook(() => useDashboardStore());
      expect(typeof result.current.setAnnotations).toBe('function');
    });
  });
});
