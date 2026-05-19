import React from 'react';
import {MyAwesomePanelProps} from './types';

/**
 * My Awesome Panel - 메인 컴포넌트
 * 
 * 이 컴포넌트는 실제 데이터를 시각화합니다.
 * 여기서는 간단한 예제로 카드 스타일의 정보 표시를 보여줍니다.
 */
export const MyAwesomePanel: React.FC<MyAwesomePanelProps> = ({
  options = {},
  query,
  args = {},
  id,
  dashboardId,
}) => {
  const {
    primaryColor = '#3b82f6',
    secondaryColor = '#8b5cf6',
    size = 'medium',
    showLabels = true,
    showLegend = true,
    animated = true,
    title = 'My Awesome Panel',
    description = 'This is an example external plugin',
  } = options;

  // 크기에 따른 스타일
  const sizeStyles = {
    small: 'h-32 text-sm',
    medium: 'h-48 text-base',
    large: 'h-64 text-lg',
  };

  // 애니메이션 클래스
  const animationClass = animated ? 'transition-all duration-300 hover:scale-105' : '';

  return (
    <div className="w-full h-full p-4 flex items-center justify-center">
      <div
        className={`
          ${sizeStyles[size]}
          ${animationClass}
          w-full max-w-2xl
          rounded-lg shadow-lg
          overflow-hidden
        `}
        style={{
          background: `linear-gradient(135deg, ${primaryColor}, ${secondaryColor})`,
        }}
      >
        <div className="h-full flex flex-col justify-center items-center text-white p-6">
          {/* 타이틀 */}
          {showLabels && (
            <h2 className="text-2xl font-bold mb-2">{title}</h2>
          )}
          
          {/* 설명 */}
          {showLabels && description && (
            <p className="text-sm opacity-90 mb-4">{description}</p>
          )}
          
          {/* 플러그인 정보 */}
          <div className="mt-4 space-y-2 text-center">
            <div className="text-xs opacity-75">
              <strong>Plugin Type:</strong> External Panel Plugin
            </div>
            {query?.query && (
              <div className="text-xs opacity-75">
                <strong>Query:</strong> {query.query.substring(0, 50)}...
              </div>
            )}
            {id && (
              <div className="text-xs opacity-75">
                <strong>Panel ID:</strong> {id}
              </div>
            )}
            {dashboardId && (
              <div className="text-xs opacity-75">
                <strong>Dashboard ID:</strong> {dashboardId}
              </div>
            )}
          </div>
          
          {/* 범례 */}
          {showLegend && (
            <div className="mt-4 flex gap-4 text-xs">
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full" style={{backgroundColor: primaryColor}} />
                <span>Primary</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full" style={{backgroundColor: secondaryColor}} />
                <span>Secondary</span>
              </div>
            </div>
          )}
          
          {/* 데모 아이콘 */}
          <div className="mt-6 text-4xl">
            🎨
          </div>
        </div>
      </div>
    </div>
  );
};

export default MyAwesomePanel;
