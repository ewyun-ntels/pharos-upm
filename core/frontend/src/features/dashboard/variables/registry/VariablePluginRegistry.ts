import React from 'react';
import type { Action, FilterConfig, Kind } from '@pharos/shared/types/dashboard';
import type { FilterValue } from '../../types/filter.types';

/**
 * VariableProps: Variable 컴포넌트가 렌더링 시 받는 props
 *
 * PanelProps 패턴을 따라 설계
 * - FilterConfig (스키마) + 런타임 필드
 */
export interface VariableProps<T> {
  // FilterConfig 필드 (스키마에서 정의됨)
  id: string;
  type: string;
  kind: Kind;
  label?: string;
  datasourceName?: string;
  query?: string;

  // Plugin-specific options (타입 안전)
  options?: T;

  // 런타임 필드 (VariableItems에서 주입)
  dashboardId: string;
  index?: number;

  // 콜백
  onChange?: (value: FilterValue) => void;

  // 권한 (viewer → Panel 모드, editor/owner → Run 모드)
  permission: Action;
}

/**
 * Variable Plugin 정보
 */
export interface VariablePluginInfo {
  /** 변수 타입 (select, datetime, etc) */
  id: string;
  /** 사용자에게 표시되는 이름 */
  label: string;
  /** 아이콘 */
  icon?: React.ReactNode;
  /** 설명 */
  description?: string;
  /** 카테고리 */
  category?: string;
  /**
   * API 데이터 로드가 필요한지 여부
   * - true: select 등 API에서 options를 가져와야 하는 변수
   * - false: datetime/step/refresh 등 즉시 사용 가능한 변수
   * @default false
   */
  requiresData?: boolean;
}

/**
 * Variable Editor Props
 * Variable Settings 에디터에서 사용
 *
 * @template T - Plugin-specific options 타입
 */
export interface VariableEditorProps<T> {
  /** 편집 중인 FilterConfig 데이터 (Partial이므로 각 editor가 기본값 처리) */
  initialData?: Partial<FilterConfig & {
    options?: T;  // Plugin-specific options
  }>;

  /** 저장 콜백 */
  onSubmit: (data: FilterConfig) => void;

  /** 취소 콜백 */
  onCancel: () => void;

  /** 변경 콜백 (실시간 업데이트) */
  onChange?: (data: FilterConfig) => void;

  /** Dashboard context */
  dashboardId?: string;
}

/**
 * Variable Preview Props
 * Variable Settings 에디터에서 미리보기에 사용
 *
 * @template T - Plugin-specific options 타입
 */
export interface VariablePreviewProps<T> {
  /** 미리보기할 FilterConfig 데이터 */
  data: FilterConfig & {
    options?: T;
  };

  /** Dashboard context */
  dashboardId?: string;
}

/**
 * Variable Plugin
 *
 * @template T - Plugin-specific options 타입
 */
export interface VariablePlugin<T> {
  info: VariablePluginInfo;
  /** 변수 컴포넌트 - VariableProps를 받음 */
  component: React.ComponentType<VariableProps<T>>;
  /** 에디터 컴포넌트 - VariableEditorProps를 받음 (Optional) */
  editor?: React.ComponentType<VariableEditorProps<T>>;
  /** 미리보기 컴포넌트 - VariablePreviewProps를 받음 (Optional) */
  preview?: React.ComponentType<VariablePreviewProps<T>>;
}

/**
 * Variable Plugin Registry
 *
 * PanelPluginRegistry 패턴을 따라 설계
 */
class VariablePluginRegistry {
  private plugins = new Map<string, VariablePlugin<unknown>>();

  register<T>(plugin: VariablePlugin<T>): void {
    this.plugins.set(plugin.info.id, plugin as VariablePlugin<unknown>);
  }

  get(type: string): VariablePlugin<unknown> | undefined {
    return this.plugins.get(type);
  }

  getAll(): VariablePlugin<unknown>[] {
    return Array.from(this.plugins.values());
  }

  /**
   * 변수 타입의 미리보기를 가져옵니다.
   *
   * @param type - variable 타입
   * @returns 미리보기 컴포넌트 또는 undefined
   */
  getPreview(type: string): React.ComponentType<VariablePreviewProps<unknown>> | undefined {
    const plugin = this.get(type);
    return plugin?.preview;
  }

  /**
   * 특정 variable 타입이 API 데이터를 필요로 하는지 확인
   * @param type - variable 타입
   * @returns API 데이터 필요 여부 (기본값: false)
   */
  requiresData(type: string): boolean {
    const plugin = this.get(type);
    return plugin?.info.requiresData ?? false;
  }
}

export const variablePluginRegistry = new VariablePluginRegistry();
