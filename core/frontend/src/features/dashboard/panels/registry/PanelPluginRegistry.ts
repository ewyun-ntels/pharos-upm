import React from 'react';
import { ChartQueryArgs } from '@types';
import { DateTimeRangeValue } from '@pharos/shared/components/ui-extension';
import { getPluginContext, PluginContext } from '@utils/pluginHooks';
import type { Action, Annotation, DataProvider, Kind, Panel } from '@pharos/shared/types/dashboard';
import type { FilterMeta } from '../../hooks/slices/types';

type TimeRangeCallback = (
  rangeStartTime: DateTimeRangeValue,
  rangeEndTime: DateTimeRangeValue,
) => void;

/**
 * PanelProps: Panel 컴포넌트가 렌더링 시 받는 props
 *
 * 🎯 역할: SharedCard (DB 데이터) + 런타임 필드
 * 📍 위치: Panel 컴포넌트 직접 사용
 *
 * 🔄 데이터 흐름:
 * 1. SharedCard (DB) → toPanelData → PanelProps (options 타입 변환)
 * 2. CardRenderer → PanelProps에 런타임 주입 (args, callbacks)
 * 3. PanelProps → Panel Component (최종 렌더링)
 *
 * ⚠️ SharedCard와의 차이:
 * - SharedCard: options는 any, 순수 DB 데이터
 * - PanelProps: options는 타입 안전<T>, + 런타임 필드 (args, callbacks)
 *
 * @template T - Plugin-specific options 타입 (PiePanelOptions, TablePanelOptions 등)
 * @see SharedCard - DB 스키마 (shared/types/dashboard)
 */
export interface PanelProps<T> {
  // DB 필드 (SharedCard와 동일)
  id?: string;
  title?: string;
  renderType?: string;
  description?: string;
  bgTransparent?: boolean;
  dataProvider?: DataProvider; // chartQuery는 이미 배열!

  // Plugin-specific options (타입 안전)
  options?: T;

  // 런타임 필드 (CardRenderer에서 주입)
  args?: ChartQueryArgs;
  refetchInterval?: number | false;
  timeRangeCallback?: TimeRangeCallback;
  onPanelUpdate?: (updates: Partial<T>) => void;

  // Panel action dispatch (plugin → dashboard)
  onPanelAction?: (action: PanelAction) => void;

  // Dashboard context
  dashboardId?: string;
  displayName?: string;
  kind?: Kind;
  permission: Action;

  // Panel-scoped annotations (scope === 'panel' && panel_id === id)
  annotations?: Annotation[];
  // Global annotations (scope === 'global')
  globalAnnotations?: Annotation[];
  onAnnotationCreate?: (annotation: Annotation) => Promise<void>;
  onAnnotationDelete?: (annotationId: string) => Promise<void>;

  // Variable 치환용 (replaceVariables, 드릴다운 링크 등에서 활용)
  filterMetas?: Map<string, FilterMeta>;

  // Plugin Context (외부 플러그인용)
  pluginContext?: PluginContext;
  info?: PanelPluginInfo;
  key?: number;
}

export interface PanelPluginInfo {
  /** 플러그인의 고유 식별자 (내부적으로 value로 사용) */
  id: string;
  /** 사용자에게 표시되는 이름 (내부적으로 label로 사용) */
  label: string;
  /** SelectBox 호환성을 위한 value (id의 alias) */
  value?: string;
  icon?: React.ReactNode;
  description?: string;
  category?: string;
}

/**
 * Panel Editor Options Props
 * PanelProps와 일관성을 유지하며, 에디터에서 실제 Panel로 전달되는 구조
 *
 * ✅ Props 기반 구조 - React Hook Form context 제거
 */
export interface PanelEditorOptionsProps<T> {
  // Data Provider (SharedCard와 동일)
  dataProvider?: DataProvider; // chartQuery는 이미 배열!
  onDataProviderChange?: (newDataProvider: DataProvider) => void;
  // Plugin-specific options
  options?: T; // 플러그인별 옵션 (타입 안전)
  onOptionsChange?: (newOptions: T) => void;
  // Panel metadata
  info?: PanelPluginInfo;
}

/**
 * Panel Editor용 설정
 *
 * ⚠️ info는 PanelPlugin에서 관리 - 중복 제거
 *
 * @template T - Plugin-specific options 타입
 */
export interface PanelEditorConfig<T> {
  /** 옵션 UI 컴포넌트 - PanelEditorOptionsProps를 받음 */
  OptionsComponent: React.ComponentType<PanelEditorOptionsProps<T>>;

  /**
   * Editor Form 데이터 → Panel Props 변환
   * Editor에서 저장한 데이터를 Panel이 렌더링할 형식으로 변환
   * @param formData - Panel 데이터
   * @param permission - 사용자 권한 (viewer일 때 dashboardProvider 사용)
   */
  // 데이터 변환만 담당 (permission 등 런타임 props는 PanelRenderer가 주입)
  toPanelData: (panel: Panel, permission: Action) => Partial<Panel> | Promise<Partial<Panel>>;

