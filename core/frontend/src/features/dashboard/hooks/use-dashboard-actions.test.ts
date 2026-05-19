import { renderHook, act, waitFor } from '@testing-library/react';
import { useDashboardActions } from './use-dashboard-actions';
import { useList, useUpdate } from '@/lib/data-provider';
import { DASHBOARD_PROVIDER_NAME } from '@providers/dashboard-provider';
import { UI_CONFIG_PROVIDER_NAME } from '@providers/ui-config-provider';
import { useToast } from '@hooks/use-toast';
import { useFavoritesStore } from './use-favorites-store';
import { useDashboardStore } from '@features/dashboard/hooks/use-dashboard-store';
import { useDashboardNavStore } from './use-dashboard-nav-store';

// Mock all dependencies
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: jest.fn(),
}));

const mockInvalidate = jest.fn().mockResolvedValue(undefined);
const mockMutateAsync = jest.fn().mockResolvedValue({ data: {} });
const mockDeleteAsync = jest.fn().mockResolvedValue({ data: {} });
jest.mock('@/lib/data-provider', () => ({
  ...jest.requireActual('@/lib/data-provider'),
  useInvalidate: () => mockInvalidate,
  useList: jest.fn(),
  useUpdate: jest.fn(),
  useDelete: jest.fn(),
}));

jest.mock('@hooks/use-toast');
jest.mock('./use-favorites-store');
jest.mock('@features/dashboard/hooks/use-dashboard-store');
jest.mock('./use-dashboard-nav-store', () => ({
  useDashboardNavStore: {
    getState: jest.fn(),
  },
}));
jest.mock('@providers/dashboard-provider', () => ({
  dashboardProvider: {
    update: jest.fn(),
    getList: jest.fn(),
    getOne: jest.fn(),
    create: jest.fn(),
    deleteOne: jest.fn(),
  },
  DASHBOARD_RESOURCES: {
    DASHBOARD: 'dashboard',
    FAVORITES: 'dashboard/favorites',
    TIMESERIES: 'timeseries',
    RAW: 'raw',
  },
  DASHBOARD_PROVIDER_NAME: 'dashboardProvider',
}));

jest.mock('@providers/ui-config-provider', () => ({
  uiConfigProvider: {
    update: jest.fn(),
    deleteOne: jest.fn(),
    getList: jest.fn(),
  },
  UI_CONFIG_RESOURCES: {
    HOME: 'home',
    CONFIG: 'config',
  },
  UI_CONFIG_PROVIDER_NAME: 'uiConfigProvider',
}));

// Import after mocks
import { useNavigate } from 'react-router-dom';
import { dashboardProvider } from '@providers/dashboard-provider';
import { uiConfigProvider } from '@providers/ui-config-provider';
import { useDelete } from '@/lib/data-provider';

// Create typed mocks with explicit types
const mockUseNavigate = jest.mocked(useNavigate);
const mockUseList = jest.mocked(useList);
const mockUseUpdate = jest.mocked(useUpdate);
const mockUseDelete = jest.mocked(useDelete);
const mockUseToast = jest.mocked(useToast);
const mockUseFavoritesStore = jest.mocked(useFavoritesStore);
const mockUseDashboardStore = jest.mocked(useDashboardStore);
const mockUseDashboardNavStore = useDashboardNavStore as unknown as { getState: jest.Mock };
const mockDashboardProviderUpdate = jest.mocked(dashboardProvider.update);
const mockUiConfigProviderUpdate = jest.mocked(uiConfigProvider.update);
const mockUiConfigProviderDeleteOne = jest.mocked(uiConfigProvider.deleteOne);

