'use client';

import React, {useState, useEffect, useLayoutEffect, useCallback, useRef} from 'react';
import {Responsive, LayoutItem, ResponsiveLayouts, Layout, verticalCompactor} from 'react-grid-layout';
import {cn} from '@lib/utils';
import {LoadingIndicator} from '@pharos/shared/components/ui-extension';
import {useDashboardData, useDashboardActions} from '@features/dashboard/hooks/use-dashboard-store';
import {Panel} from '@pharos/shared/types/dashboard';
import {BoxFullView} from './boxFullView';
import {compactLayout, convertJsonToGridLayout, selectiveCompactLayout} from './utils/layoutUtils';
import {PanelRenderer} from './PanelRenderer';
import GridItem from './GridItem';

export default function GridLayout() {
  const {dashboard: dashboardData, permission, id: dashboardId, panels} = useDashboardData();

  const {batchUpdatePanels} = useDashboardActions();

  const containerRef = useRef<HTMLDivElement>(null);
  const [containerWidth, setContainerWidth] = useState<number | null>(null);
  // lazy initializer: 마운트 시점에 panels가 있으면 즉시 올바른 레이아웃으로 초기화
  // ({lg: []} → 실제 레이아웃 변경에 의한 초기 애니메이션 방지)
  const [gridLayouts, setGridLayouts] = useState<ResponsiveLayouts>(() => ({
    lg: panels.length > 0 ? convertJsonToGridLayout(panels) : [],
  }));
  // 레이아웃이 처음 안정화되기 전까지 CSS transition 비활성화
  const [isLayoutReady, setIsLayoutReady] = useState(false);
  const [fullScreenCardId, setFullScreenCardId] = useState<string | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const isDraggingRef = useRef(false);

  // paint 전에 컨테이너 너비를 측정하여 레이아웃 2-pass 재계산(확장→축소 현상) 방지
  useLayoutEffect(() => {
    if (!containerRef.current) return;
    setContainerWidth(containerRef.current.offsetWidth);

    const ro = new ResizeObserver(() => {
      if (containerRef.current) {
        setContainerWidth(containerRef.current.offsetWidth);
      }
    });
    ro.observe(containerRef.current);
    return () => ro.disconnect();
  }, []);

  useEffect(() => {
    const handleMouseDown = (e: MouseEvent) => {
      if (!(e.target instanceof HTMLElement)) return;
      if (e.target.closest('.react-resizable-handle')) {
        document.body.style.userSelect = 'none';
        document.body.style.setProperty('-webkit-user-select', 'none');
      }
    };

    const handleMouseUp = () => {
      document.body.style.userSelect = '';
      document.body.style.removeProperty('-webkit-user-select');
    };

    window.addEventListener('mousedown', handleMouseDown);
    window.addEventListener('mouseup', handleMouseUp);

    return () => {
      window.removeEventListener('mousedown', handleMouseDown);
      window.removeEventListener('mouseup', handleMouseUp);
    };
  }, []);

  const updateLayout = useCallback((panels: Panel[]) => {
    const newLayout = convertJsonToGridLayout(panels);
    setGridLayouts({lg: newLayout});
  }, []);

  // 드래그 중이 아닐 때만 compactLayout 적용 (성능 최적화)
  const onLayoutChange = useCallback(
    (layout: Layout) => {
      if (layout.length === 0 || isDraggingRef.current) return;

      const compactedLayout = compactLayout([...layout], panels);
      setGridLayouts({lg: compactedLayout});
    },
    [panels],
  );

  const onDrag = useCallback((layout: Layout) => {
    setGridLayouts({lg: [...layout]});
  }, []);

  const onDragStart = useCallback((_layout: Layout) => {
    isDraggingRef.current = true;
    setIsDragging(true);
  }, []);

  const onDragStop = useCallback(
    (layout: Layout) => {
      isDraggingRef.current = false;
      setIsDragging(false);

      const compactedLayout = selectiveCompactLayout([...layout], panels);
      setGridLayouts({lg: compactedLayout});

      const hasChanges = panels.some((panel) => {
        const layoutItem = compactedLayout.find((item) => item.i === panel.id.toString());
        if (!layoutItem) return false;

        return (
          panel.layout?.x !== layoutItem.x ||
          panel.layout?.y !== layoutItem.y ||
          panel.layout?.w !== layoutItem.w ||
          panel.layout?.h !== layoutItem.h
        );
      });

      if (hasChanges) {
        // useDashboardStore 중심의 배치 레이아웃 업데이트
        const updates: Array<{panelId: string; updates: Partial<Panel>}> = [];

        panels.forEach((panel) => {
          const layoutItem = compactedLayout.find((item) => item.i === panel.id.toString());
          if (layoutItem) {
            updates.push({
              panelId: panel.id,
              updates: {
                layout: {
                  ...panel.layout,
                  x: layoutItem.x,
                  y: layoutItem.y,
                  w: layoutItem.w,
                  h: layoutItem.h,
                },
              },
            });
          }
        });

        // 배치 업데이트로 성능 최적화
        batchUpdatePanels(updates);
      }
    },
    [panels, batchUpdatePanels],
  );

  const onResizeStop = useCallback(
    (layout: Layout) => {
      onDragStop(layout);
    },
    [onDragStop],
  );

  useEffect(() => {
    if (panels.length > 0) {
      updateLayout(panels);
    }
  }, [panels, updateLayout]);

  // 레이아웃이 처음 그려진 후 transition 활성화 (double rAF로 paint 완료 보장)
  useEffect(() => {
    if ((gridLayouts.lg?.length ?? 0) > 0 && !isLayoutReady) {
      requestAnimationFrame(() =>
        requestAnimationFrame(() => setIsLayoutReady(true))
      );
    }
  // isLayoutReady는 한 번만 실행되어야 하므로 의존성에서 제외
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [gridLayouts.lg?.length]);

  if (!dashboardData || !gridLayouts) {
    return <LoadingIndicator />;
  }

  return (
    <>
      <style>{`
        .dashboard-grid .react-resizable-handle {
          transition: opacity 200ms;
          z-index: 10;
        }
        .dashboard-grid .react-resizable-handle::after {
          border-right: 2px solid var(--color-muted-foreground);
          border-bottom: 2px solid var(--color-muted-foreground);
        }
        .dashboard-grid.is-dragging .drag-handle {
          opacity: 0 !important;
          pointer-events: none !important;
        }
        .dashboard-grid.is-dragging [data-option-menu] {
          opacity: 0 !important;
          pointer-events: none !important;
        }
      `}</style>
      <div
        ref={containerRef}
        className={cn('dashboard-grid relative w-full', !isLayoutReady && '[&_.react-grid-item]:!transition-none', isDragging && 'is-dragging')}
      >
        {containerWidth !== null && (
        <Responsive
          width={containerWidth}
          onLayoutChange={onLayoutChange}
          onDrag={onDrag}
          onDragStart={onDragStart}
          onDragStop={onDragStop}
          onResizeStop={onResizeStop}
          layouts={gridLayouts}
          breakpoints={{lg: 1200, md: 996}}
          cols={{lg: 24, md: 24}}
          rowHeight={40}
          margin={[8, 8] as [number, number]}
          containerPadding={[0, 0] as [number, number]}
          compactor={verticalCompactor}
          className={cn()}
          dragConfig={{handle: '.drag-handle'}}
        >
          {(gridLayouts.lg ?? []).map((item: LayoutItem) => {
            return (
              <div key={item.i}>
                <GridItem item={item} onFullscreen={setFullScreenCardId} dashboardId={dashboardId} />
              </div>
            );
          })}
        </Responsive>
      )}

      {fullScreenCardId &&
        (() => {
          const fullScreenItem = gridLayouts.lg?.find((item) => item.i === fullScreenCardId);
          const panel = panels.find((panel) => panel.id === fullScreenCardId);
          return fullScreenItem && panel ? (
            <BoxFullView open={true} onClose={() => setFullScreenCardId(null)}>
              <PanelRenderer
                item={fullScreenItem}
                panel={panel}
                dashboardId={dashboardId}
                displayName={dashboardData?.displayName || ''}
                permission={permission}
              />
            </BoxFullView>
          ) : null;
        })()}
      </div>
    </>
  );
}
