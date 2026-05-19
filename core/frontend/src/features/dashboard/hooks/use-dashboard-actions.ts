import React, { useState } from 'react';
import { useList, useUpdate, useDelete } from '@/lib/data-provider';
import { useNavigate } from 'react-router-dom';
import { useToast } from '@hooks/use-toast';
import { useFavoritesStore } from './use-favorites-store';
import { useDashboardNavStore } from './use-dashboard-nav-store';
import { useDashboardStore } from '@features/dashboard/hooks/use-dashboard-store';
import { DASHBOARD_PROVIDER_NAME, DASHBOARD_RESOURCES } from '@providers/dashboard-provider';
import { UI_CONFIG_PROVIDER_NAME, UI_CONFIG_RESOURCES } from '@providers/ui-config-provider';
import { UseDashboardActionsReturn } from './types';
import { DashboardConfig, DashboardData, DashboardFolder } from '@pharos/shared/types/dashboard';

export function useDashboardActions(): UseDashboardActionsReturn {
  const { addFavorite, removeFavorite } = useFavoritesStore();
  const { toast } = useToast();
  const navigate = useNavigate();
  const [favoriteLoadingId, setFavoriteLoadingId] = useState<string | null>(null);
  const [homeId, setHomeId] = useState<string>('');
  const {
    deleteDashboard,
    createDashboard,
    updateFavorite: updateFavoriteStore,
  } = useDashboardStore();

  const { query: { data: dashboardsData, refetch: refetchDashboards } } = useList<DashboardData>({
    resource: DASHBOARD_RESOURCES.DASHBOARD,
    dataProviderName: DASHBOARD_PROVIDER_NAME,
  });

  const { query: { data: homeData } } = useList<{ dashboard_id: string }>({
    resource: UI_CONFIG_RESOURCES.HOME,
    dataProviderName: UI_CONFIG_PROVIDER_NAME,
  });

  const { mutateAsync: updateConfig } = useUpdate();
  const { mutateAsync: deleteConfig } = useDelete();

  React.useEffect(() => {
    const id = (homeData?.data as any)?.dashboard_id;
    if (id) setHomeId(id);
  }, [homeData]);

  // Extract folders from dashboard list meta
  const folders: DashboardFolder[] = (dashboardsData as any)?.meta?.folders ?? [];

  const handleSetHomeId = async (id: string) => {
    try {
      if (id === homeId) {
        // 이미 홈으로 설정되어 있으면 해제
        await deleteConfig({
          resource: UI_CONFIG_RESOURCES.HOME,
          id: '',
          dataProviderName: UI_CONFIG_PROVIDER_NAME,
        });
        setHomeId('');
        toast({ description: 'Home unset successfully' });
      } else {
        // 새로운 홈 설정
        await updateConfig({
          resource: UI_CONFIG_RESOURCES.HOME,
          id: '',
          values: { dashboard_id: id },
          dataProviderName: UI_CONFIG_PROVIDER_NAME,
        });
        setHomeId(id);
        toast({ description: 'Home set successfully' });
      }
      refetchDashboards();
    } catch (error) {
      console.error('Error:', error);
      toast({ description: 'Failed to update home', variant: 'destructive' });
    }
  };

  const updateFavorite = async (data: DashboardData) => {
    const { id, config } = data;
    setFavoriteLoadingId(id);

    try {
      await updateFavoriteStore(id, !config.favorite, config);
      toast({ description: 'Updated successfully' });
      setFavoriteLoadingId(null);

      if (config.favorite) {
        removeFavorite(id);
      } else {
        addFavorite({ id, displayName: config.displayName || config.title });
      }

      void useDashboardNavStore.getState().loadNav();
      refetchDashboards();
    } catch (error) {
      toast({ description: 'Update failed', variant: 'destructive' });
      setFavoriteLoadingId(null);
      console.error('Error:', error);
    }
  };

  const handleDelete = async (
    selectedItems: string[],
    setSelectedItems: React.Dispatch<React.SetStateAction<string[]>>,
  ) => {
    if (selectedItems.length === 0) return;

    const deletePromises = selectedItems.map(async (id) => {
      try {
        await deleteDashboard(id);
        return { id, success: true };
      } catch (error) {
        console.error(`Failed to delete dashboard ${id}:`, error);
        return { id, success: false, error };
      }
    });

    const results = await Promise.allSettled(deletePromises);

    const successCount = results.filter(
      (result) => result.status === 'fulfilled' && result.value.success,
    ).length;

    const failureCount = selectedItems.length - successCount;

    if (failureCount === 0) {
      toast({ description: `${successCount} dashboard(s) deleted successfully` });
    } else if (successCount === 0) {
      toast({ description: `Failed to delete ${failureCount} dashboard(s)`, variant: 'destructive' });
    } else {
      toast({ description: `${successCount} deleted successfully, ${failureCount} failed`, variant: 'destructive' });
    }

    if (successCount > 0) {
      refetchDashboards();
    }

    setSelectedItems([]);
  };

  const handleNewDashboard = async () => {
    const jsonData: Partial<DashboardConfig> = {
      description: '',
      favorite: false,
      title: 'New Dashboard',
      panels: [],
      filters: [
        {
          id: 'datetime',
          type: 'datetime',
          kind: 'headerR',
        },
      ],
      type: 'dashboard',
    };

    try {
      const result = await createDashboard(jsonData);

      if (result && result.id) {
        navigate('/dashboards/' + result.id);
      } else {
        toast({ description: 'Dashboard created but navigation failed', variant: 'destructive' });
      }
    } catch (error) {
      console.error('[handleNewDashboard] Error:', error);
      toast({ description: 'Failed to save dashboard', variant: 'destructive' });
    }
  };

  const { saveDashboard, updateDashboard } = useDashboardStore();

  return {
    updateFavorite,
    favoriteLoadingId,
    handleDelete,
    handleNewDashboard,
    handleSetHomeId,
    homeId,
    dashboardsData,
    folders,
    refetchDashboards,
    saveDashboard,
    updateDashboard,
  };
}
