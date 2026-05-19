import { useCallback } from 'react';
import { useDashboardStore } from './use-dashboard-store';
import { useFavoritesStore } from './use-favorites-store';
import { useDashboardNavStore } from './use-dashboard-nav-store';
import { useToast } from '@hooks/use-toast';
import type { DashboardConfig } from '@pharos/shared/types/dashboard';

type GeneralConfig = Pick<DashboardConfig, 'title' | 'displayName' | 'description' | 'favorite'>;

/**
 * Dashboard General Setting 저장 공통 로직
 *
 * Dashboard Management와 Dashboard Show에서 공통으로 사용
 */
export const useHandleGeneralSave = () => {
  const { updateDashboard, saveDashboard, _loadDashboard } = useDashboardStore();
  const currentDashboardId = useDashboardStore((state) => state.id);
  const { addFavorite, removeFavorite } = useFavoritesStore();
  const loadNav = useDashboardNavStore((state) => state.loadNav);
  const { toast } = useToast();

  const handleGeneralSave = useCallback(
    async (id: string, newConfig: GeneralConfig) => {
      try {
        // 다른 dashboard를 수정하는 경우 먼저 로드
        if (currentDashboardId !== id) {
          await _loadDashboard(id);
        }

        // Only apply metadata fields to avoid overwriting panels/chartQuery with stale data
        const { title, displayName, description, favorite } = newConfig;
        updateDashboard({ title, displayName, description, favorite });
        await saveDashboard();
        toast({ description: 'Saved successfully.' });

        // Update favorites store for sidebar
        if (newConfig.favorite) {
          addFavorite({ id, displayName: newConfig.displayName || newConfig.title });
        } else {
          removeFavorite(id);
        }

        // Reload nav data to refresh sidebar with folder information
        await loadNav();
      } catch (error) {
        console.error('Failed to save dashboard:', error);
        toast({ description: 'Failed to save.', variant: 'destructive' });
        throw error; // re-throw for caller to handle
      }
    },
    [currentDashboardId, _loadDashboard, updateDashboard, saveDashboard, addFavorite, removeFavorite, loadNav, toast],
  );

  return { handleGeneralSave };
};
