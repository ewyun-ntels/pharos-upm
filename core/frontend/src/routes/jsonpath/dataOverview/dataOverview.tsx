'use client';

import React from "react";
import { cn } from "@/lib/utils";
import jsonpath from 'jsonpath';
import { z } from 'zod';

// DataOverview 전용 레이아웃 옵션 (차트 플러그인과 독립적)
const DataOverviewPropsSchema = z.object({
  data: z.any(),
  datas: z.array(z.object({
    name: z.string(),
    path: z.string(),
  })),
  // DataOverview UI 레이아웃 옵션
  layoutCol: z.boolean().optional(),
  titleTop: z.boolean().optional(),
  largeText: z.boolean().optional(),
  right: z.boolean().optional(),
});

type DataOverviewProps = z.infer<typeof DataOverviewPropsSchema>;

const extractAllValues = (obj: any, path: string) => {
  const results = jsonpath.query(obj, path);
  return results.length > 0 ? results : [];
};

function DataOverview({ 
  data,
  datas,
  layoutCol = false,
  titleTop = false,
  largeText = false,
  right = false,
}: DataOverviewProps) {

  return (
    <ul 
      className={cn("flex gap-3",
        layoutCol ? "flex-row" : "flex-col"
      )}
    >
      {datas?.map(({ name, path }, index) => {
        const values = extractAllValues(data, path);
        return (
          <li
            key={index}
            className={cn(
              "flex w-full gap-1",
              titleTop ? "flex-col items-start gap-0.5" : "flex-row items-center justify-between",
            )}
          >
            <span
              className={cn(
                "text-muted-foreground text-sm",
                titleTop && "w-full text-xs pb-1",
                right && "text-right"
              )}>
              {name}
            </span>
            <span className={cn(
                "font-semibold", 
                titleTop ? "w-full text-sm pb-2" : "text-md",
                largeText && "text-xl font-semibold",
                right && "text-right"
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
export type { DataOverviewProps };
