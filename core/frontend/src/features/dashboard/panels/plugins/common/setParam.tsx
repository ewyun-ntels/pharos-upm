import {DASHBOARD_PROVIDER_NAME, DASHBOARD_RESOURCES} from '@providers/dashboard-provider';
import type {ChartQuery} from '@pharos/shared/types/dashboard';

// Dashboard 차트를 위한 표준 dataProvider 설정
export const createDashboardChartProvider = (chartQuery?: ChartQuery[]) => {
  return {
    dataProvider: {
      chartQuery,
      dataProviderName: DASHBOARD_PROVIDER_NAME,
      resource: DASHBOARD_RESOURCES.TIMESERIES,
    },
  };
};
