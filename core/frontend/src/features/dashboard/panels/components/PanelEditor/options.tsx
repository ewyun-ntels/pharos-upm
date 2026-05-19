'use client';

import React, {useState, useEffect, useCallback, Suspense} from 'react';
import {LoadingIndicator} from '@pharos/shared/components/ui-extension';
import {useNavigate} from 'react-router-dom';
import {Input} from '@pharos/shared/components/ui';
import {Textarea} from '@pharos/shared/components/ui';
import {Button} from '@pharos/shared/components/ui';
import {Switch} from '@pharos/shared/components/ui';
import {SelectBox} from '@pharos/shared/components/ui-extension/select/box';
import {useDashboardStore, useDashboardData} from '@features/dashboard/hooks/use-dashboard-store';
import {panelPluginRegistry} from '@/features/dashboard/panels/registry/PanelPluginRegistry';
import type {PanelEditorOptionsProps} from '@features/dashboard/panels/registry/PanelPluginRegistry';
import type {Panel, DataProvider, ChartQuery} from '@pharos/shared/types/dashboard';
import {useToast} from '@hooks/use-toast';

interface OptionsPanelProps {
  panelId: string;
  dashboardId: string;
  queries: ChartQuery[];
  onDraftChange?: (draft: Partial<Panel>) => void;
}

/**
 * ✅ OptionsPanel - 전체 기능 복원
 *
 * 기능:
 * 1. Chart Type 선택 (SelectBox)
 * 2. 기본 설정 (title, description, bgTransparent)
 * 3. 플러그인별 옵션 UI (PanelPluginRegistry에서 동적 로딩)
 * 4. Save/Cancel 버튼
 *
 * 패턴:
 * - Zustand 직접 사용 (panelId로 접근)
 * - OptionsComponent 동적 로딩
 * - toPanelData 변환 후 store 업데이트
 */
