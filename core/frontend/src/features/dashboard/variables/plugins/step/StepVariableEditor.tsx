/**
 * Step Variable Editor
 * Variable Settings에서 사용
 */

'use client';

import React, {useState} from 'react';
import type {VariableEditorProps} from '@features/dashboard/variables';
import type {StepFilterOptions} from './types';
import type {FilterConfig} from '@pharos/shared/types/dashboard';
import {GeneralFields} from '@features/dashboard/variables/components/VariableEditor/GeneralFields';
import {StepFields} from '@features/dashboard/variables/components/VariableEditor/StepFields';
import {VariableEditorFooter} from '@features/dashboard/variables/components/VariableEditor/VariableEditorFooter';

export const StepVariableEditor: React.FC<VariableEditorProps<StepFilterOptions>> = (props) => {
  const {
    initialData,
    onSubmit,
    onCancel,
    onChange,
    dashboardId,
  } = props;

  const [formData, setFormData] = useState<FilterConfig>(() => ({
    id: initialData?.id || '',
    type: 'step' as const,
    kind: initialData?.kind || 'headerL',
    options: {
      description: initialData?.options?.description || '',
      rangeStepOptions: {
        '1day': initialData?.options?.rangeStepOptions?.['1day'] || ['1m', '5m', '30m', '1h'],
        '1month': initialData?.options?.rangeStepOptions?.['1month'] || ['1h', '6h', '12h', '1day'],
        '1year': initialData?.options?.rangeStepOptions?.['1year'] || ['1day', '1month'],
      },
    },
  }));

  // Notify parent of changes
  React.useEffect(() => {
    onChange?.(formData);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [formData]); // onChange 제거하여 무한 루프 방지

  const handleSubmit = () => {
    onSubmit(formData);
  };

  return (
    <div className="space-y-7">
      <GeneralFields 
        data={formData} 
        onChange={setFormData} 
      />
      
      <StepFields 
        data={formData} 
        onChange={setFormData} 
      />

      <VariableEditorFooter
        previewData={formData}
        dashboardId={dashboardId}
        onCancel={onCancel}
        onSave={handleSubmit}
      />
    </div>
  );
};
