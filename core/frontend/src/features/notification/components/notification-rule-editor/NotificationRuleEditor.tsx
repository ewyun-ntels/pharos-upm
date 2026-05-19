/**
 * Notification Rule Editor - Main component with notification type selector
 *
 * 현재는 SNMP만 지원하지만, 추후 Slack, Email, Webhook 등 추가 가능
 */

'use client';

import React from 'react';
import {
  EditorSelectTemplate,
  Option,
  Options,
} from '@pharos/shared/components/template/editor-select';
import { SnmpRuleEditor } from '../snmp-rule-editor';
import type { QueryNotificationRule } from '@pharos/shared/types/notification';

interface NotificationRuleEditorProps {
  initialType?: 'snmp' | 'slack' | 'email' | 'webhook' | 'teams';
  initialData?: Partial<QueryNotificationRule>;
  onSubmit?: (data: any) => void;
  onCancel?: () => void;
  isLoading?: boolean;
}

export function NotificationRuleEditor({
  initialType = 'snmp',
  initialData,
  onSubmit,
  onCancel,
  isLoading,
}: NotificationRuleEditorProps) {
  const handleSubmit = React.useCallback((data: any) => {
    onSubmit?.(data);
  }, [onSubmit]);

  const handleCancel = React.useCallback(() => {
    onCancel?.();
  }, [onCancel]);

  // Determine initial SNMP version from data
  const initialSnmpVersion = initialData?.version === 3 ? 'v3' : 'v2';

  return (
    <EditorSelectTemplate
      selectLabel="Notification Type"
      selectPlaceholder="Select notification type"
      defaultValue={initialType}
    >
      <Options>
        <Option value="snmp" label="SNMP">
          <SnmpRuleEditor
            initialVersion={initialSnmpVersion}
            initialData={initialData}
            onSubmit={handleSubmit}
            onCancel={handleCancel}
            isLoading={isLoading}
          />
        </Option>
        {/* TODO: Add other notification types
        <Option value="slack" label="Slack">
          <SlackEditor ... />
        </Option>
        <Option value="email" label="Email">
          <EmailEditor ... />
        </Option>
        <Option value="webhook" label="Webhook">
          <WebhookEditor ... />
        </Option>
        <Option value="teams" label="Microsoft Teams">
          <TeamsEditor ... />
        </Option>
        */}
      </Options>
    </EditorSelectTemplate>
  );
}
