/**
 * Select Variable Editor
 * Variable Settings에서 사용
 */

'use client';

import React, {useState} from 'react';
import type {VariableEditorProps} from '@features/dashboard/variables';
import type {SelectFilterOptions} from './types';
import type {FilterConfig} from '@pharos/shared/types/dashboard';
import {GeneralFields} from '@features/dashboard/variables/components/VariableEditor/GeneralFields';
import {QueryFields} from '@features/dashboard/variables/components/VariableEditor/QueryFields';
import {VariableEditorFooter} from '@features/dashboard/variables/components/VariableEditor/VariableEditorFooter';
import {Label, Checkbox} from '@pharos/shared/components/ui';

export const SelectVariableEditor: React.FC<VariableEditorProps<SelectFilterOptions>> = (props) => {
  const {
    initialData,
    onSubmit,
    onCancel,
    onChange,
    dashboardId,
  } = props;

  const [formData, setFormData] = useState<FilterConfig>(() => ({
    id: initialData?.id || '',
    type: 'select' as const,
    kind: initialData?.kind || 'headerL',
    datasourceName: initialData?.datasourceName || '',
    label: initialData?.label || '',
    query: initialData?.query || '',
    options: {
      showAllOption: (initialData?.options as SelectFilterOptions)?.showAllOption ?? false,
      defaultAllSelected: (initialData?.options as SelectFilterOptions)?.defaultAllSelected ?? false,
      customAllValue: (initialData?.options as SelectFilterOptions)?.customAllValue ?? '',
      isSearchable: (initialData?.options as SelectFilterOptions)?.isSearchable ?? false,
      isHidden: (initialData?.options as SelectFilterOptions)?.isHidden ?? false,
      isMulti: (initialData?.options as SelectFilterOptions)?.isMulti ?? false,
      showCheckbox: (initialData?.options as SelectFilterOptions)?.showCheckbox ?? false,
      autoSelectFirstOption: (initialData?.options as SelectFilterOptions)?.autoSelectFirstOption ?? true,
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

  const handleOptionChange = (field: keyof SelectFilterOptions, value: any) => {
    setFormData((prev) => ({
      ...prev,
      options: {
        ...(prev.options as SelectFilterOptions),
        [field]: value,
      },
    }));
  };

  return (
    <div className="space-y-7">
      <GeneralFields 
        data={formData} 
        onChange={setFormData} 
      />
      
      <QueryFields 
        data={formData} 
        onChange={setFormData} 
      />

      {/* Select-specific options */}
      <div className="space-y-4">
        <div className="flex items-center space-x-2">
          <Checkbox
            id="isMulti"
            checked={(formData.options as SelectFilterOptions)?.isMulti || false}
            onCheckedChange={(checked) => {
              handleOptionChange('isMulti', checked);
              // Auto-enable checkbox when multi-select is enabled
              if (checked) {
                handleOptionChange('showCheckbox', true);
              }
            }}
          />
          <Label htmlFor="isMulti" className="text-sm font-medium cursor-pointer">
            Enable multi-select
          </Label>
        </div>

        {(formData.options as SelectFilterOptions)?.isMulti && (
          <div className="flex items-center space-x-2 ml-6">
            <Checkbox
              id="showCheckbox"
              checked={(formData.options as SelectFilterOptions)?.showCheckbox || false}
              onCheckedChange={(checked) => handleOptionChange('showCheckbox', checked)}
            />
            <Label htmlFor="showCheckbox" className="text-sm font-normal cursor-pointer">
              Show checkbox in dropdown
            </Label>
          </div>
        )}

        <div className="flex items-center space-x-2">
          <Checkbox
            id="showAllOption"
            checked={(formData.options as SelectFilterOptions)?.showAllOption || false}
            onCheckedChange={(checked) => handleOptionChange('showAllOption', checked)}
          />
          <Label htmlFor="showAllOption" className="text-sm font-normal cursor-pointer">
            Show &#34;All&#34; option
          </Label>
        </div>

        {(formData.options as SelectFilterOptions)?.showAllOption && (
          <>
            <div className="flex items-center space-x-2 ml-6">
              <Checkbox
                id="defaultAllSelected"
                checked={(formData.options as SelectFilterOptions)?.defaultAllSelected || false}
                onCheckedChange={(checked) => handleOptionChange('defaultAllSelected', checked)}
              />
              <Label htmlFor="defaultAllSelected" className="text-sm font-normal cursor-pointer">
                Default to &#34;All&#34; selected
              </Label>
            </div>
            
            <div className="space-y-1 ml-6">
              <div className="gap-2">
                <Label htmlFor="customAllValue" className="text-sm font-normal">
                Custom &#34;All&#34; value (optional)
                </Label>
              </div>
              <p className="text-xs text-muted-foreground">
                When &#34;All&#34; is selected, use this value instead of all individual values. Useful for regex patterns like &#34;.*&#34;
              </p>
              <input
                id="customAllValue"
                type="text"
                placeholder="e.g., *, .*, or leave empty for all values"
                value={(formData.options as SelectFilterOptions)?.customAllValue || ''}
                onChange={(e) => handleOptionChange('customAllValue', e.target.value)}
                className="w-full px-3 py-1.5 text-sm border rounded-md"
                />
            </div>
          </>
        )}

        <div className="flex items-center space-x-2">
          <Checkbox
            id="isSearchable"
            checked={(formData.options as SelectFilterOptions)?.isSearchable || false}
            onCheckedChange={(checked) => handleOptionChange('isSearchable', checked)}
          />
          <Label htmlFor="isSearchable" className="text-sm font-normal cursor-pointer">
            Enable search
          </Label>
        </div>
        
        <div className="flex items-center space-x-2">
          <Checkbox
            id="autoSelectFirstOption"
            checked={(formData.options as SelectFilterOptions)?.autoSelectFirstOption ?? true}
            onCheckedChange={(checked) => handleOptionChange('autoSelectFirstOption', checked)}
          />
          <Label htmlFor="autoSelectFirstOption" className="text-sm font-normal cursor-pointer">
            Auto-select first option
          </Label>
        </div>

        <div className="space-y-1">
          <div className="flex items-center space-x-2">
            <Checkbox
              id="isHidden"
              checked={(formData.options as SelectFilterOptions)?.isHidden || false}
              onCheckedChange={(checked) => handleOptionChange('isHidden', checked)}
            />
            <Label htmlFor="isHidden" className="text-sm font-normal cursor-pointer">
              Hide
            </Label>
          </div>
          <p className="text-xs text-muted-foreground ml-6">
            Hide variable from UI but keep it available for queries.
          </p>
        </div>
      </div>

      <VariableEditorFooter
        previewData={formData}
        dashboardId={dashboardId}
        onCancel={onCancel}
        onSave={handleSubmit}
      />
    </div>
  );
};
