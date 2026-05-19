'use client';

import * as React from 'react';
import * as TabsPrimitive from '@radix-ui/react-tabs';

import {cn} from '../../lib';

export function getBgColorActive(title: string) {
  switch (title.toLowerCase()) {
    case 'critical':
      return `bg-[#B91C1C] text-white group-data-[state=active]:bg-[#B91C1C]`;
    case 'major':
      return 'bg-[#F97316] text-white group-data-[state=active]:bg-[#F97316]';
    case 'minor':
      return 'bg-[#EFA926] text-white group-data-[state=active]:bg-[#EFA926]';
    case 'low':
      return 'bg-[#0891B2] text-white group-data-[state=active]:bg-[#0891B2]';
    case 'normal':
      return 'bg-[#16A34A] text-white group-data-[state=active]:bg-[#16A34A]';
    default:
      return 'bg-[#64748B] text-white group-data-[state=active]:bg-[#64748B]';
  }
}

function Tabs({className, ...props}: React.ComponentProps<typeof TabsPrimitive.Root>) {
  return (
    <TabsPrimitive.Root
      data-slot="tabs"
      className={cn('flex flex-col gap-2', className)}
      {...props}
    />
  );
}

function TabsList({className, ...props}: React.ComponentProps<typeof TabsPrimitive.List>) {
  return (
    <TabsPrimitive.List
      data-slot="tabs-list"
      className={cn(
        'bg-background border text-muted-foreground inline-flex h-8 w-fit items-center justify-center rounded-lg p-1',
        className,
      )}
      {...props}
    />
  );
}

function TabsTrigger({
  className,
  badge,
  children,
  ...props
}: React.ComponentProps<typeof TabsPrimitive.Trigger> & {
  badge?: React.ReactNode;
}) {
  const bgColor = getBgColorActive(children as string);
  return (
    <TabsPrimitive.Trigger
      data-slot="tabs-trigger"
      className={cn(
        "group data-[state=active]:bg-muted data-[state=active]:text-foreground focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:outline-ring inline-flex items-center justify-center gap-1 rounded-md px-1.5 py-0.5 text-sm font-medium whitespace-nowrap transition-[color,box-shadow] focus-visible:ring-[3px] focus-visible:outline-1 disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4 cursor-pointer",
        className,
      )}
      {...props}
    >
      {children}
      {badge !== undefined && badge !== null && (
        <span
          className={cn(
            'min-w-[24px] ml-1 inline-flex items-center justify-center rounded-full text-xs py-0.5 font-semibold bg-foreground/10 text-foreground group-data-[state=active]:text-white',
            `${bgColor}`,
          )}
        >
          {badge}
        </span>
      )}
    </TabsPrimitive.Trigger>
  );
}

function TabsContent({className, ...props}: React.ComponentProps<typeof TabsPrimitive.Content>) {
  return (
    <TabsPrimitive.Content
      data-slot="tabs-content"
      className={cn('flex-1 outline-none', className)}
      {...props}
    />
  );
}

export {Tabs, TabsList, TabsTrigger, TabsContent};
