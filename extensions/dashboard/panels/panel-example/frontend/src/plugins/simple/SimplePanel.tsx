/**
 * Simple Panel Example
 * 
 * 가장 기본적인 커스텀 패널 예시입니다.
 * - Props에서 options과 기본 정보를 받아 렌더링
 * - 데이터 소스 없이 동작하는 정적 패널
 * - UI만 커스터마이징하는 경우 사용
 */

import React from 'react';
import type { PanelProps } from '@pharos/core/panel-registry';

/**
 * 패널 옵션 타입 정의
 * 사용자가 패널 편집기에서 설정할 수 있는 값들
 */
export interface SimplePanelOptions {
  /** 패널 제목 */
  title?: string;
  /** 배경색 */
  backgroundColor?: string;
  /** 텍스트 색상 */
  textColor?: string;
  /** 메시지 */
  message?: string;
  /** 아이콘 (이모지) */
  icon?: string;
}

/**
 * 패널 컴포넌트
 */
export function SimplePanel(props: PanelProps<SimplePanelOptions>) {
  // options에서 기본값과 함께 설정 추출
  const {
    title = 'Simple Panel',
    backgroundColor = '#f3f4f6',
    textColor = '#1f2937',
    message = 'This is a simple custom panel example',
    icon = '🎨',
  } = props.options || {};

  return (
    <div
      style={{
        height: '100%',
        padding: '24px',
        backgroundColor,
        borderRadius: '8px',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        gap: '16px',
      }}
    >
      <div style={{ fontSize: '64px' }}>{icon}</div>
      
      <h3
        style={{
          fontSize: '24px',
          fontWeight: '600',
          color: textColor,
          margin: 0,
        }}
      >
        {props.title || title}
      </h3>
      
      <p
        style={{
          fontSize: '16px',
          color: textColor,
          opacity: 0.8,
          margin: 0,
          textAlign: 'center',
        }}
      >
        {message}
      </p>

      <div
        style={{
          marginTop: '24px',
          padding: '16px',
          backgroundColor: 'rgba(255, 255, 255, 0.5)',
          borderRadius: '8px',
          fontSize: '12px',
          color: textColor,
          opacity: 0.6,
        }}
      >
        <p style={{ margin: 0 }}>Panel ID: {props.id}</p>
        <p style={{ margin: 0 }}>Dashboard ID: {props.dashboardId}</p>
      </div>
    </div>
  );
}
