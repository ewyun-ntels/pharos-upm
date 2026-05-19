'use client';
import React, {useMemo, useState} from 'react';
import {VariableBarHeader} from './VariableBarHeader';
import {VariableBarBody} from './VariableBarBody';
import {useDashboardStore} from '@features/dashboard/hooks/use-dashboard-store';
import {FilterConfig} from './types';
import {useVariableLayout} from '@features/dashboard/hooks/useVariableLayout';
import {useFullscreenStore} from '@features/dashboard/hooks/fullscreenStore';

interface VariableBarTemplateProps {
  id: string;
  filters?: FilterConfig[];
  body?: React.ReactNode;
}

export const VariableBarTemplate = ({
  id,
  filters = [],
  body,
}: VariableBarTemplateProps) => {
  const filterMetas = useDashboardStore((state) => state.filterState?.filterMetas ?? new Map());
  const datetime = useDashboardStore((state) => state.filterState?.datetime);
  const step = useDashboardStore((state) => state.filterState?.step);
  const refreshCount = useDashboardStore((state) => state.filterState?.refreshCount ?? 0);

  const { headerL, headerR } = useVariableLayout(filters);

  const memoizedFilterState = useMemo(() => {
    const obj: Record<string, any> = {};
    filterMetas.forEach((meta, key) => {
      obj[key] = meta.value;
    });
    obj.datetime = datetime;
    obj.step = step;
    obj.refreshCount = refreshCount;
    return obj;
  }, [filterMetas, datetime, step, refreshCount]);

  const isFullscreen = useFullscreenStore((state) => state.isFullscreen);

  const bodyWithFilterState = useMemo(() => {
    return React.Children.map(body, (child) => {
      if (React.isValidElement(child) && typeof child.type !== 'string') {
        return React.cloneElement(child as React.ReactElement<any>, {
          filterState: memoizedFilterState,
        });
      }
      return child;
    });
  }, [body, memoizedFilterState]);

  return (
    <div className="flex w-full h-full bg-dashboard-background">
      {/* 나머지 영역: 헤더 sticky + 스크롤 가능한 body */}
      <div className={`flex flex-col flex-1 min-w-0 overflow-y-auto${isFullscreen ? ' pt-4' : ''}`}>
        {!isFullscreen && (
          <div className="sticky top-0 z-10 shrink-0 pb-3 bg-dashboard-background">
            <VariableBarHeader
              dashboardId={id}
              leftItems={headerL}
              rightItems={headerR}
            />
          </div>
        )}
        {/* ✅ 블로킹 제거: 패널이 스스로 variable 준비 여부 판단 */}
        <VariableBarBody {...memoizedFilterState}>{bodyWithFilterState}</VariableBarBody>
      </div>
    </div>
  );
};
