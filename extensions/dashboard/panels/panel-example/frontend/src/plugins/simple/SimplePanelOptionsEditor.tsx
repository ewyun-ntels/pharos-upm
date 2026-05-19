/**
 * Simple Panel Options Editor
 * 
 * 패널 편집기에서 표시되는 옵션 설정 폼
 */

import React from 'react';
import type { PanelEditorOptionsProps } from '@pharos/core/panel-registry';
import type { SimplePanelOptions } from './SimplePanel';

export function SimplePanelOptionsEditor(
  props: PanelEditorOptionsProps<SimplePanelOptions>
) {
  const { options, onOptionsChange } = props;

  const handleChange = (key: keyof SimplePanelOptions, newValue: string) => {
    onOptionsChange?.({
      ...options,
      [key]: newValue,
    });
  };

  return (
    <div style={{ padding: '16px', display: 'flex', flexDirection: 'column', gap: '16px' }}>
      <div>
        <label style={{ display: 'block', fontSize: '12px', fontWeight: '600', marginBottom: '8px' }}>
          Title
        </label>
        <input
          type="text"
          value={options?.title || ''}
          onChange={(e) => handleChange('title', e.target.value)}
          placeholder="Simple Panel"
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
          Background Color
        </label>
        <input
          type="color"
          value={options?.backgroundColor || '#f3f4f6'}
          onChange={(e) => handleChange('backgroundColor', e.target.value)}
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
        <label style={{ display: 'block', fontSize: '12px', fontWeight: '600', marginBottom: '8px' }}>
          Text Color
        </label>
        <input
          type="color"
          value={options?.textColor || '#1f2937'}
          onChange={(e) => handleChange('textColor', e.target.value)}
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
        <label style={{ display: 'block', fontSize: '12px', fontWeight: '600', marginBottom: '8px' }}>
          Message
        </label>
        <textarea
          value={options?.message || ''}
          onChange={(e) => handleChange('message', e.target.value)}
          placeholder="This is a simple custom panel example"
          rows={3}
          style={{
            width: '100%',
            padding: '8px',
            border: '1px solid #d1d5db',
            borderRadius: '4px',
            fontSize: '14px',
            resize: 'vertical',
          }}
        />
      </div>

      <div>
        <label style={{ display: 'block', fontSize: '12px', fontWeight: '600', marginBottom: '8px' }}>
          Icon (Emoji)
        </label>
        <input
          type="text"
          value={options?.icon || ''}
          onChange={(e) => handleChange('icon', e.target.value)}
          placeholder="🎨"
          maxLength={2}
          style={{
            width: '100%',
            padding: '8px',
            border: '1px solid #d1d5db',
            borderRadius: '4px',
            fontSize: '24px',
            textAlign: 'center',
          }}
        />
      </div>
    </div>
  );
}
