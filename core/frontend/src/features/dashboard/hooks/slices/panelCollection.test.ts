import {
  addFront,
  batchUpdate,
  fromArray,
  removeOne,
  toList,
  updateOne,
} from './panelCollection';

function createPanel(id: string, title = id) {
  return {
    id,
    title,
    layout: { type: 'card', x: 0, y: 0, w: 6, h: 4 },
  } as any;
}

describe('panelCollection helpers', () => {
  it('should convert Panel[] to collection and back while preserving order', () => {
    const panel1 = createPanel('panel-1');
    const panel2 = createPanel('panel-2');

    const collection = fromArray([panel1, panel2]);

    expect(collection.ids).toEqual(['panel-1', 'panel-2']);
    expect(collection.entities['panel-1'].title).toBe('panel-1');
    expect(collection.entities['panel-2'].title).toBe('panel-2');

    const list = toList(collection);
    expect(list.map((panel) => panel.id)).toEqual(['panel-1', 'panel-2']);
  });

  it('should add panel to front and deduplicate existing id', () => {
    const panel1 = createPanel('panel-1', 'One');
    const panel2 = createPanel('panel-2', 'Two');
    const collection = fromArray([panel1, panel2]);

    const movedPanel2 = createPanel('panel-2', 'Two Updated');
    const next = addFront(collection, movedPanel2);

    expect(next.ids).toEqual(['panel-2', 'panel-1']);
    expect(next.entities['panel-2'].title).toBe('Two Updated');
  });

  it('should update one panel and return original when id does not exist', () => {
    const collection = fromArray([createPanel('panel-1', 'One')]);

    const updated = updateOne(collection, 'panel-1', { title: 'One Updated' });
    expect(updated.entities['panel-1'].title).toBe('One Updated');

    const unchanged = updateOne(collection, 'missing', { title: 'Missing' });
    expect(unchanged).toBe(collection);
  });

  it('should batch update existing panels only and keep ids order', () => {
    const collection = fromArray([createPanel('panel-1', 'One'), createPanel('panel-2', 'Two')]);

    const next = batchUpdate(collection, [
      { panelId: 'panel-2', updates: { title: 'Two Updated' } },
      { panelId: 'missing', updates: { title: 'Ignored' } },
    ]);

    expect(next.ids).toEqual(['panel-1', 'panel-2']);
    expect(next.entities['panel-1'].title).toBe('One');
    expect(next.entities['panel-2'].title).toBe('Two Updated');

    const emptyUpdate = batchUpdate(collection, []);
    expect(emptyUpdate).toBe(collection);
  });

  it('should remove panel id/entity and return original when id is missing', () => {
    const collection = fromArray([createPanel('panel-1'), createPanel('panel-2')]);

    const removed = removeOne(collection, 'panel-1');
    expect(removed.ids).toEqual(['panel-2']);
    expect(removed.entities['panel-1']).toBeUndefined();

    const unchanged = removeOne(collection, 'missing');
    expect(unchanged).toBe(collection);
  });
});
