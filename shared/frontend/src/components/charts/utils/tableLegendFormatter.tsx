import React, {useEffect, useRef} from 'react';
import {LegendPayload as Payload} from 'recharts';
import {ArrowUpRight, Minus, Square} from '@pharos/shared/components';
import {cn} from '../../../lib';

import {create} from 'zustand';
import {TextWithTooltip} from '@pharos/shared/components/ui-extension';

export type ScrollStore = {
  scrollPositions: Record<string, number>;
  setScrollPosition: (id: string, scrollTop: number) => void;
  removeScrollPosition: (id: string) => void;
};

export const useScrollStore = create<ScrollStore>((set) => ({
  scrollPositions: {},
  setScrollPosition: (id: string, scrollTop: number) =>
    set((state) => ({
      scrollPositions: {
        ...state.scrollPositions,
        [id]: scrollTop,
      },
    })),
  removeScrollPosition: (id: string) =>
    set((state) => {
      const newScrollPositions = {...state.scrollPositions};
      delete newScrollPositions[id];
      return {scrollPositions: newScrollPositions};
    }),
}));

export default useScrollStore;

interface PromLegendTableProps {
  id: string; // 고유한 ID를 추가합니다.
  payload: Payload[] | undefined;
  className?: string;
  metrics: Record<
    string,
    {
      min: number | undefined;
      max: number | undefined;
      average: number | undefined;
      last: number | undefined;
      total: number | undefined;
    }
  >;
  showMin: boolean;
  showMax: boolean;
  showAverage: boolean;
  showLast: boolean;
  showTotal: boolean;
  unitFn?: (value: number | undefined) => string;
  selectedLegend?: string[];
  selectedLegendFn?: (key: string | undefined) => void;
  isPieChart?: boolean;
  isBarChart?: boolean;
  showLink?: boolean;
}

const PromLegendTable: React.FC<PromLegendTableProps> = ({
  id,
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
  isBarChart = false,
  showLink = false,
}) => {
  const scrollTop = useScrollStore((state) => state.scrollPositions[id] || 0);
  const setScrollPosition = useScrollStore((state) => state.setScrollPosition);

  useEffect(() => {
    if (legendRef.current) {
      if (legendRef.current.scrollTop !== scrollTop) {
        legendRef.current.scrollTop = scrollTop;
      }
    }
  }, [scrollTop]);

  useEffect(() => {
    const handleScroll = () => {
      if (legendRef.current) {
        if (legendRef.current.scrollTop != scrollTop) {
          setScrollPosition(id, legendRef.current.scrollTop);
        }
      }
    };

    const currentRef = legendRef.current;
    if (currentRef) {
      currentRef.addEventListener('scroll', handleScroll);
    }
  }, [id, scrollTop, setScrollPosition]);

  const legendRef = useRef<HTMLDivElement>(null);

    return (
      <div ref={legendRef} className={cn(className, 'overflow-x-auto h-full')}>
      <table className={'w-full'}>
        <thead className="sticky top-0 h-4 z-[1] border-b">
          <tr>
            <th className={'p-2 py-0 text-left bg-background'}></th>
            {showMin && (
              <th
                className={
                  'p-2 py-0 text-right text-[11px] font-normal text-muted-foreground/70 bg-background '
                }
              >
                Min
              </th>
            )}
            {showMax && (
              <th
                className={
                  'p-2 py-0 text-right text-[11px] font-normal text-muted-foreground/70 bg-background '
                }
              >
                Max
              </th>
            )}
            {showAverage && (
              <th
                className={
                  'p-2 py-0 text-right text-[11px] font-normal text-muted-foreground/70 bg-background '
                }
              >
                Avg
              </th>
            )}
            {showLast && (
              <th
                className={
                  'p-2 py-0 text-right text-[11px] font-normal text-muted-foreground/70 bg-background '
                }
              >
                Last
              </th>
            )}
            {showTotal && (
              <th
                className={
                  'p-2 py-0 text-right text-[11px] font-normal text-muted-foreground/70 bg-background '
                }
              >
                Total
              </th>
            )}
          </tr>
        </thead>
        <tbody>
          {payload?.map((entry, index) => {
            const key = Object.keys(metrics)[index];
            const metric = metrics[key];
            return (
              <tr
                className={cn(
                  '[&_td]:h-[27px] border-b last:border-b-0 hover:bg-muted/50 hover:cursor-pointer',
                  selectedLegend !== undefined && !selectedLegend.includes(key) && 'opacity-50',
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
                <td
                  style={{color: entry.color}}
                  className={cn(
                    'flex gap-1 items-center text-xs truncate',
                    isPieChart && 'xl:max-w-36',
                  )}
                >
                  <div ref={legendRef} className="flex items-center gap-1 ml-2">
                    <Minus
                      className={cn(
                        'h-4 w-4 shrink-0 [&_*]:!stroke-[2px]',
                        (isPieChart || isBarChart) && 'hidden',
                      )}
                    />
                    {(isPieChart || isBarChart) && (
                      <Square
                        fill={entry.color}
                        className={'h-4 w-4 p-1 pr-0 rounded-sm shrink-0'}
                      />
                    )}
                    {/* isPieChart 또는 isBarChart true일 때만 squre 적용 */}
                    <span
                      style={{maxWidth: '300px', minWidth: '100px'}}
                      className={cn(
                        'text-foreground hover:cursor-pointer font-regular truncate',
                        isPieChart && 'hover:cursor-default',
                        selectedLegend?.length !== undefined,
                      )}
                    >
                      <TextWithTooltip text={entry.value ?? ''} />
                    </span>
                  </div>
                  {showLink && (
                    <a
                      href={`${showLink}/${key}`}
                      className={cn(
                        'h-4 w-4 py-[1px] px-1 shrink-0 flex items-center justify-center hover:border hover:cursor-pointer',
                        (isPieChart || isBarChart) && 'hidden',
                      )}
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      <ArrowUpRight stroke={entry.color} className="h-[14px] w-[14px] shrink-0" />
                    </a>
                  )}
                </td>
                {showMin && (
                  <td
                    className={'text-right pr-2 py-0.5 text-xs font-medium whitespace-nowrap'}
                    // style={{ color: entry.color }}
                  >
                    {unitFn ? unitFn(metric.min) : metric.min}
                  </td>
                )}
                {showMax && (
                  <td
                    className={'text-right pr-2 py-0.5 text-xs font-medium whitespace-nowrap'}
                    // style={{ color: entry.color }}
                  >
                    {unitFn ? unitFn(metric.max) : metric.max}
                  </td>
                )}
                {showAverage && (
                  <td
                    className={'text-right pr-2 py-0.5 text-xs font-medium whitespace-nowrap'}
                    // style={{ color: entry.color }}
                  >
                    {unitFn ? unitFn(metric.average) : metric.average}
                  </td>
                )}
                {showLast && (
                  <td
                    className={'text-right pr-2 py-0.5 text-xs font-medium whitespace-nowrap'}
                    // style={{ color: entry.color }}
                  >
                    {unitFn ? unitFn(metric.last) : metric.last}
                  </td>
                )}
                {showTotal && (
                  <td
                    className={'text-right pr-2 py-0.5 text-xs font-medium whitespace-nowrap'}
                    // style={{ color: entry.color }}
                  >
                    {unitFn ? unitFn(metric.total) : metric.total}
                  </td>
                )}
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
};

export {PromLegendTable};
export type {PromLegendTableProps};
