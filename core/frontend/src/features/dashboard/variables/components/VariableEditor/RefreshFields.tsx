'use client';

import React from 'react';
import {Label} from '@pharos/shared/components/ui';
import {Checkbox} from '@pharos/shared/components/ui';
import type {FilterConfig} from '@pharos/shared/types/dashboard';
import {REFRESH_INTERVAL, type RefreshIntervalType} from '@features/dashboard/variables/plugins/refresh/types';

export interface RefreshFieldsProps {
  data: FilterConfig;
  onChange: React.Dispatch<React.SetStateAction<FilterConfig>>;
}

/**
 * Refresh Variable 설정 필드
 * 사용 가능한 refresh interval 옵션을 체크박스로 선택
 */
export function RefreshFields({data, onChange}: RefreshFieldsProps) {
  const handleCheckboxChange = (optionValue: RefreshIntervalType, checked: boolean) => {
    const currentValues = (data.options?.refreshOptions as RefreshIntervalType[]) || [];
    let newValues: RefreshIntervalType[];

    if (checked) {
      newValues = [...currentValues, optionValue];
    } else {
      newValues = currentValues.filter((value) => value !== optionValue);
    }

    onChange((prev) => ({
      ...prev,
      options: {
        ...prev.options,
        refreshOptions: newValues,
      },
    }));
  };

  const currentValues = (data.options?.refreshOptions as RefreshIntervalType[]) || [];

  return (
    <div className="space-y-4 mt-4">
      <div>
        <Label className="text-sm font-normal">Available Refresh Intervals</Label>
        <p className="text-xs text-muted-foreground mt-1">
          Select which refresh interval options will be available in the dropdown
        </p>
        <div className="mt-3 grid grid-cols-3 gap-2 max-w-3xl">
          {REFRESH_INTERVAL.map((option) => (
            <div key={option} className="flex items-center space-x-2">
              <Checkbox
                id={`refresh-${option}`}
                checked={currentValues.includes(option)}
                onCheckedChange={(checked) =>
                  handleCheckboxChange(option, checked === true)
                }
              />
              <Label
                htmlFor={`refresh-${option}`}
                className="text-sm font-normal cursor-pointer"
              >
                {option}
              </Label>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
