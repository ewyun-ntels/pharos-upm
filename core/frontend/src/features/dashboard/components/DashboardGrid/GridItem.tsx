'use client';

import React, {memo, useCallback, useState} from 'react';
import {LayoutItem} from 'react-grid-layout';
import {GripVertical} from '@pharos/shared/components';
import {useNavigate} from 'react-router-dom';
import {cn} from '@lib/utils';
import {Button} from '@pharos/shared/components/ui';
import { ActionSchema, Panel } from '@pharos/shared/types/dashboard';
import {BoxDropdown} from './boxDropdown';
import {RowCard} from './row-card';
import {PanelRenderer} from './PanelRenderer';
import {useDashboardActions, useDashboardData} from '@features/dashboard/hooks/use-dashboard-store';

interface GridItemProps {
  item: LayoutItem;
  onFullscreen: (panelId: string) => void;
  dashboardId?: string;
}

const GridItem = memo(
  function GridItem({item, onFullscreen, dashboardId: propDashboardId}: GridItemProps) {
    const navigate = useNavigate();
    const {dashboard, permission, id: storeDashboardId} = useDashboardData();
    const {deletePanel, updatePanel, getPanel, duplicatePanel} = useDashboardActions();

    // props로 전달된 dashboardId 우선, 없으면 store에서 가져온 값 사용
    const dashboardId = propDashboardId || storeDashboardId;

    const [dropdownOpen, setDropdownOpen] = useState(false);

    const panel = getPanel(item.i);
    const displayName = dashboard?.displayName || '';

    const handleDropdownOpenChange = useCallback((open: boolean) => {
      setDropdownOpen(open);
    }, []);

    // 패널 액션 핸들러들
    const handleRemovePanel = useCallback(
      (id: string) => {
        deletePanel(id);
      },
      [deletePanel],
    );

    const handleDuplicatePanel = useCallback(
      (id: string) => {
        // ✅ async 함수의 에러를 명시적으로 처리
        void duplicatePanel(id).catch((error) => {
          console.error('Failed to duplicate panel:', error);
          // TODO: 사용자에게 에러 토스트 표시
        });
      },
      [duplicatePanel],
    );

    const handleToggleRow = useCallback(
      (panel: Panel) => {
        // useDashboardStore 중심의 업데이트 (함수형)
        updatePanel(panel.id, (currentPanel: Panel) => ({
          layout: {
            ...currentPanel.layout,
            collapsed: !currentPanel.layout.collapsed,
          },
        }));
      },
      [updatePanel],
    );

    const handleUpdateRowTitle = useCallback(
      (id: string, title: string) => {
        // useDashboardStore 중심의 업데이트 (함수형)
        updatePanel(id, () => ({title}));
      },
      [updatePanel],
    );

    if (!panel) return null;

    const isRow = panel.layout.type === 'row';
    const isRowOpen = panel.layout.collapsed === false;
    const isRowDraggable = isRow && !isRowOpen;
    const isDraggable = panel.layout?.isDraggable !== false;

    const renderPanel = (item: LayoutItem) => {
      return (
        <PanelRenderer
          item={item}
          panel={panel}
          dashboardId={dashboardId}
          displayName={displayName}
          permission={permission}
        />
      );
    };

    return (
      <div
        className={cn(
          'bg-background border transition-all duration-200 group h-full',
          'hover:z-[9]', // CSS hover로 z-index 처리
          isRow ? 'bg-transparent !p-0 border-none [&_.react-resizable-handle]:hidden' : '',
          !isRowOpen && 'bg-background',
        )}
      >
        {isDraggable && (isRowDraggable || !isRow) && (
          <Button
            variant="ghost"
            className={cn(
              // right-8: 옵션 버튼(32px) 영역과 겹치지 않도록 우측 여백 확보 (viewer는 옵션 버튼 없으므로 제외)
              'drag-handle cursor-move absolute top-0 z-[9] left-0 h-4 p-2 rounded-none hover:bg-transparent',
              permission !== ActionSchema.enum.viewer ? 'right-8' : 'right-0',
              // pointer-events-none: opacity-0 상태에서 마우스 이벤트 차단 (Safari hover 오작동 방지)
              'opacity-0 pointer-events-none group-hover:opacity-100 group-hover:pointer-events-auto transition-opacity duration-200',
              isRow &&
                'right-2 left-auto w-5 h-full opacity-50 pointer-events-auto hover:opacity-100 flex items-center justify-center',
            )}
          >
            {isRow && <GripVertical className="w-4 h-4" />}
          </Button>
        )}
        {permission !== ActionSchema.enum.viewer && (
          <div
            data-option-menu
            className={cn(
              // pointer-events-none: opacity-0 상태에서 마우스 이벤트 차단
              // [transform:translateZ(0)]: Safari GPU 컴포지팅 강제 — SVG 점 렌더링 아티팩트 방지
              'opacity-0 pointer-events-none group-hover:opacity-100 group-hover:pointer-events-auto focus-within:opacity-100 focus-within:pointer-events-auto transition-opacity duration-200 [transform:translateZ(0)]',
              dropdownOpen && 'opacity-100 pointer-events-auto',
            )}
          >
            <BoxDropdown
              isRow={isRow || false}
              onView={() => onFullscreen(item.i)}
              onSettings={() => {
                if (!dashboardId) {
                  console.error('[GridItem] dashboardId is required');
                  return;
                }
                const params = new URLSearchParams(window.location.search);
                const panelId = item.i;
                navigate(`/dashboards/${dashboardId}/edit/${panelId}${params.toString() ? '?' + params.toString() : ''}`);
              }}
              onDuplicate={() => handleDuplicatePanel(item.i)}
              onDelete={() => handleRemovePanel(item.i)}
              onOpenChange={handleDropdownOpenChange}
            />
          </div>
        )}
        <div className="h-full">
          {isRow ? (
            <RowCard
              rowData={panel}
              isOpen={panel.layout.collapsed === false}
              toggleRow={() => handleToggleRow(panel)}
              removeBox={() => handleRemovePanel(item.i)}
              updateRowTitle={handleUpdateRowTitle}
            />
          ) : (
            renderPanel(item)
          )}
        </div>
      </div>
    );
  },
  (prevProps, nextProps) => {
    return (
      prevProps.item.x === nextProps.item.x &&
      prevProps.item.y === nextProps.item.y &&
      prevProps.item.w === nextProps.item.w &&
      prevProps.item.h === nextProps.item.h &&
      prevProps.item.i === nextProps.item.i
    );
  },
);

export default GridItem;
