import {create} from 'zustand';
import {dashboardProvider, DASHBOARD_RESOURCES} from '@providers/dashboard-provider';
import {registerOnLogout} from '@pharos/shared/features/auth';
import type {DashboardData, DashboardFolder} from '@pharos/shared/types/dashboard';

interface DashboardNavStore {
  dashboards: DashboardData[];
  folders: DashboardFolder[];
  loadNav: () => Promise<void>;
}

export const useDashboardNavStore = create<DashboardNavStore>()((set) => ({
  dashboards: [],
  folders: [],

  loadNav: async () => {
    try {
      const result = await dashboardProvider.getList({
        resource: DASHBOARD_RESOURCES.DASHBOARD,
        pagination: undefined,
      });
      const folders: DashboardFolder[] = (result as any)?.meta?.folders ?? [];
      set({dashboards: result.data as DashboardData[], folders});
    } catch (error) {
      console.error('[DashboardNavStore] Failed to load nav data:', error);
    }
  },
}));

// 로그아웃 시 초기화
registerOnLogout(() => {
  useDashboardNavStore.setState({dashboards: [], folders: []});
});