const OptionsPanel = ({panelId, dashboardId, queries, onDraftChange}: OptionsPanelProps) => {
  const navigate = useNavigate();
  const {toast} = useToast();
  const updatePanel = useDashboardStore((state) => state.updatePanel);
  const panelData = useDashboardStore((state) => state.panelMap[panelId]);
  const { permission } = useDashboardData();

  const [isSaving, setIsSaving] = useState(false);

  const [OptionsComponent, setOptionsComponent] = useState<React.ComponentType<PanelEditorOptionsProps<unknown>> | null>(null);

  // ✅ 모든 편집 상태를 로컬 draft로 관리 (Save 전까지 store에 쓰지 않음)
  const [chartType, setChartType] = useState('');
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [bgTransparent, setBgTransparent] = useState(false);
  const [draftOptions, setDraftOptions] = useState<Panel['options']>({});
  const [draftDataProvider, setDraftDataProvider] = useState<DataProvider | undefined>(undefined);

  // ✅ panelId 변경 시(패널 전환 또는 최초 진입) 로컬 draft 초기화
  useEffect(() => {
    if (!panelData) return;
    setChartType(panelData.renderType || '');
    setTitle(panelData.title || '');
    setDescription(panelData.description || '');
    setBgTransparent(panelData.bgTransparent || false);
    setDraftOptions(panelData.options || {});
    setDraftDataProvider(panelData.dataProvider);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [panelId]);

  // ✅ OptionsComponent 로드 (chartType 변경 시)
  useEffect(() => {
    if (!chartType) {
      setOptionsComponent(null);
      return;
    }

    let isMounted = true;

    panelPluginRegistry
      .loadEditorConfig(chartType)
      .then((editorConfig) => {
        if (isMounted && editorConfig?.OptionsComponent) {
          setOptionsComponent(() => editorConfig.OptionsComponent);
        } else if (isMounted) {
          setOptionsComponent(null);
        }
      })
      .catch((error) => {
        console.error('[OptionsPanel] Failed to load editor config:', error);
        if (isMounted) {
          setOptionsComponent(null);
        }
      });
    return () => {
      isMounted = false;
    };
  }, [chartType]);

  // ✅ Handlers - 모두 로컬 draft만 업데이트, store 직접 쓰기 없음
  const handleChartTypeChange = useCallback(
    (value: string) => {
      setChartType(value);
      setDraftOptions({});
      onDraftChange?.({renderType: value, options: {}});
    },
    [onDraftChange],
  );

  const handleTitleChange = useCallback(
    (value: string) => {
      setTitle(value);
      onDraftChange?.({title: value});
    },
    [onDraftChange],
  );

  const handleDescriptionChange = useCallback(
    (value: string) => {
      setDescription(value);
      onDraftChange?.({description: value});
    },
    [onDraftChange],
  );

  const handleBgTransparentChange = useCallback(
    (value: boolean) => {
      setBgTransparent(value);
      onDraftChange?.({bgTransparent: value});
    },
    [onDraftChange],
  );

  const handleOptionsChange = useCallback(
    (newOptions: unknown) => {
      setDraftOptions(newOptions as Panel['options']);
      onDraftChange?.({options: newOptions as Panel['options']});
    },
    [onDraftChange],
  );

  const handleDataProviderChange = useCallback(
    (newDataProvider: DataProvider) => {
      setDraftDataProvider(newDataProvider);
      onDraftChange?.({dataProvider: newDataProvider});
    },
    [onDraftChange],
  );

  // ✅ Save 버튼 - store 업데이트 후 백그라운드 저장 및 대시보드로 이동
  const handleSave = useCallback(async () => {
    if (!panelId || !panelData) return;

    const validQueries = queries.filter((q) => q.datasourceName && q.query.trim());

    setIsSaving(true);
    try {
      // ✅ 최종 데이터 준비: 로컬 draft state 사용 (queries 포함)
      const dataProviderWithQueries = validQueries.length > 0 ? {
        ...draftDataProvider,
        chartQuery: validQueries,
      } : draftDataProvider;

      const data: Panel = {
        ...panelData,
        renderType: chartType,
        title,
        description,
        bgTransparent,
        options: draftOptions,
        dataProvider: dataProviderWithQueries,
      };

      const editorConfig = await panelPluginRegistry.loadEditorConfig(chartType);

      let finalData: Panel;

      if (editorConfig?.toPanelData) {
        const custom_data = await editorConfig.toPanelData(data, permission);

        finalData = {
          ...panelData,
          ...custom_data,
          renderType: chartType,
          bgTransparent,
          title,
          description,
          options: draftOptions,
          dataProvider: dataProviderWithQueries,
        };
      } else {
        finalData = {
          ...panelData,
          renderType: chartType,
          bgTransparent,
          title,
          description,
          options: draftOptions,
          dataProvider: dataProviderWithQueries,
        };
      }

      // dataProvider 필요 여부에 따라 제거
      const noDataProviderTypes = ['markdownViewer', 'markdownEditor', 'healthBox'];

      if (noDataProviderTypes.includes(chartType)) {
        delete finalData.dataProvider;
      }

      // 패널 업데이트 (store에만 반영, 백엔드 저장은 대시보드 Save 버튼에서 처리)
      updatePanel(panelId, finalData);

      // 즉시 대시보드로 복귀 (URL params 유지)
      const params = new URLSearchParams(window.location.search);
      navigate(`/dashboards/${dashboardId}${params.toString() ? '?' + params.toString() : ''}`);
    } catch (error) {
      console.error('Error applying panel changes:', error);
      toast({
        title: 'Apply failed',
        description: error instanceof Error ? error.message : 'Unknown error',
        variant: 'destructive',
      });
    } finally {
      setIsSaving(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    panelId,
    panelData,
    chartType,
    title,
    description,
    bgTransparent,
    draftOptions,
    draftDataProvider,
    queries,
    updatePanel,
    navigate,
    dashboardId,
    toast,
  ]);

  // ✅ Cancel 버튼 - 로컬 draft 버리고 그냥 이동 (store는 건드리지 않았으므로 원상태)
  const handleCancel = useCallback(() => {
    const params = new URLSearchParams(window.location.search);
    navigate(`/dashboards/${dashboardId}${params.toString() ? '?' + params.toString() : ''}`);
  }, [navigate, dashboardId]);

  if (!dashboardId) {
    return <div className="h-full w-full p-4">No dashboard ID provided</div>;
  }

  if (!panelData) {
    return (
      <div className="h-full w-full p-4">
        <div className="text-sm text-muted-foreground">Panel not found: {panelId}</div>
        <div className="text-xs text-muted-foreground mt-2">Waiting for panel to load...</div>
      </div>
    );
  }

  return (
    <div className="w-full h-full px-4 py-4 border-l border-border bg-background overflow-y-auto">
      <div className="flex justify-end gap-2 mb-4">
        <Button 
          size="default" 
          variant="outline"
          disabled={isSaving}
          onClick={handleCancel}
        >
          Cancel
        </Button>
        <Button 
          size="default" 
          disabled={isSaving}
          onClick={handleSave}
        >
          {isSaving ? 'Saving...' : 'Save'}
        </Button>
      </div>
      <div className="flex justify-between items-center mb-2">
        <SelectBox
          options={panelPluginRegistry.getAllInfoAsSelectOptions()}
          onChange={handleChartTypeChange}
          value={chartType}
          size="full"
          className="!h-12"
          autoSelectFirstOption={false}
          placeholder="Select panel type"
        ></SelectBox>
      </div>

      {/* Basic Settings - 모든 패널 공통 */}
      <div className="mb-4 px-1">
        <h3 className="text-sm font-semibold mb-3">Basic Settings</h3>

        {/* Title */}
        <div className="mb-3 flex flex-col gap-2">
          <label htmlFor="title">
            <small className="text-sm text-muted-foreground font-right w-fit text-nowrap leading-none">
              Title
            </small>
          </label>
          <Input
            id="title"
            type="text"
            value={title}
            onChange={(e) => handleTitleChange(e.target.value)}
            className="px-2 py-1"
          />
        </div>

        {/* Description */}
        <div className="mb-3 flex flex-col gap-2">
          <label htmlFor="description">
            <small className="text-sm text-muted-foreground font-right w-fit text-nowrap leading-none">
              Description
            </small>
          </label>
          <Textarea
            id="description"
            value={description}
            onChange={(e) => handleDescriptionChange(e.target.value)}
            className="px-2 py-1 shadow-none"
          />
        </div>

        {/* Background Transparent */}
        <div className="mb-3 flex items-center justify-between">
          <label htmlFor="bgTransparent">
            <small className="text-sm text-muted-foreground font-right w-fit text-nowrap leading-none">
              Background transparent
            </small>
          </label>
          <Switch
            id="bgTransparent"
            checked={bgTransparent}
            onCheckedChange={handleBgTransparentChange}
          />
        </div>
      </div>

      {/* Panel-specific Options */}
      <div className="px-1">
        {OptionsComponent ? (
          <Suspense fallback={<LoadingIndicator className="h-40" />}>
            <OptionsComponent
              dataProvider={draftDataProvider}
              onDataProviderChange={handleDataProviderChange}
              options={draftOptions}
              onOptionsChange={handleOptionsChange}
            />
          </Suspense>
        ) : null}
        {!OptionsComponent && chartType && (
          <div className="text-sm text-muted-foreground">Loading options for {chartType}...</div>
        )}
      </div>
    </div>
  );
};

export default OptionsPanel;
