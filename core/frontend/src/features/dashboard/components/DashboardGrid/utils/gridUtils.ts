import {Panel} from '@pharos/shared/types/dashboard';
import {DashboardConfig} from '@pharos/shared/types/dashboard';
import {v4 as uuidv4} from 'uuid';

function moveExistingPanelsDown(dashboard: DashboardConfig, height: number): DashboardConfig {
  if (!dashboard.panels || dashboard.panels.length === 0) {
    return dashboard;
  }

  const dashboardCopy = {...dashboard};
  if (dashboardCopy.panels) {
    // 모든 패널의 Y 좌표를 증가시킴
    dashboardCopy.panels = dashboardCopy.panels.map((panel) => ({
      ...panel,
      layout: {
        ...panel.layout,
        y: panel.layout.y + height,
      },
    }));
  }

  return dashboardCopy;
}

export function addPanel(dashboard: DashboardConfig, panelType: 'row' | 'card'): DashboardConfig {
  const newPanelId = uuidv4();

  const newPanel: Panel = {
    id: newPanelId,
    title: panelType === 'row' ? 'New Row' : 'New Panel',
    layout: {
      type: panelType,
      x: 0,
      y: 0,
      w: panelType === 'row' ? 24 : 6,
      h: panelType === 'row' ? 1 : 4,
      collapsed: panelType === 'row' ? true : undefined,
    },
  };

  const dashboardCopy = moveExistingPanelsDown(dashboard, panelType === 'row' ? 1 : 4);

  if (dashboardCopy.panels) {
    dashboardCopy.panels = [newPanel, ...dashboardCopy.panels];
  } else {
    dashboardCopy.panels = [newPanel];
  }

  return dashboardCopy;
}

/**
 * 패널 복제 함수 (Grafana-style)
 * @param dashboard 대시보드 설정
 * @param originalPanel 복제할 원본 패널
 * @returns 새로운 대시보드 설정
 */
