/**
 * Panel Slice
 *
 * 패널 CRUD + 레이아웃 관리 (panelOrder, panelMap).
 */
import { v4 as uuidv4 } from 'uuid';
import type { Panel } from '@pharos/shared/types/dashboard';
import type { StoreSet, StoreGet, DashboardStore } from './types';
import {
  addFront,
  batchUpdate,
  fromArray,
  removeOne,
  toList,
  updateOne,
} from './panelCollection';

export type PanelSlice = Pick<
  DashboardStore,
  | 'panelOrder' | 'panelMap'
  | 'createPanel' | 'addPanel' | 'updatePanel' | 'updatePanelOptions'
  | 'batchUpdatePanels' | 'deletePanel' | 'duplicatePanel' | 'getPanel'
>;

export const createPanelSlice = (set: StoreSet, get: StoreGet): PanelSlice => ({
  panelOrder: [],
  panelMap: {},

  createPanel: (panelType = 'timeSeries') => {
    const { panelOrder, panelMap, id: dashboardId } = get();

    const panelId = uuidv4();

    console.log('[createPanel] Creating new panel:', {
      panelId,
      panelType,
      dashboardId,
      currentPanelCount: panelOrder.length,
    });

    const newPanel: Panel = {
      id: panelId,
      title: 'New Panel',
      renderType: panelType,
      description: '',
      bgTransparent: false,
      options: {},
      dataProvider: {
        chartQuery: [
          {
            query: '',
            datasourceName: '',
            label: '',
          },
        ],
      },
      layout: {
        type: 'card',
        h: 5,
        w: 6,
        x: 0,
        y: 0,
      },
    };

    const col = addFront({ ids: panelOrder, entities: panelMap }, newPanel);

    set({
      panelOrder: col.ids,
      panelMap: col.entities,
      hasUnsavedChanges: true,
    });

    console.log('[createPanel] Panel created successfully:', {
      panelId,
      panelExists: !!col.entities[panelId],
      totalPanels: panelOrder.length + 1,
    });

    return panelId;
  },

  addPanel: (panel) => {
    const { panelOrder, panelMap, dashboard } = get();
    if (!dashboard) return;

    const panelId = uuidv4();
    const newPanel: Panel = {
      ...panel,
      id: panelId,
      title: panel.title ?? '',
      layout: {
        type: 'card',
        h: 5,
        minW: 2,
        w: 6,
        x: 0,
        y: 0,
      },
    };

    const col = addFront({ ids: panelOrder, entities: panelMap }, newPanel);

    set({
      panelOrder: col.ids,
      panelMap: col.entities,
      hasUnsavedChanges: true,
    });
  },

  updatePanel: (panelId, updates) => {
    const { panelMap, panelOrder } = get();
    const panel = panelMap[panelId];
    if (!panel) return;

    const updateData = typeof updates === 'function' ? updates(panel) : updates;
    const col = updateOne({ ids: panelOrder, entities: panelMap }, panelId, updateData);

    set({
      panelOrder: col.ids,
      panelMap: col.entities,
      hasUnsavedChanges: true,
    });
  },

  updatePanelOptions: (panelId, optionUpdates) => {
    const { panelMap, panelOrder } = get();
    const panel = panelMap[panelId];
    if (!panel) return;

    const updatedOptions = { ...(panel.options || {}), ...optionUpdates };
    const col = updateOne({ ids: panelOrder, entities: panelMap }, panelId, {
      options: updatedOptions,
    });

    set({
      panelOrder: col.ids,
      panelMap: col.entities,
      hasUnsavedChanges: true,
    });
  },

  batchUpdatePanels: (updates) => {
    const { panelMap, panelOrder } = get();
    const col = batchUpdate({ ids: panelOrder, entities: panelMap }, updates);

    set({ panelMap: col.entities, panelOrder: col.ids, hasUnsavedChanges: true });
  },

  deletePanel: (panelId) => {
    const { panelOrder, panelMap } = get();
    const col = removeOne({ ids: panelOrder, entities: panelMap }, panelId);

    set({
      panelOrder: col.ids,
      panelMap: col.entities,
      hasUnsavedChanges: true,
    });
  },

  duplicatePanel: async (panelId) => {
    const { panelMap, panelOrder, dashboard } = get();
    const originalPanel = panelMap[panelId];

    if (!originalPanel || !dashboard) return;

    const currentDashboard = {
      ...dashboard,
      panels: toList({ ids: panelOrder, entities: panelMap }),
    };

    const { duplicatePanel: duplicatePanelUtil } = await import(
      '../../components/DashboardGrid/utils/gridUtils'
    );
    const updatedDashboard = duplicatePanelUtil(currentDashboard, originalPanel);

    const sortedPanels = updatedDashboard.panels!.sort((a, b) => {
      if (a.layout.y === b.layout.y) {
        return a.layout.x - b.layout.x;
      }
      return a.layout.y - b.layout.y;
    });

    const col = fromArray(sortedPanels);

    set({
      panelOrder: col.ids,
      panelMap: col.entities,
      dashboard: {
        ...updatedDashboard,
        panels: sortedPanels,
      },
      hasUnsavedChanges: true,
    });
  },

  getPanel: (panelId) => {
    const { panelMap } = get();
    return panelMap[panelId];
  },
});
