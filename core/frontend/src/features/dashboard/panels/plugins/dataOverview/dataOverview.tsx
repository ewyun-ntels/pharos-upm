'use client';

import React from "react";
import { cn } from "@lib/utils";
import jsonpath from 'jsonpath';
import type { PanelProps } from '@pharos/core/panel-registry';
import { DataOverviewPanelOptions } from './types';

type DataOverviewChartProps = PanelProps<DataOverviewPanelOptions> & {
  data?: any;
};

const extractAllValues = (obj: any, path: string) => {
  const results = jsonpath.query(obj, path);
  return results.length > 0 ? results : [];
};

function DataOverview(props: DataOverviewChartProps) {
  const { data, options } = props;
  const datas = options?.datas || [];
  
  const chartOptions = {
    layoutCol: options?.layoutCol || false,
    titleTop: options?.titleTop || false,
    largeText: options?.largeText || false,
    right: options?.right || false,
  };

  return (
    <ul 
      className={cn("flex gap-3",
        chartOptions.layoutCol ? "flex-row" : "flex-col"
      )}
    >
      {datas?.map(({ name, path }, index) => {
        const values = extractAllValues(data, path);
        return (
          <li
            key={index}
            className={cn(
              "flex w-full gap-1",
              chartOptions.titleTop ? "flex-col items-start gap-0.5" : "flex-row items-center justify-between",
            )}
          >
            <span
              className={cn(
                "text-muted-foreground text-sm",
                chartOptions.titleTop && "w-full text-xs pb-1",
                chartOptions.right && "text-right"
              )}>
              {name}
            </span>
            <span className={cn(
                "font-semibold", 
                chartOptions.titleTop ? "w-full text-sm pb-2" : "text-md",
                chartOptions.largeText && "text-xl font-semibold",
                chartOptions.right && "text-right"
              )}>
              {values.length > 0 ? values.join(", ") : "-"}
            </span>
          </li>
        );
      })}
    </ul>
  );
}

export { DataOverview };
export type { DataOverviewChartProps };
