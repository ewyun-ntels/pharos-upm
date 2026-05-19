import {DashboardConfig} from '@pharos/shared/types/dashboard';
import {Panel} from '@pharos/shared/types/dashboard';

export function changeRowTitle(dashboard: DashboardConfig, id: string, title: string): DashboardConfig {
  if (!dashboard.panels) return dashboard;
  
  const updatedPanels = dashboard.panels.map((panel: Panel) => {
    if (panel.id === id) {
      return {
        ...panel,
        title,
      };
    }
    return panel;
  });

  return {...dashboard, panels: updatedPanels};
}
