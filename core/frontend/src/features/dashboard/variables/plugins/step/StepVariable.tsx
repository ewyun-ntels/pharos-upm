import React, {useEffect} from 'react';
import {VariablePlugin, variablePluginRegistry, VariableProps} from '@features/dashboard/variables';
import {ChartColumnIncreasing} from '@pharos/shared/components';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
  VariableSelect,
} from '@pharos/shared/components/ui-extension';
import {useDashboardStore} from '@features/dashboard/hooks/use-dashboard-store';
import {getStepValue} from '@features/dashboard/utils/datetime-step-utils';
import type {StepFilterOptions} from './types';
import {STEP, STEP_DEFAULT} from './types';
import {StepVariableEditor} from './StepVariableEditor';
import {StepVariablePreview} from './StepVariablePreview';
import {URL_PARAM_STEP} from '@features/dashboard/hooks/url';

/**
 * Step Variable Component
 * URL 파라미터: var_step (URL 동기화는 subscribe에서 자동 처리)
 */
export const StepVariable: React.FC<VariableProps<StepFilterOptions>> = (props) => {
  const {options} = props;

  const step = useDashboardStore((state) => state.filterState?.step);
  const stepOptions = useDashboardStore((state) => state.filterState?.stepOptions);
  const setStep = useDashboardStore((state) => state.setStep);
  const setRangeStepOptions = useDashboardStore((state) => state.setRangeStepOptions);

  // URL에서 초기값 읽기
  const urlStep = new URLSearchParams(window.location.search).get(URL_PARAM_STEP);

  // rangeStepOptions 등록 (config → runtime store)
  useEffect(() => {
    if (options?.rangeStepOptions) {
      setRangeStepOptions(options.rangeStepOptions);
    }
  }, [options?.rangeStepOptions, setRangeStepOptions]);

  // 초기값 설정 (URL > savedValue > initValue > STEP_DEFAULT)
  // _loadDashboard에서 resolveInitialState로 step이 이미 설정되어 있으면 skip.
  // savedValue: 이전에 "Save current step as dashboard default"로 저장된 값.
  useEffect(() => {
    if (!step) {
      const initialStep =
        urlStep ||
        options?.savedValue ||
        options?.initValue ||
        STEP_DEFAULT;
      const stepValue = getStepValue(initialStep);
      if (stepValue !== null) {
        setStep({ step: initialStep, stepValue });
      }
    }
  }, [step, urlStep, options, setStep]);

  const currentStepOptions = stepOptions && stepOptions.length > 0 ? stepOptions : STEP;

  // 값 변경 시 setStep만 호출 (URL 동기화는 subscribe에서 자동 처리)
  const handleChange = (v: string | string[]) => {
    const value = Array.isArray(v) ? v[0] : v;
    const stepValue = getStepValue(value);
    if (stepValue !== null) {
      setStep({ step: value, stepValue });
    }
  };

  return (
    <div className="flex items-center h-8 text-foreground border border-input rounded-md focus-within:border-ring focus-within:ring-ring/50 focus-within:ring-[3px] transition-[color,box-shadow] bg-background">
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <div className="px-2 flex items-center border-r border-input">
              <ChartColumnIncreasing className="w-3.5 h-3.5" />
            </div>
          </TooltipTrigger>
          <TooltipContent variant="icon" side="bottom">
            <p>Step</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
      <VariableSelect
        options={currentStepOptions?.map((s) => ({label: s, value: s})) || []}
        value={step?.step || options?.initValue || STEP_DEFAULT}
        onChange={handleChange}
        searchable={false}
        showLabel={false}
      />
    </div>
  );
};

/**
 * Step Variable Plugin
 */
export const stepVariablePlugin: VariablePlugin<StepFilterOptions> = {
  info: {
    id: 'step',
    label: 'Step',
    description: 'Time step selector',
    category: 'time',
    requiresData: false,
  },
  component: StepVariable,
  editor: StepVariableEditor,
  preview: StepVariablePreview,
};

// Register
variablePluginRegistry.register(stepVariablePlugin);
