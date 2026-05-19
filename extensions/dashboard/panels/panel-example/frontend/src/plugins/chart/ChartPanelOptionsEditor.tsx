/**
 * Chart Panel Options Editor
 * 
 * 차트 패널의 옵션 설정 폼
 */

import React from 'react';
import type { PanelEditorOptionsProps } from '@pharos/core/panel-registry';
import type { ChartPanelOptions } from './ChartPanel';

export function ChartPanelOptionsEditor(
  props: PanelEditorOptionsProps<ChartPanelOptions>
) {
  const { options, onOptionsChange } = props;

  const handleChange = (key: keyof ChartPanelOptions, newValue: any) => {
    onOptionsChange?.({
      ...options,
      [key]: newValue,
    });
  };

  return (
    <div style={{ padding: '16px', display: 'flex', flexDirection: 'column', gap: '16px' }}>
      <div>
        <label style={{ display: 'block', fontSize: '12px', fontWeight: '600', marginBottom: '8px' }}>
          Chart Title
        </label>
        <input
          type="text"
          value={options?.title || ''}
          onChange={(e) => handleChange('title', e.target.value)}
          placeholder="Chart Panel Example"
          style={{
            width: '100%',
            padding: '8px',
            border: '1px solid #d1d5db',
            borderRadius: '4px',
            fontSize: '14px',
          }}
        />
      </div>

      <div>
        <label style={{ display: 'block', fontSize: '12px', fontWeight: '600', marginBottom: '8px' }}>
          Chart Color
        </label>
        <input
          type="color"
          value={options?.color || '#3b82f6'}
          onChange={(e) => handleChange('color', e.target.value)}
          style={{
            width: '100%',
            height: '40px',
            border: '1px solid #d1d5db',
            borderRadius: '4px',
            cursor: 'pointer',
          }}
        />
      </div>

      <div>
        <label style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer' }}>
          <input
            type="checkbox"
            checked={options?.showLegend ?? true}
            onChange={(e) => handleChange('showLegend', e.target.checked)}
            style={{ width: '16px', height: '16px', cursor: 'pointer' }}
          />
          <span style={{ fontSize: '12px', fontWeight: '600' }}>Show Legend</span>
        </label>
      </div>

      <div>
        <label style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer' }}>
          <input
            type="checkbox"
            checked={options?.showDataPoints ?? false}
            onChange={(e) => handleChange('showDataPoints', e.target.checked)}
            style={{ width: '16px', height: '16px', cursor: 'pointer' }}
          />
          <span style={{ fontSize: '12px', fontWeight: '600' }}>Show Data Points</span>
        </label>
      </div>

      <div
        style={{
          marginTop: '16px',
          padding: '12px',
          backgroundColor: '#f3f4f6',
          borderRadius: '6px',
          fontSize: '12px',
          color: '#6b7280',
        }}
      >
        <p style={{ margin: '0 0 8px 0', fontWeight: '600' }}>💡 Tip</p>
        <p style={{ margin: 0 }}>
          이 패널은 데이터 소스 탭에서 설정한 쿼리를 통해 데이터를 로드합니다.
          useDashboardQuery hook을 사용하는 방법을 확인하세요.
        </p>
      </div>
    </div>
  );
}
