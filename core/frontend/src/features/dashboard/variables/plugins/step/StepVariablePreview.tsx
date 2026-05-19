import React from 'react';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@pharos/shared/components/ui';
import type {VariablePreviewProps} from '../../registry/VariablePluginRegistry';
import type {StepFilterOptions} from './types';

export function StepVariablePreview({data}: VariablePreviewProps<StepFilterOptions>) {
  const daySteps = data.options?.rangeStepOptions?.['1day'] || [];
  const monthSteps = data.options?.rangeStepOptions?.['1month'] || [];
  const yearSteps = data.options?.rangeStepOptions?.['1year'] || [];

  return (
    <div className="p-4 space-y-4">
      <div className="text-sm text-muted-foreground text-center">
        Step options by time period
      </div>

      <div className="space-y-3">
        {daySteps.length > 0 && (
          <div className="flex items-center space-x-1">
            <div className="text-xs font-medium text-muted-foreground w-40">
              Day period (≤ 1 day):
            </div>
            <Select value={daySteps[0] || ''}>
              <SelectTrigger className="h-8 w-32">
                <SelectValue placeholder="Day steps" />
              </SelectTrigger>
              <SelectContent>
                {daySteps.map((step: string) => (
                  <SelectItem key={step} value={step}>
                    {step}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        )}

        {monthSteps.length > 0 && (
          <div className="flex items-center space-x-1">
            <div className="text-xs font-medium text-muted-foreground w-40">
              Month period (≤ 1 month):
            </div>
            <Select value={monthSteps[0] || ''}>
              <SelectTrigger className="h-8 w-32">
                <SelectValue placeholder="Month steps" />
              </SelectTrigger>
              <SelectContent>
                {monthSteps.map((step: string) => (
                  <SelectItem key={step} value={step}>
                    {step}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        )}

        {yearSteps.length > 0 && (
          <div className="flex items-center space-x-1">
            <div className="text-xs font-medium text-muted-foreground w-40">
              Year period (≤ 1 year):
            </div>
            <Select value={yearSteps[0] || ''}>
              <SelectTrigger className="h-8 w-32">
                <SelectValue placeholder="Year steps" />
              </SelectTrigger>
              <SelectContent>
                {yearSteps.map((step: string) => (
                  <SelectItem key={step} value={step}>
                    {step}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        )}

        {daySteps.length === 0 && monthSteps.length === 0 && yearSteps.length === 0 && (
          <div className="text-sm text-muted-foreground text-center">
            No step options configured
          </div>
        )}
      </div>
    </div>
  );
}
