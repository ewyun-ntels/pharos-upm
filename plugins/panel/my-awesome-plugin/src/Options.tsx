import React from 'react';
import {UseFormReturn, FieldValues} from 'react-hook-form';

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
 * My Awesome Panel의 옵션 타입
 */
interface MyAwesomePanelOptions {
  primaryColor?: string;
  secondaryColor?: string;
  size?: 'small' | 'medium' | 'large';
  showLabels?: boolean;
  showLegend?: boolean;
  animated?: boolean;
  title?: string;
  description?: string;
}

/**
 * Options 컴포넌트
 * 
 * Panel Editor에서 사용되는 옵션 UI입니다.
 * watch 대신 props로 직접 받아서 타입 안전!
 */
export const Options: React.FC<PanelEditorOptionsProps<MyAwesomePanelOptions>> = ({
  context,
  options = {},
  onOptionsChange,
}) => {
  const {register} = context;
  
  // props로 직접 받은 options 사용 - 타입 안전!
  const primaryColor = options.primaryColor || '#3b82f6';
  const secondaryColor = options.secondaryColor || '#8b5cf6';
  const size = options.size || 'medium';
  const showLabels = options.showLabels ?? true;
  const showLegend = options.showLegend ?? true;
  const animated = options.animated ?? true;
  const titleValue = options.title || '';
  const descriptionValue = options.description || '';

  return (
    <div className="space-y-4 p-4">
      <h3 className="text-lg font-semibold">My Awesome Panel Options</h3>
      
      {/* 색상 설정 */}
      <div className="space-y-2">
        <h4 className="text-sm font-medium">Colors</h4>
        
        <div className="flex gap-4">
          <div className="flex-1">
            <label className="block text-xs mb-1">Primary Color</label>
            <div className="flex gap-2">
              <input
                type="color"
                {...register?.('myAwesomeOptions.primaryColor')}
                className="w-12 h-8 rounded cursor-pointer"
              />
              <input
                type="text"
                value={primaryColor}
                {...register?.('myAwesomeOptions.primaryColor')}
                className="flex-1 px-2 py-1 text-sm border rounded"
                placeholder="#3b82f6"
              />
            </div>
          </div>
          
          <div className="flex-1">
            <label className="block text-xs mb-1">Secondary Color</label>
            <div className="flex gap-2">
              <input
                type="color"
                {...register?.('myAwesomeOptions.secondaryColor')}
                className="w-12 h-8 rounded cursor-pointer"
              />
              <input
                type="text"
                value={secondaryColor}
                {...register?.('myAwesomeOptions.secondaryColor')}
                className="flex-1 px-2 py-1 text-sm border rounded"
                placeholder="#8b5cf6"
              />
            </div>
          </div>
        </div>
      </div>
      
      {/* 크기 설정 */}
      <div>
        <label className="block text-sm font-medium mb-2">Size</label>
        <select
          {...register?.('myAwesomeOptions.size')}
          className="w-full px-3 py-2 border rounded"
          value={size}
        >
          <option value="small">Small</option>
          <option value="medium">Medium</option>
          <option value="large">Large</option>
        </select>
      </div>
      
      {/* 체크박스 옵션들 */}
      <div className="space-y-2">
        <label className="flex items-center gap-2">
          <input
            type="checkbox"
            {...register?.('myAwesomeOptions.showLabels')}
            className="rounded"
            checked={showLabels}
          />
          <span className="text-sm">Show Labels</span>
        </label>
        
        <label className="flex items-center gap-2">
          <input
            type="checkbox"
            {...register?.('myAwesomeOptions.showLegend')}
            className="rounded"
            checked={showLegend}
          />
          <span className="text-sm">Show Legend</span>
        </label>
        
        <label className="flex items-center gap-2">
          <input
            type="checkbox"
            {...register?.('myAwesomeOptions.animated')}
            className="rounded"
            checked={animated}
          />
          <span className="text-sm">Enable Animation</span>
        </label>
      </div>
      
      {/* 텍스트 입력 */}
      <div className="space-y-2">
        <div>
          <label className="block text-sm font-medium mb-1">Title</label>
          <input
            type="text"
            {...register?.('myAwesomeOptions.title')}
            className="w-full px-3 py-2 border rounded"
            placeholder="My Awesome Panel"
          />
        </div>
        
        <div>
          <label className="block text-sm font-medium mb-1">Description</label>
          <textarea
            {...register?.('myAwesomeOptions.description')}
            className="w-full px-3 py-2 border rounded"
            rows={3}
            placeholder="This is an example external plugin"
          />
        </div>
      </div>
      
      {/* 미리보기 */}
      <div className="mt-6 p-4 border rounded bg-gray-50">
        <h4 className="text-sm font-medium mb-2">Preview</h4>
        <div
          className="h-24 rounded flex items-center justify-center text-white font-semibold"
          style={{
            background: `linear-gradient(135deg, ${primaryColor}, ${secondaryColor})`,
          }}
        >
          {titleValue || 'My Awesome Panel'}
        </div>
      </div>
    </div>
  );
};

export default Options;
