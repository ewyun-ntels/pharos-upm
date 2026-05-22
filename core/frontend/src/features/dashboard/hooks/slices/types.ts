/**
 * Dashboard Store Shared Types
 *
 * 모든 slice에서 공유하는 타입 정의.
 * 순환 의존성 방지를 위해 타입을 별도 파일로 분리.
 * use-dashboard-store.ts에서 re-export하여 기존 import 경로 유지.
 */
import type {
  DashboardConfig,
  Panel,
  FilterConfig,
  DashboardCreateResponse,
  Action,
  Annotation,
} from '@pharos/shared/types/dashboard';
import type { DateTimeRangeValue } from '@pharos/shared/components/ui-extension';
import type { FilterValue, DateTime, Step } from '../../types/filter.types';
import { DependencyGraph } from '../../utils/dependency-graph';

// ─── FilterMeta ──────────────────────────────────────────────────────────────

export interface FilterMeta {
  // === 설정 (변하지 않음) ===
  id: string;
  query: string;
  datasourceName: string;
  datasourceType?: string;
  options?: FilterConfig['options'];

  // === 런타임 상태 (변함) ===
  value?: FilterValue;
  isFetching: boolean;
  isQuerySuccess: boolean;
  error?: string;
  data?: any;
  lastRefreshCount?: number;
}

// ─── FilterState ─────────────────────────────────────────────────────────────

export interface FilterState {
  filterMetas: Map<string, FilterMeta>;

  datetime?: DateTime;
  step?: Step;
  stepOptions?: string[];
  rangeStepOptions?: Record<string, string[]>;
  refreshCount: number;
  refreshInterval?: number | false;
}

export function createFilterState(): FilterState {
  return {
    filterMetas: new Map(),
    datetime: undefined,
    step: undefined,
    stepOptions: undefined,
    refreshCount: 0,
    refreshInterval: undefined,
  };
}

// ─── DependencyGraph (Store 외부 ref) ────────────────────────────────────────
// DependencyGraph는 클래스 인스턴스라 Immer freeze와 호환되지 않는다.
// Store 외부의 ref로 관리하여 setAutoFreeze(false) 없이도 안전하게 동작하도록 한다.

let _dependencyGraph = new DependencyGraph();

export function getDependencyGraph(): DependencyGraph {
  return _dependencyGraph;
}

export function setDependencyGraph(graph: DependencyGraph): void {
  _dependencyGraph = graph;
}

export function resetDependencyGraph(): void {
  _dependencyGraph = new DependencyGraph();
}

// ─── DashboardStore ──────────────────────────────────────────────────────────

export type LoadDashboardOptions = {
  forceReload?: boolean;
  urlParams?: URLSearchParams;
  datasourceTypeMap?: Record<string, string>;
};

export type DashboardStore = {
  // === 기본 상태 ===
  id: string | undefined;
  permission: Action;
  dashboard: DashboardConfig | undefined;
  hasUnsavedChanges: boolean;

  // === 패널 관리 ===
  panelOrder: string[];
  panelMap: Record<string, Panel>;

  // === 필터 설정 (서버) ===
  filters: FilterConfig[];

  // === 필터 런타임 상태 ===
  filterState: FilterState;

  // === API 상태 ===
  loading: boolean;
  error: unknown;

  // === Annotation 상태 ===
  storeAnnotations: Annotation[];

  // === 대시보드 CRUD 액션 ===
  createDashboard: (data: Partial<DashboardConfig>) => Promise<DashboardCreateResponse>;
  updateDashboard: (updates: Partial<DashboardConfig>) => void;
  saveDashboard: (options?: { saveTimeRange?: boolean; saveStep?: boolean; saveRefreshInterval?: boolean }) => Promise<void>;
  deleteDashboard: (id: string) => Promise<void>;
  duplicateDashboard: (id: string) => Promise<DashboardCreateResponse>;

  // === 패널 관리 액션 ===
  createPanel: (panelType?: string) => string;
  addPanel: (panel: Partial<Omit<Panel, 'id'>>) => void;
  updatePanel: (
    panelId: string,
    updates: Partial<Panel> | ((panel: Panel) => Partial<Panel>),
  ) => void;
  updatePanelOptions: (panelId: string, optionUpdates: NonNullable<Panel['options']>) => void;
  batchUpdatePanels: (updates: Array<{ panelId: string; updates: Partial<Panel> }>) => void;
  deletePanel: (panelId: string) => void;
  duplicatePanel: (panelId: string) => Promise<void>;
  getPanel: (panelId: string) => Panel | undefined;

  // === 필터/변수 관리 ===
  setFilters: (filters: FilterConfig[] | ((prev: FilterConfig[]) => FilterConfig[])) => void;

  // === 필터 런타임 액션 ===
  setFilter: (id: string, value: FilterValue, options?: { skipCascade?: boolean }) => void;
  getFilter: (id: string) => FilterValue | undefined;
  setFilterMeta: (id: string, updates: Partial<FilterMeta>) => void;
  setDatetime: (startTime: DateTimeRangeValue, endTime: DateTimeRangeValue, options?: { skipRefreshCount?: boolean }) => void;
  setStep: (step: Step) => void;
  setRefreshInterval: (interval: number | false) => void;
  incrementRefreshCount: () => void;
  setFilterMetas: (filterMetas: Omit<FilterMeta, 'value' | 'isFetching' | 'isQuerySuccess' | 'error' | 'data'>[]) => { success: boolean; errors: string[] };
  setStepOptions: (options: string[] | undefined) => void;
  setRangeStepOptions: (options: Record<string, string[]>) => void;

  // === 즐겨찾기 ===
  updateFavorite: (id: string, favorite: boolean, data?: Partial<DashboardConfig>) => Promise<void>;

  // === Annotation 액션 ===
  setAnnotations: (annotations: Annotation[]) => void;
  addAnnotation: (annotation: Annotation) => Promise<void>;
  removeAnnotation: (annotationId: string) => Promise<void>;

  // === 내부 로딩 ===
  _loadDashboard: (dashboardId: string, options?: LoadDashboardOptions) => Promise<void>;

  // === 유틸리티 ===
  reset: () => void;
};

// ─── Slice 유틸리티 타입 ─────────────────────────────────────────────────────

/** Zustand set 함수 타입 (slice에서 사용) */
export type StoreSet = {
  (partial: Partial<DashboardStore>): void;
  (updater: (state: DashboardStore) => DashboardStore | Partial<DashboardStore>): void;
};

/** Zustand get 함수 타입 (slice에서 사용) */
export type StoreGet = () => DashboardStore;