  /**
   * Panel Props → Editor Form 데이터 역변환 (편집 시)
   * DB에서 불러온 Panel 데이터를 Editor가 편집할 형식으로 변환
   */
  toFormData?: (panelData: PanelProps<T>) => unknown;
}

/**
 * PanelActionAPI — 액션 핸들러가 대시보드에 할 수 있는 것들
 * 베이스가 노출하는 제한된 인터페이스 (플러그인이 직접 store 접근 불가)
 */
export interface PanelActionAPI<T> {
  getOptions: () => T | undefined;
  updateOptions: (updates: Partial<T>) => void;
}

/**
 * PanelAction — 플러그인이 emit하는 이벤트
 * type 문자열은 각 플러그인이 소유 (베이스는 의미를 모름)
 */
export interface PanelAction<TPayload = unknown> {
  type: string;
  payload: TPayload;
}

export interface PanelPlugin<T> {
  info: PanelPluginInfo;
  /** 패널 컴포넌트 - PanelProps를 받아야 함 (직접 전달 or lazy loading) */
  component:
  | React.ComponentType<PanelProps<T>>
  | (() => Promise<React.ComponentType<PanelProps<T>>>);
  /** Panel Editor 설정 (옵션 UI + 데이터 변환) */
  editor?: () => Promise<PanelEditorConfig<T>>;
  /**
   * 플러그인이 자신의 액션 타입에 대한 핸들러를 선언.
   * PanelRenderer는 타입을 모르고 그냥 위임만 함.
   */
  actionHandlers?: Record<string, (payload: unknown, api: PanelActionAPI<T>) => void>;
}

class PanelPluginRegistry {
  private plugins = new Map<string, PanelPlugin<unknown>>();
  // ✅ React.lazy() 및 로드된 컴포넌트 캐시 (renderType별 stable reference 보장)
  private lazyComponentCache = new Map<string, React.LazyExoticComponent<React.ComponentType<PanelProps<unknown>>>>();
  private componentCache = new Map<string, React.ComponentType<PanelProps<unknown>>>();
  // ✅ editorConfig 캐시: plugin.editor()는 async import이므로 최초 1회만 호출
  private editorConfigCache = new Map<string, PanelEditorConfig<unknown> & { info: PanelPluginInfo }>();

  register<T>(plugin: PanelPlugin<T>): void {
    // id를 키로 사용 (value가 있으면 value 사용)
    const key = plugin.info.value || plugin.info.id;
    this.plugins.set(key, plugin as PanelPlugin<unknown>);
  }

  unregister(type: string): boolean {
    return this.plugins.delete(type);
  }

  get(type: string): PanelPlugin<unknown> | undefined {
    return this.plugins.get(type);
  }

  getAll(): PanelPlugin<unknown>[] {
    return Array.from(this.plugins.values());
  }

  getAllInfo(): PanelPluginInfo[] {
    return Array.from(this.plugins.values()).map((plugin) => plugin.info);
  }

  /**
   * SelectBox와 호환되는 형태로 플러그인 정보 반환
   * id를 value로, label을 그대로 사용
   */
  getAllInfoAsSelectOptions(): Array<{ label: string; value: string; icon?: React.ReactNode }> {
    return Array.from(this.plugins.values()).map((plugin) => ({
      label: plugin.info.label,
      value: plugin.info.value || plugin.info.id,
      icon: plugin.info.icon,
      description: plugin.info.description,
      category: plugin.info.category,
    }));
  }

  exists(type: string): boolean {
    return this.plugins.has(type);
  }

  async loadComponent(type: string): Promise<React.ComponentType<PanelProps<unknown>> | null> {
    // ✅ 이미 로드된 컴포넌트는 캐시에서 즉시 반환 (중복 로드 방지)
    if (this.componentCache.has(type)) {
      return this.componentCache.get(type)!;
    }

    const plugin = this.get(type);
    if (!plugin) return null;

    try {
      const component = plugin.component;

      // 함수 호출 시도 (lazy loading 함수인 경우)
      if (typeof component === 'function') {
        try {
          const lazyLoader = component as () => Promise<React.ComponentType<PanelProps<unknown>>>;
          const result = lazyLoader();
          // Promise를 반환하면 lazy loading
          if (result && typeof result.then === 'function') {
            const importedModule = await result as { default?: React.ComponentType<PanelProps<unknown>> } | React.ComponentType<PanelProps<unknown>>;
            const resolved = (importedModule as { default?: React.ComponentType<PanelProps<unknown>> }).default
              ?? importedModule as React.ComponentType<PanelProps<unknown>>;
            this.componentCache.set(type, resolved);
            return resolved;
          }
          // Promise가 아니면 컴포넌트 자체 반환 (함수형 컴포넌트)
          this.componentCache.set(type, component as React.ComponentType<PanelProps<unknown>>);
          return component as React.ComponentType<PanelProps<unknown>>;
        } catch (e) {
          // 호출 실패 시 컴포넌트 자체 반환 (클래스/함수형 컴포넌트)
          this.componentCache.set(type, component as React.ComponentType<PanelProps<unknown>>);
          return component as React.ComponentType<PanelProps<unknown>>;
        }
      }

      // 함수가 아니면 그대로 반환
      this.componentCache.set(type, component as React.ComponentType<PanelProps<unknown>>);
      return component as React.ComponentType<PanelProps<unknown>>;
    } catch (error) {
      console.error(`Failed to load component for type: ${type}`, error);
      return null;
    }
  }

