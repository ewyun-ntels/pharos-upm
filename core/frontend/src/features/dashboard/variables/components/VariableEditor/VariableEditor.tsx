/**
 * Variable Editor - Plugin Registry 기반
 * 
 * Variable Plugin Registry에서 에디터를 동적으로 로드
 */

'use client';

import React from 'react';
import {
  EditorSelectTemplate,
  Option,
  Options,
} from '@pharos/shared/components/template/editor-select';
import {variablePluginRegistry} from '@features/dashboard/variables';
import type {FilterConfig} from '@pharos/shared/types/dashboard';

interface VariableEditorProps {
  initialType?: string;
  initialData?: Partial<FilterConfig>;
  onSubmit: (data: FilterConfig) => void;
  onCancel: () => void;
  onChange?: (data: FilterConfig) => void;
  dashboardId?: string;
}

export function VariableEditor({
  initialType = 'select',
  initialData,
  onSubmit,
  onCancel,
  onChange,
  dashboardId,
}: VariableEditorProps) {
  // Get all registered variable plugins
  const allPlugins = variablePluginRegistry.getAll();

  // Filter plugins that have editors
  const pluginsWithEditors = allPlugins.filter(plugin => plugin.editor);

  return (
    <EditorSelectTemplate
      title="Variable Type"
      // selectLabel="Variable Type"
      selectPlaceholder="Select variable type"
      defaultValue={initialType}
    >
      <Options>
        {pluginsWithEditors.map((plugin) => {
          const Editor = plugin.editor!;
          
          return (
            <Option key={plugin.info.id} value={plugin.info.id} label={plugin.info.label}>
              <Editor
                initialData={initialData}
                onSubmit={onSubmit}
                onCancel={onCancel}
                onChange={onChange}
                dashboardId={dashboardId}
              />
            </Option>
          );
        })}
      </Options>
    </EditorSelectTemplate>
  );
}
