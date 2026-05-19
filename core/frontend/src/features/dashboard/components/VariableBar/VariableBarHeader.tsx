'use client';

import {VariableItems} from './VariableItems';
import {LayoutHeaderProps} from './types';

export const VariableBarHeader = ({
  dashboardId,
  leftItems,
  rightItems,
}: LayoutHeaderProps) => (
  <div className="w-full">
    <div className="flex flex-wrap items-start gap-2 px-4 pt-3">
      <div className="flex items-center gap-2 flex-wrap">
        <VariableItems dashboardId={dashboardId} items={leftItems} />
      </div>
      <div className="flex items-center gap-4 ml-auto">
        <VariableItems dashboardId={dashboardId} items={rightItems} />
      </div>
    </div>
  </div>
);
