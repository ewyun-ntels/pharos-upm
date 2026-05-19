import {create} from 'zustand';
import {dashboardProvider, DASHBOARD_RESOURCES} from '@providers/dashboard-provider';
import {registerOnLogout} from '@pharos/shared/features/auth';

export interface FavoriteItem {
  id: string;
  displayName: string;
}

interface FavoritesStore {
  favorites: FavoriteItem[];
  setFavorites: (favorites: FavoriteItem[]) => void;
  addFavorite: (item: FavoriteItem) => void;
  removeFavorite: (id: string) => void;
  isFavorite: (id: string) => boolean;
  loadFavorites: () => Promise<void>;
}

export const useFavoritesStore = create<FavoritesStore>()((set, get) => ({
  favorites: [],

  setFavorites: (favorites) => set({favorites}),

  addFavorite: (item) =>
    set((state) => ({
      favorites: state.favorites.some((f) => f.id === item.id)
        ? state.favorites
            .map((f) => (f.id === item.id ? item : f))
            .sort((a, b) => a.displayName.localeCompare(b.displayName))
        : [...state.favorites, item].sort((a, b) =>
            a.displayName.localeCompare(b.displayName),
          ),
    })),

  removeFavorite: (id) =>
    set((state) => ({
      favorites: state.favorites.filter((f) => f.id !== id),
    })),

  isFavorite: (id) => get().favorites.some((f) => f.id === id),

  loadFavorites: async () => {
    try {
      const result = await dashboardProvider.getList({
        resource: DASHBOARD_RESOURCES.FAVORITES,
        pagination: undefined,
      });
      set({favorites: result.data as FavoriteItem[]});
    } catch (error) {
      console.error('[FavoritesStore] Failed to load favorites:', error);
    }
  },
}));

// 로그아웃 시 favorites store 초기화
registerOnLogout(() => {
  useFavoritesStore.getState().setFavorites([]);
});
