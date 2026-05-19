import React from 'react';
import type {PanelEditorOptionsProps} from '@pharos/core/panel-registry';
import {MarkdownViewerPanelOptions} from './types';
import {useOptionsChange} from '../common/useOptionsChange';

export const Options: React.FC<PanelEditorOptionsProps<MarkdownViewerPanelOptions>> = ({
  options = {},
  onOptionsChange,
}) => {
  const handleChange = useOptionsChange(options, onOptionsChange);

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium mb-1">Markdown Content</label>
        <textarea
          className="w-full h-64 px-3 py-2 border rounded-md"
          value={options.content || ''}
          onChange={(e) => handleChange(d => { d.content = e.target.value; })}
          placeholder="Enter markdown content..."
        />
      </div>

      <div>
        <label className="block text-sm font-medium mb-1">No Data Message</label>
        <input
          type="text"
          className="w-full px-3 py-2 border rounded-md"
          value={options.chartDataNotExistMessage || ''}
          onChange={(e) => handleChange(d => { d.chartDataNotExistMessage = e.target.value; })}
          placeholder="Custom message when data doesn't exist"
        />
      </div>
    </div>
  );
};
