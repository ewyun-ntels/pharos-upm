import React, {useState, useEffect} from 'react';
import {VariableEditorProps} from '../../index';
import type {DateTimeFilterOptions} from './types';
import type {FilterConfig} from '@pharos/shared/types/dashboard';
import {GeneralFields} from '@features/dashboard/variables/components/VariableEditor/GeneralFields';
import {VariableEditorFooter} from '@features/dashboard/variables/components/VariableEditor/VariableEditorFooter';

/**
 * Datetime Variable Editor
 * - GeneralFields: id 입력
 * - Footer: Preview + Cancel/Save 버튼
 */
export const DatetimeVariableEditor: React.FC<VariableEditorProps<DateTimeFilterOptions>> = ({
  initialData,
  onSubmit,
  onCancel,
  onChange,
  dashboardId,
}) => {
  const [formData, setFormData] = useState<FilterConfig>(() => ({
    id: initialData?.id || '',
    type: 'datetime' as const,
    kind: initialData?.kind || 'headerL',
    options: initialData?.options || {},
  }));

  // 🔄 state 변화를 상위로 전파
  useEffect(() => {
    onChange?.(formData);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [formData]); // onChange 제거하여 무한 루프 방지

  const handleSubmit = () => {
    onSubmit(formData);
  };

  return (
    <div className="space-y-7">
      {/* ID 입력 필드 */}
      <GeneralFields data={formData} onChange={setFormData} />

      {/* Preview + 버튼 */}
      <VariableEditorFooter
        previewData={formData}
        dashboardId={dashboardId}
        onCancel={onCancel}
        onSave={handleSubmit}
      />
    </div>
  );
};
