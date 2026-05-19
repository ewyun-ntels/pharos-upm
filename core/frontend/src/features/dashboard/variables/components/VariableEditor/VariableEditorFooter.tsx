/**
 * Variable Editor Footer
 * Preview + Cancel/Save buttons
 */

'use client';

import React from 'react';
import {Button} from '@pharos/shared/components/ui';
import {VariablePreview} from './VariablePreview';
import type {FilterConfig} from '@pharos/shared/types/dashboard';

interface VariableEditorFooterProps {
  previewData: FilterConfig;
  dashboardId?: string;
  onCancel: () => void;
  onSave: () => void;
  saveDisabled?: boolean;
}

export function VariableEditorFooter({
  previewData,
  dashboardId,
  onCancel,
  onSave,
  saveDisabled = false,
}: VariableEditorFooterProps) {
  return (
    <div className="space-y-4 mt-8">
      {/* Preview Section */}
      <VariablePreview data={previewData} dashboardId={dashboardId} />

      {/* Action Buttons */}
      <div className="flex gap-2 pt-2">
        <Button variant="outline" onClick={onCancel}>
          Cancel
        </Button>
        <Button onClick={onSave} disabled={saveDisabled}>
          Save
        </Button>
      </div>
    </div>
  );
}
