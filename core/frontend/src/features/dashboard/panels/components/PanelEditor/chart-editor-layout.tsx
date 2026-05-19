'use client';

import LeftBottomPanel from '@features/dashboard/panels/components/PanelEditor/left-bottom';
import LeftTopPanel from '@features/dashboard/panels/components/PanelEditor/left-top';
import OptionsPanel from '@features/dashboard/panels/components/PanelEditor/options';
import React, { useCallback, useEffect, useState, useMemo } from 'react';
import { FilterConfig } from '@features/dashboard/components/VariableBar/types';
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from '@pharos/shared/components/ui';
import { useDashboardStore } from '@features/dashboard/hooks/use-dashboard-store';
import { ChartQuery, Panel } from '@pharos/shared/types/dashboard';

interface ChartEditorLayoutProps {
  id: string; // 대시보드 ID
  panelId: string; // 패널 ID
  filters: FilterConfig[] | null;
}

const createEmptyQuery = (): ChartQuery => ({
  datasourceName: '',
  query: '',
  label: '',
});

export default function ChartEditorLayout({ id: dashboardId, panelId, filters }: ChartEditorLayoutProps) {
  const panel = useDashboardStore((state) => state.panelMap[panelId]);
  
  // ✅ queries를 여기서 관리 (left-bottom과 options 공유)
  const [queries, setQueries] = useState<ChartQuery[]>([createEmptyQuery()]);

  // ✅ draft 패턴: options에서 변경된 내용을 임시 보관 (Save 전까지 store 비접촉)
  const [draftPanel, setDraftPanel] = useState<Partial<Panel>>({});
  const handleDraftChange = useCallback((partial: Partial<Panel>) => {
    setDraftPanel((prev) => ({...prev, ...partial}));
  }, []);

  // ✅ 미리보기 전용 forceRefresh (store 비접촉 - Run All Queries용)
  //
  // 왜 이렇게 구현했나?
  // - "Run All Queries" 버튼은 store를 직접 수정하지 않음 (Save 버튼만 store 수정)
  // - 대신 previewForceRefresh timestamp를 변경하여 effectivePanel을 업데이트
  // - effectivePanel은 overridePanel prop으로 left-top에 전달되어 미리보기만 갱신
  // - 이로써 Save 전까지는 store가 깨끗하게 유지되고, Cancel 시 롤백 가능
  const [previewForceRefresh, setPreviewForceRefresh] = useState<number | undefined>(undefined);

  // Run All Queries 콜백: store 직접 수정 없이 effectivePanel에만 반영
  const handleRunQueries = useCallback(() => {
    setPreviewForceRefresh(Date.now());
  }, []);

  // panelId 변경 시 draft 초기화
  useEffect(() => {
    setDraftPanel({});
    setPreviewForceRefresh(undefined);
  }, [panelId]);

  // forceRefresh를 제외한 쿼리 내용 해시
  const queryContentHash = useMemo(() => {
    const chartQueries = panel?.dataProvider?.chartQuery;
    if (!chartQueries || chartQueries.length === 0) return '';
    return JSON.stringify(
      chartQueries.map((q) => ({
        datasourceName: q.datasourceName,
        query: q.query,
        label: q.label,
      })),
    );
  }, [panel?.dataProvider?.chartQuery]);

  // draft + store panel 병합 → LeftTopPanel에 전달할 effectivePanel
  //
  // effectivePanel 구조:
  // 1. store panel + draftPanel 병합 (options 등 UI 변경사항 반영)
  // 2. previewForceRefresh가 있으면 (Run All Queries 클릭 시):
  //    - 로컬 queries를 dataProvider.chartQuery에 덮어쓰기
  //    - forceRefresh timestamp를 각 query에 추가
  //    - 이로써 useDashboardQuery가 변경을 감지하고 재실행
  // 3. previewForceRefresh가 없으면 store의 chartQuery 그대로 사용
  const effectivePanel = useMemo(() => {
    if (!panel) return undefined;
    const merged = {...panel, ...draftPanel} as Panel;

    const validQueries = queries.filter((q) => q.datasourceName && q.query.trim());
    if (previewForceRefresh && validQueries.length > 0) {
      return {
        ...merged,
        dataProvider: {
          ...merged.dataProvider,
          chartQuery: validQueries.map((q) => ({...q, forceRefresh: previewForceRefresh})),
        },
      };
    }

    return merged;
  }, [panel, draftPanel, queries, previewForceRefresh]);

  // panel 변경 시 로컬 state 초기화
  useEffect(() => {
    const chartQueries = panel?.dataProvider?.chartQuery;
    if (chartQueries && chartQueries.length > 0) {
      setQueries(
        chartQueries.map((q) => ({
          datasourceName: q.datasourceName || '',
          query: q.query || '',
          label: q.label || '',
        })),
      );
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [queryContentHash]);

  return (
    <ResizablePanelGroup
      orientation="horizontal"
      className="relative grow h-full overflow-hidden"
    >
      <ResizablePanel
        className="bg-transparent relative"
        defaultSize="75%"
      >
        <ResizablePanelGroup orientation="vertical" className={"w-full"}>
          <ResizablePanel>
            <LeftTopPanel
              filters={filters || []}
              dashboardId={dashboardId}
              panelId={panelId || ''}
              overridePanel={effectivePanel}
            />
          </ResizablePanel>
          <ResizableHandle withHandle className={"justify-center"}/>
          <ResizablePanel>
            <LeftBottomPanel 
              panelId={panelId || ''} 
              queries={queries}
              onQueriesChange={setQueries}
              onRunQueries={handleRunQueries}
            />
          </ResizablePanel>
        </ResizablePanelGroup>
      </ResizablePanel>

      <ResizableHandle withHandle />

      <ResizablePanel
        className="bg-white relative h-full max-h-screen"
      >
        <OptionsPanel 
          panelId={panelId || ''} 
          dashboardId={dashboardId} 
          queries={queries}
          onDraftChange={handleDraftChange}
        />
      </ResizablePanel>
    </ResizablePanelGroup>
  );
}
