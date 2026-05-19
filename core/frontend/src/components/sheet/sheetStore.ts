import React from 'react';
import {create} from 'zustand';
import {defaultSheetXSize} from '../layout/layout';

type SheetStore = {
  sheetStopXSize: number;
  sheet: React.JSX.Element | undefined;
  setSheet: (newSheet: React.JSX.Element) => void;
  setLayout: (v: number) => void;
  setClose: () => void;
};

const useSheetStore = create<SheetStore>()((set) => ({
  sheetStopXSize: 0, // 0일경우 sheet가 disable인 것으로 가정
  sheet: undefined,
  setSheet: (newSheet: React.JSX.Element) =>
    set((state) => {
      state.sheet = newSheet;
      state.sheetStopXSize = defaultSheetXSize;
      return {};
    }),
  setLayout: (size) => set(() => ({sheetStopXSize: size})),
  setClose: () => set(() => ({sheetStopXSize: 0})),
}));

export {useSheetStore};
export type {SheetStore};
