'use client';

import React from 'react';
import {Label} from '@pharos/shared/components/ui';
import {Checkbox} from '@pharos/shared/components/ui';
import type {FilterConfig} from '@pharos/shared/types/dashboard';

/**
 * Step 타입의 범위별 기본 스텝 옵션
 */
const STEP_OPTIONS = {
  day: [
    {value: '1m', label: '1m'},
    {value: '5m', label: '5m'},
    {value: '30m', label: '30m'},
    {value: '1h', label: '1h'},
  ],
  month: [
    {value: '1h', label: '1h'},
    {value: '6h', label: '6h'},
    {value: '12h', label: '12h'},
    {value: '1day', label: '1day'},
  ],
  year: [
    {value: '1day', label: '1day'},
    {value: '1month', label: '1month'},
  ],
};

export interface StepFieldsProps {
  data: FilterConfig;
  onChange: React.Dispatch<React.SetStateAction<FilterConfig>>;
}

export function StepFields({data, onChange}: StepFieldsProps) {
  const handleStepChange = (range: '1day' | '1month' | '1year', value: string[]) => {
    onChange((prev) => ({
      ...prev,
      options: {
        ...prev.options,
        rangeStepOptions: {
          ...prev.options?.rangeStepOptions,
          [range]: value,
        },
      },
    }));
  };

  const handleCheckboxChange = (
    range: '1day' | '1month' | '1year',
    optionValue: string,
    checked: boolean,
  ) => {
    const currentValues = data.options?.rangeStepOptions?.[range] || [];
    let newValues: string[];

    if (checked) {
      newValues = [...currentValues, optionValue];
    } else {
      newValues = currentValues.filter((value: string) => value !== optionValue);
    }

    handleStepChange(range, newValues);
  };

  const renderStepCheckboxes = (
    label: string,
    range: '1day' | '1month' | '1year',
    options: Array<{value: string; label: string}>,
  ) => {
    const currentValue = data.options?.rangeStepOptions?.[range] || [];

    return (
      <div>
        <Label className="text-sm font-normal">{label}</Label>
        <div className="mt-2 space-y-2">
          {options.map((option) => (
            <div key={option.value} className="flex items-center space-x-2">
              <Checkbox
                id={`${range}-${option.value}`}
                checked={currentValue.includes(option.value)}
                onCheckedChange={(checked) =>
                  handleCheckboxChange(range, option.value, checked === true)
                }
              />
              <Label
                htmlFor={`${range}-${option.value}`}
                className="text-sm font-normal cursor-pointer"
              >
                {option.label}
              </Label>
            </div>
          ))}
        </div>
      </div>
    );
  };

  return (
    <div className="space-y-4 mt-4">
      {renderStepCheckboxes('Day', '1day', STEP_OPTIONS.day)}
      {renderStepCheckboxes('Month', '1month', STEP_OPTIONS.month)}
      {renderStepCheckboxes('Year', '1year', STEP_OPTIONS.year)}
    </div>
  );
}
