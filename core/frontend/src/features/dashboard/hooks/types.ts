import React from 'react';
import {DashboardConfig, DashboardData, DashboardFolder} from '@pharos/shared/types/dashboard';
import {GetListResponse} from '@/lib/data-provider';

/**
 * Return type of the useDashboardActions hook.
 */
export interface UseDashboardActionsReturn {
  updateFavorite: (data: DashboardData) => Promise<void>;
  favoriteLoadingId: string | null;
  handleDelete: (
    selectedItems: string[],
    setSelectedItems: React.Dispatch<React.SetStateAction<string[]>>
  ) => Promise<void>;
  handleNewDashboard: () => Promise<void>;
  handleSetHomeId: (id: string) => Promise<void>;
  homeId: string;
  dashboardsData: GetListResponse<DashboardData> | undefined;
  folders: DashboardFolder[];
  refetchDashboards: () => void;
  saveDashboard: (options?: { saveTimeRange?: boolean; saveStep?: boolean; saveRefreshInterval?: boolean }) => Promise<void>;
  updateDashboard: (updates: Partial<DashboardConfig>) => void;
}
