import {
  HorizontalAlignmentType,
  VerticalAlignmentType,
} from 'recharts/types/component/DefaultLegendContent';
import * as React from 'react';

export type UniqueMetricValue = {
  key: string;
  min: number | undefined;
  max: number | undefined;
  average: number | undefined;
  last: number | undefined;
  total: number | undefined;
  fill?: string; // pie chart에서 사용
};

export type UniqueMetric = Record<string, UniqueMetricValue>;

export type ChartMetricLabel = {
  metric: any;
  _query: string;
  _name: string;
  _label: string;
};

export type LegendType = {
  align: HorizontalAlignmentType | undefined;
  verticalAlign: VerticalAlignmentType | undefined;
  className: string;
  wrapperStyleWidth: React.CSSProperties;
};
export const legendTypeBottom: LegendType = {
  align: 'center',
  verticalAlign: 'bottom',
  //일반 bottom형 legend는 max값 2줄, 줄바뀜, 센터정렬
  className: 'w-full h-full [&>div]:flex-row [&>div]:flex-wrap [&>div]:justify-center',
  wrapperStyleWidth: {right: 0, top: '70%', height: '30%', width: '100%'},
};

// 아래 소스 legendTypeBottomTable가 적용되고 있지 않음
export const legendTypeBottomTable: LegendType = {
  align: 'center',
  verticalAlign: 'bottom',
  className: 'h-full',
  wrapperStyleWidth: {right: 0, top: '70%', height: '30%', width: '100%'},
};

export const legendTypeRight: LegendType = {
  align: 'right',
  verticalAlign: 'middle',
  //일반 right형 legend는 데이터 적을때 middle정렬, 여러줄일 때 start로 정렬되면서 스크롤 auto
  className: 'h-full [&>div]:flex-col pr-1',
  wrapperStyleWidth: {
    paddingLeft: 10,
    paddingBottom: 10,
    paddingTop: 10,
    paddingRight: 0,
    top: '0%',
    right: 0,
    height: '100%',
    width: '50%',
  },
};
