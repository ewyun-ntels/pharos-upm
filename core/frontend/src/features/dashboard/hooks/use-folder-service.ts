import {axiosInstance} from '@lib/axios';
import type {DashboardFolder} from '@pharos/shared/types/dashboard';

const API_URL = '/folders';
const DASHBOARD_API_URL = '/dashboard';

export const folderService = {
  list(): Promise<DashboardFolder[]> {
    return axiosInstance.get<DashboardFolder[]>(API_URL).then((r) => r.data ?? []);
  },

  create(name: string): Promise<DashboardFolder> {
    return axiosInstance.post<DashboardFolder>(API_URL, {name}).then((r) => r.data);
  },

  rename(id: string, name: string): Promise<void> {
    return axiosInstance.put(`${API_URL}/${id}`, {name}).then(() => undefined);
  },

  delete(id: string): Promise<void> {
    return axiosInstance.delete(`${API_URL}/${id}`).then(() => undefined);
  },

  moveDashboard(dashboardId: string, folderId: string | null): Promise<void> {
    return axiosInstance
      .patch(`${DASHBOARD_API_URL}/${dashboardId}/folder`, {folderId})
      .then(() => undefined);
  },
};