  /**
   * renderType별로 안정적인 React.lazy() 컴포넌트를 반환 (캐시)
   * PanelRenderer에서 useMemo 대신 사용하여 anti-pattern 제거
   * 같은 renderType은 항상 동일한 lazy 인스턴스를 반환하므로 remount 시에도 재로드 없음
   */
  getLazyComponent(type: string): React.LazyExoticComponent<React.ComponentType<PanelProps<unknown>>> | null {
    if (!this.plugins.has(type)) return null;

    if (!this.lazyComponentCache.has(type)) {
      const lazyComp = React.lazy(async () => {
        const component = await this.loadComponent(type);
        if (!component) {
          throw new Error(`Failed to load panel component: ${type}`);
        }
        return { default: component };
      });
      this.lazyComponentCache.set(type, lazyComp);
    }

    return this.lazyComponentCache.get(type)!;
  }

  /**
   * Panel Editor 설정 로드 (옵션 UI + 데이터 변환 함수)
   *
   * ✅ plugin.info를 함께 반환하여 중복 제거
   */
  async loadEditorConfig(
    type: string,
  ): Promise<(PanelEditorConfig<unknown> & { info: PanelPluginInfo }) | null> {
    // ✅ 캐시 히트: plugin.editor()는 async import이므로 최초 1회만 실행
    if (this.editorConfigCache.has(type)) {
      return this.editorConfigCache.get(type)!;
    }

    const plugin = this.get(type);
    if (!plugin?.editor) return null;

    try {
      const editorConfig = await plugin.editor() as PanelEditorConfig<unknown> | { default: PanelEditorConfig<unknown> };
      const config = (editorConfig as { default?: PanelEditorConfig<unknown> }).default ?? editorConfig as PanelEditorConfig<unknown>;

      // plugin.info를 config에 추가하여 반환 (중복 제거)
      const result = {
        ...config,
        info: plugin.info,
      };
      this.editorConfigCache.set(type, result);
      return result;
    } catch (error) {
      console.error(`Failed to load editor config for type: ${type}`, error);
      return null;
    }
  }

  /**
   * 패널을 렌더링합니다. Suspense와 에러 처리를 포함합니다.
   * Plugin Context를 자동으로 주입하여 외부 플러그인이 hooks와 필드를 사용할 수 있게 합니다.
   *
   * @param type - 패널 타입 (renderType)
   * @param props - 패널 컴포넌트에 전달할 props
   * @param fallback - 로딩 중 표시할 컴포넌트 (선택사항)
   * @returns React Element 또는 에러 메시지
   */
  render<T>(type: string, props: PanelProps<T>, fallback?: React.ReactElement): React.ReactElement {
    const registry = this;

    // ✅ React Component로 만들어서 store 구독 가능하게!
    const PanelRenderer: React.FC = () => {
      // 플러그인 존재 여부 확인
      if (!registry.exists(type)) {
        return React.createElement(
          'div',
          {
            className:
              'flex items-center justify-center w-full h-full text-muted-foreground text-sm',
          },
          `Panel type "${type}" not found in registry`,
        );
      }

      if (process.env.NODE_ENV !== 'production') {
        console.log('[PanelRenderer] Render:', {
          type,
          propsId: props.id,
          propsTitle: props.title,
        });
      }

      // Plugin Context 주입 (외부 플러그인용)
      const pluginContext = getPluginContext();

      // Lazy 컴포넌트 래퍼 - loadComponent가 null을 반환할 수 있으므로 처리
      const LazyComponent = React.lazy(async () => {
        const component = await registry.loadComponent(type);
        if (!component) {
          throw new Error(`Failed to load panel component: ${type}`);
        }
        return { default: component };
      });

      // Plugin Context를 props에 추가
      const enhancedProps = {
        ...props,
        pluginContext,
      };

      return React.createElement(
        React.Suspense,
        { fallback: fallback || React.createElement('div', null, 'Loading...') },
        React.createElement(LazyComponent, enhancedProps),
      );
    };

    return React.createElement(PanelRenderer);
  }

  clear(): void {
    this.plugins.clear();
    this.lazyComponentCache.clear();
    this.componentCache.clear();
    this.editorConfigCache.clear();
  }
}

export const panelPluginRegistry = new PanelPluginRegistry();
