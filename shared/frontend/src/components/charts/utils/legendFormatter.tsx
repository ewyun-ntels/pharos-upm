import React from 'react';
import {LegendPayload as Payload} from 'recharts';
import {Minus, Square} from '@pharos/shared/components';
import {cn} from '../../../lib';
import {UniqueMetric, ChartMetricLabel} from './chartType';
import {TextWithTooltip} from '@pharos/shared/components/ui-extension';

interface PromLegendProps {
  payload: Payload[] | undefined;
  className?: string;
  metrics: UniqueMetric;
  showMin: boolean;
  showMax: boolean;
  showAverage: boolean;
  showLast: boolean;
  showTotal: boolean;
  unitFn?: (value: number | undefined) => string;
  selectedLegend?: string[];
  selectedLegendFn?: (key: string | undefined) => void;
  isPieChart?: boolean;
}

const PromLegend: React.FC<PromLegendProps> = ({
  payload,
  className,
  metrics,
  showMin,
  showMax,
  showAverage,
  showLast,
  showTotal,
  unitFn,
  selectedLegend,
  selectedLegendFn,
  isPieChart = false,
}) => {
  return (
    <div
      // overflow scroll의 경우 오른쪽하단에 스크롤 바가 생기는데 이를 없애기 위해 overflow-auto로 변경
      // right시 middle정렬을 위해 [div > div] 2단구조로 변경후 chartType.ts에서 가로,세로 center 정렬 각각 지정됨
      className={cn('grid place-items-center overflow-auto', className)}
    >
      <div className="flex">
        {payload?.map((entry, index) => {
          const key = Object.keys(metrics)[index];
          const metric = metrics[key];
          const isSelected = selectedLegend?.includes(key);
          const hasSelection = Boolean(selectedLegend?.length);

          return (
            <div
              className={cn(
                'whitespace-nowrap font-regular transition-opacity rounded-[2px] hover:cursor-pointer hover:bg-muted/50',
                hasSelection && !isSelected && 'opacity-50',
              )}
              key={key}
              onClick={() => {
                if (selectedLegendFn) {
                  if (selectedLegend?.includes(key)) {
                    selectedLegendFn(undefined);
                  } else {
                    selectedLegendFn(key);
                  }
                }
              }}
            >
              <div
                className={'inline-flex items-center h-5 p-1 gap-1'}
                style={{color: entry.color}}
              >
                <Minus
                  fill={entry.color}
                  className={cn('h-4 w-4 shrink-0 **:stroke-[2px]!', isPieChart && 'hidden')}
                />
                {isPieChart && (
                  <Square fill={entry.color} className={'h-4 w-4 p-1 pr-0 rounded-sm shrink-0'} />
                )}
                {/* isPieChart 또는 isBarChart true일 때만 squre 적용 */}
                <TextWithTooltip text={entry.value ?? ''} className="truncate max-w-52 text-foreground" />
                {showMin && ` (Min: ${unitFn ? unitFn(metric.min) : metric.min})`}
                {showMax && ` (Max: ${unitFn ? unitFn(metric.max) : metric.max})`}
                {showAverage && ` (Avg: ${unitFn ? unitFn(metric.average) : metric.average})`}
                {showLast && ` (Last: ${unitFn ? unitFn(metric.last) : metric.last})`}
                {showTotal && ` (Total: ${unitFn ? unitFn(metric.total) : metric.total})`}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

const formatLegendLabel = (key: string, legendRule: string | undefined): string => {
  // JSON 형식이 아니면 바로 반환
  if (!key || key[0] !== '{') return key;
  try {
    const decodeKey = JSON.parse(key) as ChartMetricLabel;
    if (typeof decodeKey.metric === 'string') {
      return decodeKey.metric;
    }

    let formattedLabel = legendRule;
    if (formattedLabel) {
      for (const [k, v] of Object.entries(decodeKey.metric)) {
        const escapedKey = k.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
        const regex = new RegExp(`\\{${escapedKey}\\}`, 'g');
        formattedLabel = formattedLabel.replace(regex, v as string);
      }

      // custom label 처리
      const regex = new RegExp('\\{_label}', 'g');
      formattedLabel = formattedLabel.replace(regex, decodeKey._label);

      // 치환 후 아직 {placeholder} 패턴이 남아있으면 legendRule 매칭 실패 → 기본 포맷으로 폴백
      if (/\{[^}]+}/.test(formattedLabel)) {
        const metricLabels = Object.entries(decodeKey.metric)
          .map(([k, v]) => `${k}=${v}`)
          .join(', ');
        return metricLabels || decodeKey._label || key;
      }

      return formattedLabel;
    }

    // legendRule이 없을 때 기본 포맷 적용
    // metric 객체의 모든 key=value를 쉼표로 구분하여 표시
    const metricLabels = Object.entries(decodeKey.metric)
      .map(([k, v]) => `${k}=${v}`)
      .join(', ');

    return metricLabels || decodeKey._label || key;
  } catch (error) {
    console.error('Failed to parse key as JSON:', key, error);
    return key;
  }
};

export {PromLegend, formatLegendLabel};
export type {PromLegendProps};
