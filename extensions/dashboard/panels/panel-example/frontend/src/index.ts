/**
 * Panel Example Extension
 * 
 * 커스텀 패널 개발 가이드를 위한 예시 Extension입니다.
 * 
 * 이 Extension은 두 가지 패널 예시를 제공합니다:
 * 1. SimplePanel - 데이터 소스 없이 UI만 커스터마이징하는 기본 패널
 * 2. ChartPanel - useDashboardQuery hook을 사용하여 시계열 데이터를 로드하는 차트 패널
 * 
 * ## 패널 개발 가이드
 * 
 * ### 1. 패널 컴포넌트 작성
 * - `PanelProps<T>` 타입을 사용하여 props 정의
 * - options에서 사용자 설정 값 받기
 * - pluginContext로 core hooks 접근 (useDashboardQuery 등)
 * 
 * ### 2. 옵션 에디터 작성
 * - `PanelEditorOptionsProps<T>` 타입 사용
 * - value, onChange prop으로 양방향 바인딩
 * 
 * ### 3. 패널 플러그인 등록
 * - info: 패널 메타데이터 (id, label, category 등)
 * - component: 패널 컴포넌트
 * - editor: 옵션 에디터와 데이터 변환 함수
 * 
 * ### 4. Extension 등록
 * - registerExtension으로 extension 메타데이터 등록
 * - panelPluginRegistry.register로 패널 등록
 */

import { SimplePanel, SimplePanelOptions } from './plugins/simple/SimplePanel';
import { SimplePanelOptionsEditor } from './plugins/simple/SimplePanelOptionsEditor';
import { ChartPanel, ChartPanelOptions } from './plugins/chart/ChartPanel';
import { ChartPanelOptionsEditor } from './plugins/chart/ChartPanelOptionsEditor';

import { panelPluginRegistry, withTimeseriesCard } from '@pharos/core/panel-registry';
import type { PanelPlugin } from '@pharos/core/panel-registry';
import { registerExtension } from '@pharos/core/extension-registry';

// Extension metadata
export const metadata = {
  name: 'panel-example',
  version: '1.0.0',
  displayName: 'Panel Examples',
  description: 'Custom panel development examples and templates',
};

/**
 * Simple Panel Plugin
 * 데이터 소스 없이 UI만 커스터마이징하는 기본 패널
 */
export const simplePanelPlugin: PanelPlugin = {
  info: {
    id: 'simple-panel',
    label: 'Simple Panel',
    description: '데이터 소스 없이 UI를 커스터마이징하는 기본 패널 예시',
    category: 'Examples',
  },
  component: SimplePanel,
  editor: async () => ({
    OptionsComponent: SimplePanelOptionsEditor,
    toPanelData: (formData: SimplePanelOptions) => {
      return {
        ...formData,
        type: 'simple-panel',
      };
    },
    toFormData: (panelData: any) => {
      return panelData;
    },
    getDefaults: () => ({
      chartOptions: {
        title: 'Simple Panel',
        backgroundColor: '#f3f4f6',
        textColor: '#1f2937',
        message: 'This is a simple custom panel example',
        icon: '🎨',
      },
    }),
  }),
};

/**
 * Chart Panel Plugin
 * useDashboardQuery hook을 사용하여 시계열 데이터를 로드하는 차트 패널
 */
export const chartPanelPlugin: PanelPlugin = {
  info: {
    id: 'chart-panel-example',
    label: 'Chart Panel Example',
    description: 'useDashboardQuery hook을 사용한 데이터 기반 패널 예시',
    category: 'Examples',
  },
  component: withTimeseriesCard(ChartPanel),
  editor: async () => ({
    OptionsComponent: ChartPanelOptionsEditor,
    toPanelData: (formData: ChartPanelOptions) => {
      return {
        ...formData,
        type: 'chart-panel-example',
      };
    },
    toFormData: (panelData: any) => {
      return panelData;
    },
    getDefaults: () => ({
      chartOptions: {
        title: 'Chart Panel Example',
        color: '#3b82f6',
        showLegend: true,
        showDataPoints: false,
      },
    }),
  }),
};

// Register Panel Plugins
panelPluginRegistry.register(simplePanelPlugin);
panelPluginRegistry.register(chartPanelPlugin);

// Register Extension
registerExtension({
  ...metadata,
});

// Export panel components (optional)
export { SimplePanel } from './plugins/simple/SimplePanel';
export type { SimplePanelOptions } from './plugins/simple/SimplePanel';
export { ChartPanel } from './plugins/chart/ChartPanel';
export type { ChartPanelOptions } from './plugins/chart/ChartPanel';
