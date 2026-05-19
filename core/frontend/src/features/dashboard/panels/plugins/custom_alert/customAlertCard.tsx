'use client';

import {CardContent, CardHeader, CardTitle, Tooltip, TooltipContent, TooltipTrigger} from '@pharos/shared/components/ui';
import React from 'react';
import AlertTable from './alertTable';
import {ChartCard} from '../common/card';
import {PanelProps} from '@pharos/core/panel-registry';
import {CustomAlertPanelOptions} from './types';
import {Info} from 'lucide-react';

function CustomAlertCard(props: PanelProps<CustomAlertPanelOptions>) {
  return (
    <ChartCard className={props.bgTransparent ? 'bg-transparent border-none' : ''}>
      <CardHeader className="text-sm p-2 px-4 pb-3">
        <CardTitle className="flex items-center gap-2">
          {props.title}
          {props.description && (
            <Tooltip>
              <TooltipTrigger asChild>
                <Info className="h-4 w-4 text-muted-foreground cursor-help" />
              </TooltipTrigger>
              <TooltipContent>
                <p className="whitespace-pre-line">{props.description}</p>
              </TooltipContent>
            </Tooltip>
          )}
        </CardTitle>
      </CardHeader>
      <CardContent className="p-4 pt-0 pb-3 overflow-auto" style={{height: `calc(100% - 60px)`}}>
        <AlertTable {...props} />
      </CardContent>
    </ChartCard>
  );
}

export {CustomAlertCard};
