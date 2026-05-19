import React from 'react';
import {Button} from '@pharos/shared/components/ui';
import {ResponsiveLayouts} from 'react-grid-layout';
import {DashboardConfig} from '@pharos/shared/types/dashboard';
import {addPanel} from './utils/gridUtils';
import {convertJsonToGridLayout} from './utils/layoutUtils';

interface GridButtonGroupProps {
  dashboard: DashboardConfig;
  setDashboard: React.Dispatch<React.SetStateAction<DashboardConfig>>;
  setGridLayouts: React.Dispatch<React.SetStateAction<ResponsiveLayouts>>;
  onResetLayout: () => void;
}

export default function GridButtonGroup({
  dashboard,
  setDashboard,
  setGridLayouts,
  onResetLayout,
}: GridButtonGroupProps) {
  const handleNewPanel = (type: 'row' | 'card') => {
    const newDashboard = addPanel(dashboard, type);
    setDashboard(newDashboard);
    const newLayout = convertJsonToGridLayout(newDashboard.panels ?? []);
    setGridLayouts({lg: newLayout});
  };

  return (
    <div className="absolute right-4 bottom-4 z-10 flex gap-2">
      <Button onClick={() => handleNewPanel('row')} size="sm" variant="default">
        + Add Row
      </Button>
      <Button onClick={() => handleNewPanel('card')} size="sm" variant="default">
        + Add Card
      </Button>
      <Button onClick={() => onResetLayout()} size="sm" variant="outline">
        Reset
      </Button>
    </div>
  );
}
