import {useCallback} from 'react';
import {folderService} from './use-folder-service';

/**
 * Provides folder mutation operations.
 * The caller is responsible for calling refetchDashboards() after mutations
 * to refresh the list (folders come back via useDashboardActions).
 */
export function useFolders() {
  const createFolder = useCallback(async (name: string) => {
    await folderService.create(name);
  }, []);

  const renameFolder = useCallback(async (id: string, name: string) => {
    await folderService.rename(id, name);
  }, []);

  const deleteFolder = useCallback(async (id: string) => {
    await folderService.delete(id);
  }, []);

  const moveDashboard = useCallback(async (dashboardId: string, folderId: string | null) => {
    await folderService.moveDashboard(dashboardId, folderId);
  }, []);

  return {createFolder, renameFolder, deleteFolder, moveDashboard};
}
