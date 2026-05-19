/**
 * Alert Rule Editor - Main component with type selector
 */

'use client';

import React from 'react';
import {
  EditorSelectTemplate,
  Option,
  Options,
} from '@pharos/shared/components/template/editor-select';
import {QueryAlertRuleEditor} from '@features/alert';
import {EventStatusAlertRuleEditor} from '@features/alert';
import {EventHistoryAlertRuleEditor} from '@features/alert';
import type {QueryAlertRule, EventStatusAlertRule, EventHistoryAlertRule} from '@pharos/shared/types/alert';

interface AlertRuleEditorProps {
  initialType?: 'query' | 'event-status' | 'event-history';
  initialData?: Partial<QueryAlertRule | EventStatusAlertRule | EventHistoryAlertRule>;
  onSubmit?: (data: any) => void;
  onCancel?: () => void;
  isLoading?: boolean;
}

export function AlertRuleEditor({
  initialType = 'query',
  initialData,
  onSubmit,
  onCancel,
  isLoading,
}: AlertRuleEditorProps) {
  const handleSubmit = React.useCallback((data: any) => {
    onSubmit?.(data);
  }, [onSubmit]);

  const handleCancel = React.useCallback(() => {
    onCancel?.();
  }, [onCancel]);

  return (
    <EditorSelectTemplate
      selectLabel="Alert Type"
      selectPlaceholder="Select alert type"
      defaultValue={initialType}
    >
      <Options>
          <Option value="query" label="Query Alert">
            <QueryAlertRuleEditor
              initialData={initialData as Partial<QueryAlertRule>}
              onSubmit={handleSubmit}
              onCancel={handleCancel}
              isLoading={isLoading}
            />
          </Option>
          <Option value="event-status" label="Event Status Alert">
            <EventStatusAlertRuleEditor
              initialData={initialData as Partial<EventStatusAlertRule>}
              onSubmit={handleSubmit}
              onCancel={handleCancel}
              isLoading={isLoading}
            />
          </Option>
          <Option value="event-history" label="Event History Alert">
            <EventHistoryAlertRuleEditor
              initialData={initialData as Partial<EventHistoryAlertRule>}
              onSubmit={handleSubmit}
              onCancel={handleCancel}
              isLoading={isLoading}
            />
        </Option>
      </Options>
    </EditorSelectTemplate>
  );
}