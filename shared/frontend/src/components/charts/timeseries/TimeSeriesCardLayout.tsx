/**
 * TimeSeriesCardLayout - Pure Card Layout Component
 * 
 * 순수한 Card UI wrapper입니다. 데이터 fetching 로직이 없습니다.
 * 
 * 용도:
 * - 어디서든 TimeSeries 차트를 Card로 감싸고 싶을 때
 * - title, description과 함께 차트 표시
 * 
 * @example
 * <TimeSeriesCardLayout
 *   title="CPU Usage"
 *   description="Last 24 hours"
 *   bgTransparent={false}
 * >
 *   <TimeSeriesChart {...chartProps} />
 * </TimeSeriesCardLayout>
 */

'use client';

import React from 'react';
import {Card, CardContent, CardHeader, CardTitle, Tooltip, TooltipContent, TooltipTrigger} from '../../ui';
import {cn} from '../../../lib';
import {Info} from 'lucide-react';

export interface TimeSeriesCardLayoutProps {
  /** Card title */
  title?: string;
  
  /** Card description */
  description?: string;
  
  /** Transparent background */
  bgTransparent?: boolean;
  
  /** Chart content */
  children: React.ReactNode;
  
  /** Additional className */
  className?: string;
}

/**
 * TimeSeriesCardLayout Component
 * 
 * 순수한 Card layout입니다. children으로 차트를 받습니다.
 */
export function TimeSeriesCardLayout({
  title,
  description,
  bgTransparent = false,
  children,
  className,
}: TimeSeriesCardLayoutProps) {
  return (
    <Card 
      className={cn(
        'flex flex-col h-full gap-0 py-2 pb-0 rounded-none shadow-none border-none',
        bgTransparent && 'bg-transparent',
        className
      )}
    >
      {title && (
        <CardHeader className="p-2 px-4 overflow-hidden">
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            {title}
            {description && (
              <Tooltip>
                <TooltipTrigger asChild>
                  <Info className="h-4 w-4 text-muted-foreground cursor-help" />
                </TooltipTrigger>
                <TooltipContent>
                  <p className="whitespace-pre-line">{description}</p>
                </TooltipContent>
              </Tooltip>
            )}
          </CardTitle>
        </CardHeader>
      )}
      <CardContent 
        className="flex-1 px-0 pb-0"
        style={{minHeight: '10px'}}
      >
        {children}
      </CardContent>
    </Card>
  );
}