export function duplicatePanel(dashboard: DashboardConfig, originalPanel: Panel): DashboardConfig {
  if (!dashboard.panels) return dashboard;

  const GRID_COLUMN_COUNT = 24;

  const newPanel: Panel = {
    ...originalPanel,
    id: uuidv4(),
    title: `${originalPanel.title} (Copy)`,
    layout: {
      ...originalPanel.layout,
    },
  };

  // 충돌 체크 함수
  const isPositionOccupied = (x: number, y: number, w: number, h: number): boolean => {
    return dashboard.panels!.some((panel) => {
      const layout = panel.layout;
      return !(
        x >= layout.x + layout.w ||
        x + w <= layout.x ||
        y >= layout.y + layout.h ||
        y + h <= layout.y
      );
    });
  };

  const findEmptySpaceInRow = (
    targetY: number,
    needWidth: number,
    needHeight: number,
    preferredX?: number,
  ): number | null => {
    const panelsInRow = dashboard
      .panels!.filter(
        (panel) => panel.layout.y <= targetY && panel.layout.y + panel.layout.h > targetY,
      )
      .sort((a, b) => a.layout.x - b.layout.x);

    const emptySpaces: number[] = [];
    let currentX = 0;

    for (const panel of panelsInRow) {
      if (currentX + needWidth <= panel.layout.x) {
        emptySpaces.push(currentX);
      }
      currentX = panel.layout.x + panel.layout.w;
    }

    if (currentX + needWidth <= GRID_COLUMN_COUNT) {
      emptySpaces.push(currentX);
    }

    if (emptySpaces.length === 0) {
      return null;
    }

    if (preferredX !== undefined) {
      return emptySpaces.reduce((prev, curr) =>
        Math.abs(curr - preferredX) < Math.abs(prev - preferredX) ? curr : prev,
      );
    }

    return emptySpaces[0];
  };

  const findOptimalPosition = (
    originalLayout: any,
    needWidth: number,
    needHeight: number,
  ): {x: number; y: number; needsPush: boolean} => {
    // 1순위: 원본 패널 오른쪽 (같은 행)
    const rightX = originalLayout.x + originalLayout.w;
    if (rightX + needWidth <= GRID_COLUMN_COUNT) {
      if (!isPositionOccupied(rightX, originalLayout.y, needWidth, needHeight)) {
        return {x: rightX, y: originalLayout.y, needsPush: false};
      }
    }

    // 2순위: 원본 패널 왼쪽 (같은 행)
    const leftX = originalLayout.x - needWidth;
    if (leftX >= 0) {
      if (!isPositionOccupied(leftX, originalLayout.y, needWidth, needHeight)) {
        return {x: leftX, y: originalLayout.y, needsPush: false};
      }
    }

    // 3순위: 원본 패널과 같은 행에서 빈 공간 찾기 (원본 패널 위치 기준)
    const emptyXInSameRow = findEmptySpaceInRow(
      originalLayout.y,
      needWidth,
      needHeight,
      originalLayout.x,
    );
    if (emptyXInSameRow !== null) {
      return {x: emptyXInSameRow, y: originalLayout.y, needsPush: false};
    }

    // 4순위: 원본 패널 이후 행들에서 빈 공간 찾기
    const candidatePositions = findBestPositionInLaterRows(originalLayout, needWidth, needHeight);
    if (candidatePositions.length > 0) {
      return candidatePositions[0];
    }

    // 5순위: 새로운 행 생성 (원본 패널 다음 첫 x=0 패널 위치)
    const maxY = Math.max(...dashboard.panels!.map((p) => p.layout.y + p.layout.h));
    return {x: 0, y: maxY, needsPush: false};
  };

  // 원본 패널 이후 행들에서 최적의 위치를 찾는 정교한 함수
  const findBestPositionInLaterRows = (
    originalLayout: any,
    needWidth: number,
    needHeight: number,
  ): {x: number; y: number; needsPush: boolean}[] => {
    const candidates: {x: number; y: number; needsPush: boolean; score: number}[] = [];
    const allYPositions = [...new Set(dashboard.panels!.map((p) => p.layout.y))].sort(
      (a, b) => a - b,
    );
    const laterRows = allYPositions.filter((y) => y > originalLayout.y);

    for (const rowY of laterRows) {
      const maxRowHeight = Math.max(
        1,
        ...dashboard
          .panels!.filter((p) => p.layout.y <= rowY && p.layout.y + p.layout.h > rowY)
          .map((p) => p.layout.h),
      );

      // 각 Y 위치에서 빈 공간 찾기
      for (let testY = rowY; testY <= rowY + maxRowHeight - needHeight; testY++) {
        const emptyX = findEmptySpaceInRow(testY, needWidth, needHeight, originalLayout.x);

        if (emptyX !== null) {
          // 점수 계산: 원본 패널에 가까울수록, X 위치가 비슷할수록 높은 점수
          const yDistance = Math.abs(testY - originalLayout.y);
          const xDistance = Math.abs(emptyX - originalLayout.x);
          const score = 1000 - yDistance * 10 - xDistance * 2;

          candidates.push({
            x: emptyX,
            y: testY,
            needsPush: false,
            score: score,
          });
        }
      }

      // 해당 행에 공간이 부족하다면 패널을 아래로 밀어내는 옵션도 고려
      if (candidates.length === 0) {
        const panelsInRow = dashboard.panels!.filter(
          (p) => p.layout.y <= rowY && p.layout.y + p.layout.h > rowY,
        );

        if (panelsInRow.length > 0) {
          let testX = originalLayout.x;

          if (testX + needWidth <= GRID_COLUMN_COUNT) {
            const yDistance = Math.abs(rowY - originalLayout.y);
            const xDistance = Math.abs(testX - originalLayout.x);
            const score = 500 - yDistance * 10 - xDistance * 2;

            candidates.push({
              x: testX,
              y: rowY,
              needsPush: true,
              score: score,
            });
          }

          testX = Math.max(0, originalLayout.x - needWidth);
          if (testX >= 0) {
            const yDistance = Math.abs(rowY - originalLayout.y);
            const xDistance = Math.abs(testX - originalLayout.x);
            const score = 500 - yDistance * 10 - xDistance * 2;

            candidates.push({
              x: testX,
              y: rowY,
              needsPush: true,
              score: score,
            });
          }
        }
      }
    }

    return candidates
      .sort((a, b) => b.score - a.score)
      .map(({x, y, needsPush}) => ({x, y, needsPush}));
  };

  const pushPanelsDown = (insertY: number, pushHeight: number): Panel[] => {
    return dashboard.panels!.map((panel) => {
      if (panel.layout.y >= insertY) {
        return {
          ...panel,
          layout: {
            ...panel.layout,
            y: panel.layout.y + pushHeight,
          },
        };
      }
      return panel;
    });
  };

  const originalLayout = originalPanel.layout;
  let updatedPanels = dashboard.panels;

  const optimalPosition = findOptimalPosition(originalLayout, originalLayout.w, originalLayout.h);

  if (optimalPosition.needsPush) {
    updatedPanels = pushPanelsDown(optimalPosition.y, originalLayout.h);
  }

  newPanel.layout.x = optimalPosition.x;
  newPanel.layout.y = optimalPosition.y;

  const dashboardCopy = {...dashboard};
  dashboardCopy.panels = [...updatedPanels, newPanel];

  return dashboardCopy;
}
