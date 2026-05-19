'use client';

import {CardContent, CardDescription, CardHeader, CardTitle} from '@pharos/shared/components/ui';
import React from 'react';
import {DataOverview, DataOverviewProps} from './dataOverview';
import {z} from 'zod';
import {Panel} from '@pharos/shared/types/dashboard';
import {ChartCard} from '@features/dashboard/panels/plugins/common/card';

const CardPropsSchema = z.custom<Panel>();

type DataOverviewCardProps = z.infer<typeof CardPropsSchema>;

type CardProps = DataOverviewCardProps & DataOverviewProps;

// type DataOverviewCardProps = {
//   data: any;
//   datas: { name: string; path: string }[];
//   name?: string;
//   layoutCol?: boolean;
//   titleTop?: boolean;
//   largeText?: boolean;
//   right?: boolean;
// };

function DataOverviewCard(props: CardProps) {
  return (
    <ChartCard>
      <CardHeader className="text-sm p-2 pb-3 px-4 [&>div]:leading-5">
        <CardTitle>{props.title}</CardTitle>
        <CardDescription>{props.description}</CardDescription>
      </CardHeader>
      <CardContent className="p-4 py-0 overflow-auto" style={{height: `calc(100%`}}>
        <DataOverview {...(props as CardProps)} datas={props.datas ?? []} />
      </CardContent>
    </ChartCard>
  );
}

export {DataOverviewCard};
