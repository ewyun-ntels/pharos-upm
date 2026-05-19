/**
 * Panel Plugin Registry - Main Export
 * 
 * 내부(core) 및 외부(extension) 모두에서 사용 가능한 통합 export
 */

// Core registry and types
export {panelPluginRegistry} from './PanelPluginRegistry';
export type {
  PanelPlugin,
  PanelPluginInfo,
  PanelProps,
  PanelEditorConfig,
  PanelEditorOptionsProps,
  PanelAction,
  PanelActionAPI,
} from './PanelPluginRegistry';

// Re-export DataProvider from shared
export type {DataProvider} from '@pharos/shared/types/dashboard';

// Variable 관련 타입 (extensions에서 filterMetas 활용 시 필요)
export type {FilterMeta} from '../../hooks/slices/types';
export type {FilterValue} from '../../types/filter.types';

// Panel 개발 공통 유틸리티 (core + extension 모두 사용 가능)
//
// 차트 패널 (Card 포함): component: withCardChartPanel(MyChartContent)
// Card 없이 전체 영역:   component: withChartPanel(MyContent)
//
// withTimeseriesCard는 deprecated - withCardChartPanel 사용 권장
export {withChartPanel, ChartDataWrapper} from '../plugins/common/withChartPanel';
export type {PanelContentProps} from '../plugins/common/withChartPanel';
export {withCardChartPanel, withTimeseriesCard} from '../plugins/common/withCardChartPanel';

/**
 * 새로운 등록 방식:
 *
 * 1. Core 패널: 각 panel.plugin.ts에서 직접 panelPluginRegistry.register() 호출
 * 2. Extension: index.ts에서 직접 panelPluginRegistry.register() 호출
 * 3. 모든 등록이 명시적이고 추적 가능
 *
 * ✅ 장점:
 * - webpack magic 제거
 * - 디버깅 용이
 * - Extension과 일관된 방식
 * - 명시적 의존성 관리
 */
