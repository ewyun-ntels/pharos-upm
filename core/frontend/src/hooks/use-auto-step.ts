import {useEffect, useState} from 'react';
import type {DateTimeRangeValue} from '@pharos/shared/components/ui-extension';
import {getAbsoluteValueTimestamp} from '@pharos/shared/components/ui-extension';
import parse from 'parse-duration';

export type StepType = 'auto' | '1m' | '5m' | '15m' | '30m' | '1h';

export const STEP: StepType[] = ['auto', '1m', '5m', '15m', '30m', '1h'];

export const calculateAutoStep = (start: DateTimeRangeValue, end: DateTimeRangeValue): number => {
  const startTime = getAbsoluteValueTimestamp(start);
  const endTime = getAbsoluteValueTimestamp(end);
  const diffMinutes = (endTime - startTime) / 60;

  if (diffMinutes <= 30) {
    return 60; // 1분
  } else if (diffMinutes <= 120) {
    return 300;
  } else if (diffMinutes <= 300) {
    return 900;
  } else if (diffMinutes <= 720) {
    return 1800;
  } else {
    return 3600;
  }
};

export const useAutoStep = (
  startTime: DateTimeRangeValue,
  endTime: DateTimeRangeValue,
  initialStep: StepType = '5m',
) => {
  const [step, setStep] = useState<number>(() => {
    const parsed = parse(initialStep);
    return parsed ? parsed / 1000 : 300;
  });
  const [stepValue, setStepValue] = useState<StepType>(initialStep);

  const handleStepChange = (v: string) => {
    if (v === 'auto' || ['1m', '5m', '15m', '30m', '1h'].includes(v)) {
      if (v === 'auto') {
        const autoStep = calculateAutoStep(startTime, endTime);
        setStep(autoStep);
      } else {
        const parsedStep = parse(v);
        if (parsedStep) {
          setStep(parsedStep / 1000);
        }
      }
      setStepValue(v as StepType);
    }
  };

  useEffect(() => {
    if (stepValue === 'auto') {
      const autoStep = calculateAutoStep(startTime, endTime);
      setStep(autoStep);
    }
  }, [startTime, endTime, stepValue]);

  return {
    step,
    stepValue,
    handleStepChange,
    setStep,
    setStepValue,
  };
};