describe('useDashboardActions', () => {
  const mockNavigate = jest.fn();
  const mockToast = jest.fn();
  const mockRefetchDashboards = jest.fn();
  const mockRefetchSettings = jest.fn();
  const mockDeleteDashboard = jest.fn();
  const mockCreateDashboard = jest.fn();
  const mockUpdateFavoriteStore = jest.fn();
  const mockSetHome = jest.fn();
  const mockLoadNav = jest.fn();
  const mockSaveDashboard = jest.fn();
  const mockUpdateDashboardStore = jest.fn();
  const mockAddFavorite = jest.fn();
  const mockRemoveFavorite = jest.fn();
  const mockSetFavorites = jest.fn();

  const mockDashboardsData = {
    data: [
      { id: '1', title: 'Dashboard 1', displayName: 'Dashboard 1', favorite: false },
      { id: '2', title: 'Dashboard 2', displayName: 'Dashboard 2', favorite: true },
    ],
  };

  const mockSettingData = {
    data: {
      dashboard_id: 'dashboard-1',
    },
  };

  beforeEach(() => {
    jest.clearAllMocks();

    // Setup basic mocks
    mockUseNavigate.mockReturnValue(mockNavigate);
    mockDashboardProviderUpdate.mockResolvedValue({ data: {} });
    mockUiConfigProviderUpdate.mockResolvedValue({ data: {} });
    mockUiConfigProviderDeleteOne.mockResolvedValue({ data: {} });
    mockInvalidate.mockResolvedValue(undefined);
    mockMutateAsync.mockResolvedValue({ data: {} });
    mockDeleteAsync.mockResolvedValue({ data: {} });
    mockUseUpdate.mockReturnValue({ mutateAsync: mockMutateAsync } as any);
    mockUseDelete.mockReturnValue({ mutateAsync: mockDeleteAsync } as any);

    mockUseToast.mockReturnValue({
      toast: mockToast,
      dismiss: jest.fn(),
      toasts: [],
    });

    mockUseDashboardNavStore.getState.mockReturnValue({
      loadNav: mockLoadNav,
    });

    // Zustand store는 함수처럼 동작해야 하므로 mockImplementation 사용
    mockUseFavoritesStore.mockImplementation(
      () =>
        ({
          addFavorite: mockAddFavorite,
          removeFavorite: mockRemoveFavorite,
          setFavorites: mockSetFavorites,
        }) as ReturnType<typeof useFavoritesStore>,
    );

    mockUseDashboardStore.mockImplementation(
      () =>
        ({
          deleteDashboard: mockDeleteDashboard,
          createDashboard: mockCreateDashboard,
          updateFavorite: mockUpdateFavoriteStore,
          setHome: mockSetHome,
          saveDashboard: mockSaveDashboard,
          updateDashboard: mockUpdateDashboardStore,
        }) as ReturnType<typeof useDashboardStore>,
    );

    // react-query QueryObserverResult 타입에 맞게 반환값 확장
    const baseListResult = {
      data: undefined,
      refetch: jest.fn(),
      error: undefined,
      isError: false,
      isLoading: false,
      isLoadingError: false,
      isSuccess: true,
      isFetched: true,
      isRefetching: false,
      status: 'success',
      isRefetchError: false,
      dataUpdatedAt: 0,
      errorUpdatedAt: 0,
      failureCount: 0,
      failureReason: undefined,
      isFetchedAfterMount: true,
      isIdle: false,
      isPaused: false,
      isPlaceholderData: false,
      isPreviousData: false,
      isStale: false,
      remove: jest.fn(),
      fetchStatus: 'idle',
      errorUpdateCount: 0,
      isFetching: false,
      isInitialLoading: false,
    } as any;
    mockUseList.mockImplementation((params) => {
      if (params && params.resource === 'dashboard') {
        return {
          query: {
            ...baseListResult,
            data: mockDashboardsData,
            refetch: mockRefetchDashboards,
          },
        } as any;
      } else if (params && params.resource === 'home') {
        return {
          query: {
            ...baseListResult,
            data: mockSettingData,
            refetch: mockRefetchSettings,
          },
        } as any;
      }
      return {
        query: {
          ...baseListResult,
          data: null,
          refetch: jest.fn(),
        },
      } as any;
    });
  });

  describe('Initial state and data loading', () => {
    it('should initialize with correct initial state', () => {
      const { result } = renderHook(() => useDashboardActions());

      expect(result.current.favoriteLoadingId).toBeNull();
      expect(result.current.dashboardsData).toEqual(mockDashboardsData);
      expect(result.current.homeId).toBe('dashboard-1');
    });

    it('should call useList with correct parameters', () => {
      renderHook(() => useDashboardActions());

      expect(mockUseList).toHaveBeenNthCalledWith(1, {
        resource: 'dashboard',
        dataProviderName: DASHBOARD_PROVIDER_NAME,
      });

      expect(mockUseList).toHaveBeenNthCalledWith(2, {
        resource: 'home',
        dataProviderName: UI_CONFIG_PROVIDER_NAME,
      });
    });

    it('should expose folders from dashboard list meta', () => {
      const folders = [
        { id: 'folder-1', title: 'Folder 1' },
        { id: 'folder-2', title: 'Folder 2' },
      ] as any;

      mockUseList.mockImplementation((params) => {
        if (params && params.resource === 'dashboard') {
          return {
            query: {
              data: {
                data: [],
                meta: { folders },
              },
              refetch: mockRefetchDashboards,
            },
          } as any;
        }

        return {
          query: {
            data: mockSettingData,
            refetch: mockRefetchSettings,
          },
        } as any;
      });

      const { result } = renderHook(() => useDashboardActions());

      expect(result.current.folders).toEqual(folders);
    });
  });

  describe('handleSetHomeId', () => {
    it('should successfully set home dashboard', async () => {
      const { result } = renderHook(() => useDashboardActions());

      await act(async () => {
        await result.current.handleSetHomeId('new-home-id');
      });

      expect(mockMutateAsync).toHaveBeenCalledWith({
        resource: 'home',
        id: '',
        values: { dashboard_id: 'new-home-id' },
        dataProviderName: 'uiConfigProvider',
      });
      expect(mockToast).toHaveBeenCalledWith({
        description: 'Home set successfully',
      });
      expect(mockRefetchDashboards).toHaveBeenCalled();
    });

    it('should unset home when clicking already set home', async () => {
      const { result } = renderHook(() => useDashboardActions());

      // homeId is already set to 'dashboard-1' from mockSettingData
      await act(async () => {
        await result.current.handleSetHomeId('dashboard-1');
      });

      expect(mockDeleteAsync).toHaveBeenCalledWith({
        resource: 'home',
        id: '',
        dataProviderName: 'uiConfigProvider',
      });
      expect(mockToast).toHaveBeenCalledWith({
        description: 'Home unset successfully',
      });
      expect(mockRefetchDashboards).toHaveBeenCalled();
    });

    it('should handle setHome error', async () => {
      const error = new Error('Set home failed');
      mockMutateAsync.mockRejectedValue(error);

      const { result } = renderHook(() => useDashboardActions());

      await act(async () => {
        await result.current.handleSetHomeId('new-home-id');
      });

      expect(mockToast).toHaveBeenCalledWith({
        description: 'Failed to update home',
        variant: 'destructive',
      });
    });
  });

  describe('updateFavorite', () => {
    it('should successfully add to favorites', async () => {
      const { result } = renderHook(() => useDashboardActions());

      const dashboardData = {
        id: 'test-id',
        config: {
          favorite: false,
          title: 'Test Dashboard',
          displayName: 'Test Display Name',
          description: '',
          filters: [],
          type: 'dashboard',
        },
        permission: 'editor' as const,
      };

      await act(async () => {
        await result.current.updateFavorite(dashboardData);
      });

      await waitFor(() => {
        expect(mockUpdateFavoriteStore).toHaveBeenCalledWith(
          'test-id',
          true,
          dashboardData.config,
        );
        expect(mockAddFavorite).toHaveBeenCalledWith({
          id: 'test-id',
          displayName: 'Test Display Name',
        });
        expect(result.current.favoriteLoadingId).toBeNull();
        expect(mockRefetchDashboards).toHaveBeenCalled();
        expect(mockLoadNav).toHaveBeenCalledTimes(1);
      });
    });

    it('should successfully remove from favorites', async () => {
      const { result } = renderHook(() => useDashboardActions());

      const dashboardData = {
        id: 'test-id',
        config: {
          favorite: true,
          title: 'Test Dashboard',
          displayName: 'Test Display Name',
          description: '',
          filters: [],
          type: 'dashboard',
        },
        permission: 'editor' as const,
      };

      await act(async () => {
        await result.current.updateFavorite(dashboardData);
      });

      await waitFor(() => {
        expect(mockUpdateFavoriteStore).toHaveBeenCalledWith(
          'test-id',
          false,
          dashboardData.config,
        );
        expect(mockRemoveFavorite).toHaveBeenCalledWith('test-id');
        expect(result.current.favoriteLoadingId).toBeNull();
      });
    });

    it('should handle updateFavorite error', async () => {
      const error = new Error('Update failed');
      mockUpdateFavoriteStore.mockRejectedValue(error);

      const { result } = renderHook(() => useDashboardActions());

      const dashboardData = {
        id: 'test-id',
        config: {
          favorite: false,
          title: 'Test Dashboard',
          displayName: 'Test Display Name',
          description: '',
          filters: [],
          type: 'dashboard',
        },
        permission: 'editor' as const,
      };

      await act(async () => {
        await result.current.updateFavorite(dashboardData);
      });

      await waitFor(() => {
        expect(mockToast).toHaveBeenCalledWith({
          description: 'Update failed',
          variant: 'destructive',
        });
        expect(result.current.favoriteLoadingId).toBeNull();
        expect(mockLoadNav).not.toHaveBeenCalled();
      });
    });
  });

  describe('handleDelete', () => {
    const mockSetSelectedItems = jest.fn();

    beforeEach(() => {
      mockSetSelectedItems.mockClear();
    });

    it('should return early if no items selected', async () => {
      const { result } = renderHook(() => useDashboardActions());

      await act(async () => {
        await result.current.handleDelete([], mockSetSelectedItems);
      });

      expect(mockDeleteDashboard).not.toHaveBeenCalled();
      expect(mockToast).not.toHaveBeenCalled();
    });

    it('should successfully delete all selected dashboards', async () => {
      mockDeleteDashboard.mockResolvedValue(undefined);

      const { result } = renderHook(() => useDashboardActions());

      await act(async () => {
        await result.current.handleDelete(['id1', 'id2'], mockSetSelectedItems);
      });

      expect(mockDeleteDashboard).toHaveBeenCalledWith('id1');
      expect(mockDeleteDashboard).toHaveBeenCalledWith('id2');
      expect(mockToast).toHaveBeenCalledWith({
        description: '2 dashboard(s) deleted successfully',
      });
      expect(mockRefetchDashboards).toHaveBeenCalled();
      expect(mockSetSelectedItems).toHaveBeenCalledWith([]);
    });

    it('should handle all deletions failing', async () => {
      mockDeleteDashboard.mockRejectedValue(new Error('Delete failed'));

      const { result } = renderHook(() => useDashboardActions());

      await act(async () => {
        await result.current.handleDelete(['id1', 'id2'], mockSetSelectedItems);
      });

      expect(mockToast).toHaveBeenCalledWith({
        description: 'Failed to delete 2 dashboard(s)',
        variant: 'destructive',
      });
      expect(mockRefetchDashboards).not.toHaveBeenCalled();
      expect(mockSetSelectedItems).toHaveBeenCalledWith([]);
    });

    it('should handle partial deletion failures', async () => {
      mockDeleteDashboard
        .mockResolvedValueOnce(undefined)
        .mockRejectedValueOnce(new Error('Delete failed'));

      const { result } = renderHook(() => useDashboardActions());

      await act(async () => {
        await result.current.handleDelete(['id1', 'id2'], mockSetSelectedItems);
      });

      expect(mockToast).toHaveBeenCalledWith({
        description: '1 deleted successfully, 1 failed',
        variant: 'destructive',
      });
      expect(mockRefetchDashboards).toHaveBeenCalled();
      expect(mockSetSelectedItems).toHaveBeenCalledWith([]);
    });
  });

  describe('handleNewDashboard', () => {
    it('should successfully create new dashboard and navigate', async () => {
      const mockResult = { id: 'new-dashboard-id' };
      mockCreateDashboard.mockResolvedValue(mockResult);

      const { result } = renderHook(() => useDashboardActions());

      await act(async () => {
        await result.current.handleNewDashboard();
      });

      expect(mockCreateDashboard).toHaveBeenCalledWith({
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
      });

      expect(mockNavigate).toHaveBeenCalledWith('/dashboards/new-dashboard-id');
    });

    it('should handle createDashboard error', async () => {
      const error = new Error('Create failed');
      mockCreateDashboard.mockRejectedValue(error);

      const { result } = renderHook(() => useDashboardActions());

      // handleNewDashboard는 내부적으로 에러를 catch하므로 에러가 throw되지 않아야 함
      await act(async () => {
        await expect(result.current.handleNewDashboard()).resolves.toBeUndefined();
      });

      expect(mockToast).toHaveBeenCalledWith({
        description: 'Failed to save dashboard',
        variant: 'destructive',
      });
      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it('should handle createDashboard returning null/undefined', async () => {
      mockCreateDashboard.mockResolvedValue(null);

      const { result } = renderHook(() => useDashboardActions());

      await act(async () => {
        await result.current.handleNewDashboard();
      });

      expect(mockNavigate).not.toHaveBeenCalled();
      expect(mockToast).toHaveBeenCalledWith({
        description: 'Dashboard created but navigation failed',
        variant: 'destructive',
      });
    });
  });

  describe('Return values', () => {
    it('should return all expected functions and values', () => {
      const { result } = renderHook(() => useDashboardActions());

      expect(result.current).toHaveProperty('updateFavorite');
      expect(result.current).toHaveProperty('favoriteLoadingId');
      expect(result.current).toHaveProperty('handleDelete');
      expect(result.current).toHaveProperty('handleNewDashboard');
      expect(result.current).toHaveProperty('handleSetHomeId');
      expect(result.current).toHaveProperty('homeId');
      expect(result.current).toHaveProperty('dashboardsData');
      expect(result.current).toHaveProperty('refetchDashboards');
      expect(result.current).toHaveProperty('saveDashboard');
      expect(result.current).toHaveProperty('updateDashboard');

      expect(typeof result.current.updateFavorite).toBe('function');
      expect(typeof result.current.handleDelete).toBe('function');
      expect(typeof result.current.handleNewDashboard).toBe('function');
      expect(typeof result.current.handleSetHomeId).toBe('function');
      expect(typeof result.current.refetchDashboards).toBe('function');
      expect(typeof result.current.saveDashboard).toBe('function');
      expect(typeof result.current.updateDashboard).toBe('function');
    });

    it('should delegate saveDashboard and updateDashboard to store actions', async () => {
      const { result } = renderHook(() => useDashboardActions());

      await act(async () => {
        await result.current.saveDashboard();
      });

      act(() => {
        result.current.updateDashboard({ title: 'Updated from actions' } as any);
      });

      expect(mockSaveDashboard).toHaveBeenCalled();
      expect(mockUpdateDashboardStore).toHaveBeenCalledWith({ title: 'Updated from actions' });
    });
  });

  describe('Edge cases', () => {
    beforeEach(() => {
      jest.clearAllMocks();
      // Reset mocks for edge case tests
      mockUseNavigate.mockReturnValue(mockNavigate);
      mockUseToast.mockReturnValue({
        toast: mockToast,
        dismiss: jest.fn(),
        toasts: [],
      });
      mockUseFavoritesStore.mockReturnValue({
        addFavorite: mockAddFavorite,
        removeFavorite: mockRemoveFavorite,
        setFavorites: mockSetFavorites,
      });
      mockUseDashboardNavStore.getState.mockReturnValue({
        loadNav: mockLoadNav,
      });
      mockUseDashboardStore.mockReturnValue({
        deleteDashboard: mockDeleteDashboard,
        createDashboard: mockCreateDashboard,
        updateFavorite: mockUpdateFavoriteStore,
        setHome: mockSetHome,
        saveDashboard: mockSaveDashboard,
        updateDashboard: mockUpdateDashboardStore,
      });
    });

    it('should handle undefined settingData', () => {
      const baseListResult = {
        data: undefined,
        refetch: jest.fn(),
        error: undefined,
        isError: false,
        isLoading: false,
        isLoadingError: false,
        isSuccess: true,
        isFetched: true,
        isRefetching: false,
        status: 'success',
        isRefetchError: false,
        dataUpdatedAt: 0,
        errorUpdatedAt: 0,
        failureCount: 0,
        failureReason: undefined,
        isFetchedAfterMount: true,
        isIdle: false,
        isPaused: false,
        isPlaceholderData: false,
        isPreviousData: false,
        isStale: false,
        remove: jest.fn(),
        fetchStatus: 'idle',
        errorUpdateCount: 0,
        isFetching: false,
        isInitialLoading: false,
      };
      mockUseList.mockImplementation((params) => {
        if (params && params.resource === 'dashboard') {
          return {
            query: {
              ...baseListResult,
              data: mockDashboardsData,
              refetch: mockRefetchDashboards,
            },
          } as any;
        } else if (params && params.resource === 'home') {
          return {
            query: {
              ...baseListResult,
              data: null,
              refetch: mockRefetchSettings,
            },
          } as any;
        }
        return {
          query: {
            ...baseListResult,
            data: null,
            refetch: jest.fn(),
          },
        } as any;
      });

      const { result } = renderHook(() => useDashboardActions());

      expect(result.current.homeId).toBe('');
    });

    it('should handle missing menu-home in settingData', () => {
      const baseListResult = {
        data: undefined,
        refetch: jest.fn(),
        error: undefined,
        isError: false,
        isLoading: false,
        isLoadingError: false,
        isSuccess: true,
        isFetched: true,
        isRefetching: false,
        status: 'success',
        isRefetchError: false,
        dataUpdatedAt: 0,
        errorUpdatedAt: 0,
        failureCount: 0,
        failureReason: undefined,
        isFetchedAfterMount: true,
        isIdle: false,
        isPaused: false,
        isPlaceholderData: false,
        isPreviousData: false,
        isStale: false,
        remove: jest.fn(),
        fetchStatus: 'idle',
        errorUpdateCount: 0,
        isFetching: false,
        isInitialLoading: false,
      };
      mockUseList.mockImplementation((params) => {
        if (params && params.resource === 'dashboard') {
          return {
            query: {
              ...baseListResult,
              data: mockDashboardsData,
              refetch: mockRefetchDashboards,
            },
          } as any;
        } else if (params && params.resource === 'home') {
          return {
            query: {
              ...baseListResult,
              data: { data: { 'other-setting': 'value' } },
              refetch: mockRefetchSettings,
            },
          } as any;
        }
        return {
          query: {
            ...baseListResult,
            data: null,
            refetch: jest.fn(),
          },
        } as any;
      });

      const { result } = renderHook(() => useDashboardActions());

      expect(result.current.homeId).toBe('');
    });

    it('should handle empty dashboard data', () => {
      const baseListResult = {
        data: undefined,
        refetch: jest.fn(),
        error: undefined,
        isError: false,
        isLoading: false,
        isLoadingError: false,
        isSuccess: true,
        isFetched: true,
        isRefetching: false,
        status: 'success',
        isRefetchError: false,
        dataUpdatedAt: 0,
        errorUpdatedAt: 0,
        failureCount: 0,
        failureReason: undefined,
        isFetchedAfterMount: true,
        isIdle: false,
        isPaused: false,
        isPlaceholderData: false,
        isPreviousData: false,
        isStale: false,
        remove: jest.fn(),
        fetchStatus: 'idle',
        errorUpdateCount: 0,
        isFetching: false,
        isInitialLoading: false,
      };
      mockUseList.mockImplementation((params) => {
        if (params && params.resource === 'dashboard') {
          return {
            query: {
              ...baseListResult,
              data: { data: [] },
              refetch: mockRefetchDashboards,
            },
          } as any;
        } else if (params && params.resource === 'home') {
          return {
            query: {
              ...baseListResult,
              data: mockSettingData,
              refetch: mockRefetchSettings,
            },
          } as any;
        }
        return {
          query: {
            ...baseListResult,
            data: null,
            refetch: jest.fn(),
          },
        } as any;
      });

      const { result } = renderHook(() => useDashboardActions());

      expect(result.current.dashboardsData).toEqual({ data: [] });
    });
  });
});
