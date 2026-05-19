/**
 * Example Extension for Pharos Frontend
 * 
 * This extension demonstrates how to create a frontend extension
 * that integrates with the main Pharos application.
 */

import { ExampleChartPanel, ExampleChartPanelOptions } from './panels/ExampleChartPanel';

// Extension에서 workspace 의존성 사용
import { panelPluginRegistry } from '@pharos/core/panel-registry';
import type { PanelPlugin } from '@pharos/core/panel-registry';
import { registerExtension } from '@pharos/core/extension-registry';

// Extension metadata
export const metadata = {
  name: 'example',
  version: '1.0.0',
  displayName: 'Example Extension',
  description: 'A sample extension for demonstration purposes',
};

// Panel Plugin 정의
export const exampleChartPlugin: PanelPlugin = {
  info: {
    id: 'example-chart',
    label: 'Example Chart',
    description: 'Example chart panel from external extension',
    category: 'Charts',
  },
  component: ExampleChartPanel, // 직접 전달
  editor: async () => ({
    // ✅ info는 plugin.info에서 가져오므로 중복 제거
    OptionsComponent: ExampleChartPanelOptions,
    toPanelData: (formData: any) => {
      // formData를 panel data로 변환
      return {
        ...formData,
        type: 'example-chart',
      };
    },
    toFormData: (panelData: any) => {
      // panel data를 form data로 변환 (편집 시)
      return panelData;
    },
    getDefaults: () => ({
      chartOptions: {
        title: 'Example Chart',
        showLegend: true,
        chartColor: '#3b82f6',
        lineWidth: 2,
        showDataPoints: false,
      },
    }),
  }),
};

// Register Panel Plugin to panelPluginRegistry
panelPluginRegistry.register(exampleChartPlugin);

// Register the extension (패널 플러그인만 포함)
registerExtension({
  ...metadata,
  
  // Panel Plugin은 panelPluginRegistry에 직접 등록됨
});

// Export panel components
export { ExampleChartPanel, ExampleChartPanelOptions } from './panels/ExampleChartPanel';

