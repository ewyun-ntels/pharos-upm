import type { Panel } from '@pharos/shared/types/dashboard';

export type PanelCollection = {
  ids: string[];
  entities: Record<string, Panel>;
};

export function fromArray(panels: Panel[]): PanelCollection {
  const ids: string[] = [];
  const entities: Record<string, Panel> = {};

  for (const panel of panels) {
    if (!panel?.id) continue;

    if (!Object.prototype.hasOwnProperty.call(entities, panel.id)) {
      ids.push(panel.id);
    }

    entities[panel.id] = panel;
  }

  return { ids, entities };
}

export function toList(col: PanelCollection): Panel[] {
  return col.ids
    .map((id) => col.entities[id])
    .filter((panel): panel is Panel => panel !== undefined);
}

export function addFront(col: PanelCollection, panel: Panel): PanelCollection {
  const ids = [panel.id, ...col.ids.filter((id) => id !== panel.id)];
  const entities = { ...col.entities, [panel.id]: panel };

  return { ids, entities };
}

export function updateOne(col: PanelCollection, id: string, updates: Partial<Panel>): PanelCollection {
  const current = col.entities[id];
  if (!current) return col;

  const entities = {
    ...col.entities,
    [id]: { ...current, ...updates },
  };

  return { ids: col.ids, entities };
}

export function batchUpdate(
  col: PanelCollection,
  updates: Array<{ panelId: string; updates: Partial<Panel> }>,
): PanelCollection {
  if (updates.length === 0) return col;

  let hasChanges = false;
  const entities = { ...col.entities };

  for (const { panelId, updates: panelUpdates } of updates) {
    const current = entities[panelId];
    if (!current) continue;

    entities[panelId] = { ...current, ...panelUpdates };
    hasChanges = true;
  }

  if (!hasChanges) return col;

  return { ids: col.ids, entities };
}

export function removeOne(col: PanelCollection, id: string): PanelCollection {
  const ids = col.ids.filter((panelId) => panelId !== id);
  const hasEntity = Object.prototype.hasOwnProperty.call(col.entities, id);

  if (ids.length === col.ids.length && !hasEntity) {
    return col;
  }

  const entities = { ...col.entities };
  delete entities[id];

  return { ids, entities };
}
