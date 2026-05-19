'use client';

import React, { useMemo } from 'react';
import { Card, CardContent, CardHeader, CardTitle, Tooltip, TooltipContent, TooltipTrigger } from '@pharos/shared/components/ui';
import { cn } from '@lib/utils';
import { Info } from 'lucide-react';
import type { ChartMetricData } from '@types';
import type { PanelProps } from '@pharos/core/panel-registry';
import type { FilterMeta } from '@features/dashboard/hooks/slices/types';
import { withChartPanel, PanelContentProps } from './withChartPanel';
import { replaceVariables } from '@features/dashboard/utils/variable-parser';

// ─── PanelCard (내부 구현) ────────────────────────────────────────────────────

interface PanelCardProps {
  title?: string;
  description?: string;
  bgTransparent?: boolean;
  filterMetas?: Map<string, FilterMeta>;
  children: React.ReactNode;
}

function PanelCard({ title, description, bgTransparent = false, filterMetas, children }: PanelCardProps) {
  const displayTitle = useMemo(() => replaceVariables(title, filterMetas), [title, filterMetas]);
  const displayDescription = useMemo(() => replaceVariables(description, filterMetas), [description, filterMetas]);

  return (
    <Card
      className={cn(
        'flex flex-col h-full gap-0 py-2 pb-0 rounded-none shadow-none border-none',
        bgTransparent && 'bg-transparent',
      )}
    >
      {displayTitle && (
        <CardHeader className="p-2 px-4 overflow-hidden">
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            {displayTitle}
            {displayDescription && (
              <Tooltip>
                <TooltipTrigger asChild>
                  <Info className="h-4 w-4 text-muted-foreground cursor-help" />
                </TooltipTrigger>
                <TooltipContent>
                  <p className="whitespace-pre-line">{displayDescription}</p>
                </TooltipContent>
              </Tooltip>
            )}
          </CardTitle>
        </CardHeader>
      )}
      <CardContent className="flex-1 min-h-0 px-0 pb-0" style={{ minHeight: '10px' }}>
        {children}
      </CardContent>
    </Card>
  );
}

// ─── withCardChartPanel HOC ───────────────────────────────────────────────────

/**
 * withCardChartPanel HOC
 *
 * withChartPanel을 재사용하여 Card UI만 추가합니다.
 * 데이터 fetching, ResizeObserver 등은 withChartPanel이 처리합니다.
 *
 * 사용 예:
 *   component: withCardChartPanel(MyChartContent)  // Card 포함 차트/테이블 패널
 *   component: withChartPanel(MyContent)           // Card 없이 전체 영역
 */
export function withCardChartPanel<T, TData = ChartMetricData>(
  Component: React.ComponentType<PanelContentProps<T, TData>>,
): React.ComponentType<PanelProps<T>> {
  // ✅ withChartPanel을 재사용 (중복 로직 제거)
  const ChartPanel = withChartPanel<T, TData>(Component);

  function WithCardChartPanel(props: PanelProps<T>) {
    return (
      <PanelCard
        title={props.title}
        description={props.description}
        bgTransparent={props.bgTransparent}
        filterMetas={props.filterMetas}
      >
        <ChartPanel {...props} />
      </PanelCard>
    );
  }

  WithCardChartPanel.displayName = `WithCardChartPanel(${Component.displayName || Component.name || 'Component'})`;

  return WithCardChartPanel;
}

// ✅ Backward compatibility (deprecated)
export const withTimeseriesCard = withCardChartPanel;
