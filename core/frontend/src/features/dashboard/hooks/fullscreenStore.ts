import { create } from 'zustand';

type FullscreenState = {
  isFullscreen: boolean;
  isSidebarVisible: boolean;
  isFilterLeftVisible: boolean;
  enterFullscreen: () => void;
  exitFullscreen: () => void;
  toggleFullscreen: () => void;
};

export const useFullscreenStore = create<FullscreenState>((set) => ({
  isFullscreen: false,
  isSidebarVisible: true,
  isFilterLeftVisible: true,

  enterFullscreen: () =>
    set({
      isFullscreen: true,
      isSidebarVisible: false,
      isFilterLeftVisible: false,
    }),

  exitFullscreen: () =>
    set({
      isFullscreen: false,
      isSidebarVisible: true,
      isFilterLeftVisible: true,
    }),

  toggleFullscreen: () =>
    set((state) => {
      const next = !state.isFullscreen;
      return {
        isFullscreen: next,
        isSidebarVisible: !next,
        isFilterLeftVisible: !next,
      };
    }),
}));
