import React from 'react';
import {convertUnit, UnitType} from '../../../lib';
import {UniqueMetricValue} from './chartType';
import { ChartMetric } from '../timeseries';

const calculateMetrics = (data: ChartMetric[], key: string): UniqueMetricValue => {
  const values = data.map((d) => d[key]).filter((val) => typeof val === 'number') as number[];
  return {
    key: key,
    average: values.reduce((sum, val) => sum + val, 0) / values.length,
    last: values.length ? values[values.length - 1] : undefined,
    max: values.length ? Math.max(...values) : undefined,
    min: values.length ? Math.min(...values) : undefined,
    total: values.reduce((sum, val) => sum + val, 0),
  };
};

const DataNotExistMessage = (message?: string) => {
  return (
    <div
      className={
        'flex flex-1 justify-center items-center h-full min-h-[100px] w-full text-sm text-muted-foreground rounded-md'
      }
    >
      {message ? message : 'Chart data not exists.'}
    </div>
  );
};

// ChartTooltipContentOnlyLine
// - 이 컴포넌트는 차트 사용 시 툴팁을 비활성화하고 라인만 표시하기 위해 사용
//
// 사용 예시:
// - syncId를 사용하여 여러 차트를 동기화하는 경우,
//   cursor가 특정 차트에서만 툴팁을 표시하도록 하고 다른 차트는 툴팁을 비활성화 할때
//
// 참고:
// - React.Fragment(`<></>`)를 바로 사용할 경우,
//   hook.js:608 Invalid prop `cursor` supplied to `React.Fragment` 에러가 발생
//   이는 React.Fragment가 key와 children 외의 props를 받을 수 없기 때문
//   함수형 컴포넌트로 정의하면 props를 받아도 무시하므로 이러한 에러가 발생하지 않음
const ChartTooltipContentOnlyLine = () => {
  return null;
};

interface unitProps {
  value: number | null; // 변환할 값 (KB 단위)
  unitKey: UnitType;
  unitPrefix?: string;
}

const CalculateUnit: React.FC<unitProps> = ({value, unitKey, unitPrefix}) => {
  if (value === null)
    return (
      <React.Fragment>
        -<span className="text-xs text-muted-foreground"> </span>
      </React.Fragment>
    );

  const calculate = convertUnit(value, unitKey);
  return (
    <>
      {typeof calculate.quantity === 'number' ? calculate.quantity.toFixed(2) : calculate.quantity}{' '}
      <span className="text-xs text-muted-foreground ml-1">
        {calculate.unit + (unitPrefix ? unitPrefix : '')}
      </span>
    </>
  );
};

export {calculateMetrics, DataNotExistMessage, ChartTooltipContentOnlyLine, CalculateUnit};
