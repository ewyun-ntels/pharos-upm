
import React from 'react';
import {cn} from '@lib/utils';

import { sampleData } from "./sampleData";
import { DataOverviewCard } from './dataOverview/dataOverviewCard';

export default function DataOverview() {
  return (
    <div className="flex flex-col gap-2 py-4">
      <div className={cn('flex flex-row sm:flex-wrap lg:flex-nowrap grid w-full gap-2 grid-cols-4')}>
         {/*TODO: name, jsonpath는 입력받게 추후 작업  */}
        <DataOverviewCard
          id="system-overview-1"
          title="System Overview"
          layout={{ type: 'card', x: 0, y: 0, w: 6, h: 4 }}
          datas={[
            { name: "Total Usage", path: "$.cpu.usage.total" },
            { name: "User Usage", path: "$.cpu.usage.user" },
            { name: "System Usage", path: "$.cpu.usage.system" },
            { name: "Idle Usage", path: "$.cpu.usage.idle" },
            { name: "IOWait Usage", path: "$.cpu.usage.iowait" },
          ]}
          data={sampleData}
        />
        <DataOverviewCard 
          id="system-overview-2"
          title="System Overview"
          layout={{ type: 'card', x: 6, y: 0, w: 6, h: 4 }}
          datas={[
            { name: "Total Usage", path: "$.cpu.usage.disk" },
            { name: "User Usage", path: "$.cpu.usage.sensor" },
            { name: "System Usage", path: "$.cpu.usage.system" },
            { name: "Idle Usage", path: "$.cpu.usage.idle" },
            { name: "IOWait Usage", path: "$.cpu.usage.iowait" },
          ]}
          data={sampleData}
          largeText={true}
        />
         <DataOverviewCard 
          id="system-overview-3"
          title="System Overview"
          layout={{ type: 'card', x: 12, y: 0, w: 6, h: 4 }}
          datas={[
            { name: "Total Usage", path: "$.cpu.usage.total" },
            { name: "User Usage", path: "$.cpu.usage.user" },
            { name: "System Usage", path: "$.cpu.usage.system" },
            { name: "Idle Usage", path: "$.cpu.usage.idle" },
            { name: "IOWait Usage", path: "$.cpu.usage.iowait" },
          ]}
          data={sampleData}
          titleTop={true}
        />
         <DataOverviewCard 
          id="system-overview-4"
          title="System Overview"
          layout={{ type: 'card', x: 18, y: 0, w: 6, h: 4 }}
          datas={[
            { name: "Total Usage", path: "$.cpu.usage.total" },
            { name: "User Usage", path: "$.cpu.usage.user" },
            { name: "System Usage", path: "$.cpu.usage.system" },
            { name: "Idle Usage", path: "$.cpu.usage.idle" },
            { name: "IOWait Usage", path: "$.cpu.usage.iowait" },
          ]}
          data={sampleData}
          titleTop={true}
          right={true}
        />
      </div>
      <div className={cn('flex flex-row sm:flex-wrap lg:flex-nowrap grid w-full gap-2 grid-cols-2')}>
        <DataOverviewCard 
          id="system-overview-5"
          title="System Overview - layoutCol LargeText"
          layout={{ type: 'card', x: 0, y: 4, w: 12, h: 6 }}
          datas={[
            { name: "Total Usage", path: "$.cpu.usage.total" },
            { name: "User Usage", path: "$.cpu.usage.user" },
            { name: "System Usage", path: "$.cpu.usage.system" },
            { name: "Idle Usage", path: "$.cpu.usage.idle" },
            { name: "IOWait Usage", path: "$.cpu.usage.iowait" },
          ]}
          data={sampleData}
          layoutCol={true}
          titleTop={true}
          largeText={true}
        />
         <DataOverviewCard 
          id="system-overview-6"
          title="System Overview - layoutCol"
          layout={{ type: 'card', x: 12, y: 4, w: 12, h: 6 }}
          datas={[
            { name: "Total Usage", path: "$.cpu.usage.total" },
            { name: "User Usage", path: "$.cpu.usage.user" },
            { name: "System Usage", path: "$.cpu.usage.system" },
            { name: "Idle Usage", path: "$.cpu.usage.idle" },
            { name: "IOWait Usage", path: "$.cpu.usage.iowait" },
          ]}
          data={sampleData}
          layoutCol={true}
          titleTop={true}
        />
        <DataOverviewCard 
          id="system-overview-7"
          title="System Overview - layoutCol LargeText"
          layout={{ type: 'card', x: 0, y: 10, w: 12, h: 6 }}
          datas={[
            { name: "Total Usage", path: "$.cpu.usage.total" },
            { name: "User Usage", path: "$.cpu.usage.user" },
            { name: "System Usage", path: "$.cpu.usage.system" },
            { name: "Idle Usage", path: "$.cpu.usage.idle" },
            { name: "IOWait Usage", path: "$.cpu.usage.iowait" },
          ]}
          data={sampleData}
          layoutCol={true}
          titleTop={true}
          largeText={true}
          right={true}
        />
         <DataOverviewCard 
          id="system-overview-8"
          title="System Overview - layoutCol"
          layout={{ type: 'card', x: 12, y: 10, w: 12, h: 6 }}
          datas={[
            { name: "Total Usage", path: "$.cpu.usage.total" },
            { name: "User Usage", path: "$.cpu.usage.user" },
            { name: "System Usage", path: "$.cpu.usage.system" },
            { name: "Idle Usage", path: "$.cpu.usage.idle" },
            { name: "IOWait Usage", path: "$.cpu.usage.iowait" },
          ]}
          data={sampleData}
          layoutCol={true}
          titleTop={true}
          right={true}
        />
      </div>
    </div>
  );
}
