/**
 * Plugin Configuration
 * 
 * 이 파일은 플러그인의 entry point입니다.
 * PanelPlugin 인터페이스를 구현하여 Registry에 등록됩니다.
 */

import React from 'react';
import {UseFormReturn, FieldValues} from 'react-hook-form';

/**
 * PanelPlugin 타입 정의
 * 
 * 📝 주의: 이 타입들은 @pharos/core/panel-registry의 PanelPluginRegistry.ts에 정의된
 * 공식 타입과 **완전히 동일**해야 합니다. 외부 플러그인이므로 독립적으로 정의합니다.
 * 
 * 공식 타입 위치: core/frontend/src/features/panels/plugins/registry/PanelPluginRegistry.ts
 * 버전: v2.1.1 기준
 */
interface PanelPluginInfo {
  label: string;
  value: string;
  icon?: React.ReactNode;
  description?: string;
  category?: string;
}

/**
 * Panel Editor Options Component Props
 */
interface PanelEditorOptionsProps<TOptions = any> {
  /** React Hook Form의 context (setValue 등 메서드용만 사용) */
  context: UseFormReturn<FieldValues, any, undefined>;
  /** 패널별 옵션 객체 (타입 안전) */
  options?: TOptions;
  /** 옵션 변경 핸들러 */
  onOptionsChange?: (options: Partial<TOptions>) => void;
  /** 데이터 프로바이더 (chartQuery 등) - props로 직접 전달! */
  dataProvider?: any;
  /** 패널 타입 */
  panelType?: string;
  /** 공통 패널 필드들 */
  title?: string;
  description?: string;
  bgTransparent?: boolean;
}

/**
 * Panel Editor 설정
 */
interface PanelEditorConfig {
  /**
   * 옵션 UI 컴포넌트 - UseFormReturn context를 받음
   */
  OptionsComponent: React.ComponentType<PanelEditorOptionsProps>;
  
  /**
   * Form 데이터를 Panel 데이터로 변환
   */
  toPanelData: (formData: any, ...args: any[]) => any | Promise<any>;
  
  /**
   * 기본값 설정 (선택사항)
   */
  getDefaults?: (data?: any) => any;
  
  /**
   * Panel 데이터를 Form 데이터로 역변환 (선택사항)
   */
  toFormData?: (panelData: any) => any;
}

interface PanelPlugin {
  info: PanelPluginInfo;
  component: () => Promise<React.ComponentType<any>>;
  editor?: () => Promise<PanelEditorConfig>;
}

/**
 * Plugin Icon
 * 간단한 SVG 아이콘 또는 이모지를 사용할 수 있습니다.
 */
const MyAwesomeIcon = () => {
  return React.createElement('span', {style: {fontSize: '1.2em'}}, '🎨');
};

/**
 * My Awesome Plugin 정의
 * 
 * 이 객체가 PanelPluginRegistry에 등록됩니다.
 */
export const myAwesomePlugin: PanelPlugin = {
  // 플러그인 메타데이터
  info: {
    label: 'My Awesome Panel',
    value: 'my-awesome-panel',
    icon: React.createElement(MyAwesomeIcon),
    description: 'An example external panel plugin with customizable colors and options',
    category: 'custom',
  },

  // 메인 컴포넌트 (lazy loading)
  component: async () => {
    const module = await import('./MyAwesomePanel');
    return module.MyAwesomePanel;
  },

  // Panel Editor 설정 (옵션 UI + 데이터 변환)
  editor: async () => {
    // 동적으로 Options와 setParam 로드
    const [optionsModule, setParamModule] = await Promise.all([
      import('./Options'),
      import('./setParam'),
    ]);

    return {
      // 옵션 UI 컴포넌트
      OptionsComponent: optionsModule.Options,
      
      // Form 데이터 → Panel 데이터 변환
      toPanelData: setParamModule.setParamToChart,
      
      // 기본값 설정
      getDefaults: setParamModule.setDefaultParam,
      
      // Panel 데이터 → Form 데이터 변환 (편집 시)
      toFormData: setParamModule.toFormData,
    };
  },
};

// Default export도 제공
export default myAwesomePlugin;
