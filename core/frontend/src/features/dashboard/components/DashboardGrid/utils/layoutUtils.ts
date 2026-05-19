import {LayoutItem} from 'react-grid-layout';
import {Panel} from '@pharos/shared/types/dashboard';

export function convertJsonToGridLayout(panels: Panel[]): LayoutItem[] {
  return panels.map((panel) => {
    return {
      i: panel.id.toString(), // Panel.id를 react-grid-layout의 unique key로 사용
      x: panel.layout.x,
      y: panel.layout.y,
      w: panel.layout.w,
      h: panel.layout.h,
      minW: panel.layout.type === 'row' ? 24 : 6,
      static: panel.layout.static ?? false,
      isDraggable: panel.layout.isDraggable ?? true,
      isResizable: panel.layout.isResizable ?? true,
    };
  });
}

export const compactLayout = (layout: LayoutItem[], panels: Panel[]): LayoutItem[] => {
  const rowPanels = panels.filter((panel) => panel.layout.type === 'row');
  const regularPanels = panels.filter((panel) => panel.layout.type !== 'row');

  // Row 패널의 상태 (펼쳐짐/접힘) 정보
  const rowCollapsedState: {[id: string]: boolean} = {};
  rowPanels.forEach((panel) => {
    rowCollapsedState[panel.id.toString()] = !panel.layout.collapsed; // collapsed=false면 펼쳐진 상태
  });

  // 1. Row 패널들을 Y 좌표 기준 정렬
  const rowLayouts = layout
    .filter((item) => rowPanels.some((panel) => panel.id.toString() === item.i))
    .sort((a, b) => a.y - b.y);

  // 2. 일반 패널들 - 드래그로 이동한 패널은 그 위치를 유지
  const regularLayouts = layout.filter((item) =>
    regularPanels.some((panel) => panel.id.toString() === item.i),
  );

  // 3. Row 범위 계산 함수
  const getRowRange = (rowY: number): {start: number; end: number} => {
    const nextRow = rowLayouts.find((r) => r.y > rowY);
    return {
      start: rowY + 1, // Row 바로 다음 행부터
      end: nextRow ? nextRow.y : Number.MAX_SAFE_INTEGER,
    };
  };

  // 4. 패널이 속한 Row 찾기
  const findContainingRow = (panelY: number): LayoutItem | null => {
    for (const rowLayout of rowLayouts) {
      const range = getRowRange(rowLayout.y);
      if (panelY >= range.start && panelY < range.end) {
        return rowLayout;
      }
    }
    return null;
  };

  const compactedLayout: LayoutItem[] = [];

  // 5. Row 패널들 먼저 배치 (단순 추가)
  rowLayouts.forEach((rowItem) => {
    compactedLayout.push(rowItem);
  });

  // 6. 일반 패널들을 최소한으로 처리 (react-grid-layout compaction에 의존)
  regularLayouts.forEach((item) => {
    const containingRow = findContainingRow(item.y);

    // Row에 속한 경우 해당 Row가 접혀있으면 숨김
    if (containingRow && !rowCollapsedState[containingRow.i]) {
      return;
    }

    let finalItem = {...item};

    if (containingRow) {
      const range = getRowRange(containingRow.y);

      // Row 범위를 벗어났을 때만 조정
      if (item.y < range.start) {
        finalItem.y = range.start;
      } else if (item.y >= range.end) {
        finalItem.y = range.end - 1;
      }
      // Row 범위 내에서는 원래 위치 유지 (react-grid-layout이 compaction 처리)
    }

    compactedLayout.push(finalItem);
  });

  return compactedLayout.sort((a, b) => {
    if (a.y === b.y) return a.x - b.x;
    return a.y - b.y;
  });
};

// 드래그 종료 후 사용할 선택적 컴팩트 함수 - 빈 공간 유지
export const selectiveCompactLayout = (layout: LayoutItem[], panels: Panel[]): LayoutItem[] => {
  const rowPanels = panels.filter((panel) => panel.layout.type === 'row');

  const rowCollapsedState: {[id: string]: boolean} = {};
  rowPanels.forEach((panel) => {
    rowCollapsedState[panel.id.toString()] = !panel.layout.collapsed;
  });

  const visibleLayout = layout.filter((item) => {
    const panel = panels.find((p) => p.id.toString() === item.i);
    if (!panel || panel.layout.type === 'row') return true;

    const rowLayouts = layout
      .filter((layoutItem) => rowPanels.some((panel) => panel.id.toString() === layoutItem.i))
      .sort((a, b) => a.y - b.y);

    const findContainingRow = (panelY: number): LayoutItem | null => {
      for (const rowLayout of rowLayouts) {
        const nextRow = rowLayouts.find((r) => r.y > rowLayout.y);
        const range = {
          start: rowLayout.y + 1,
          end: nextRow ? nextRow.y : Number.MAX_SAFE_INTEGER,
        };
        if (panelY >= range.start && panelY < range.end) {
          return rowLayout;
        }
      }
      return null;
    };

    const containingRow = findContainingRow(item.y);
    return !containingRow || rowCollapsedState[containingRow.i];
  });

  return visibleLayout.sort((a, b) => {
    if (a.y === b.y) return a.x - b.x;
    return a.y - b.y;
  });
};
