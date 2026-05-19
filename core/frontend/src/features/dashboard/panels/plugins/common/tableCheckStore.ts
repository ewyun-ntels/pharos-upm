import {create} from 'zustand';

type TableCheckStore = {
  regId: string;
  set: (id: string) => void;
};

const useTableCheckStore = create<TableCheckStore>()((set) => ({
  regId: '',
  set: (id) => set(() => ({regId: id})),
}));

export {useTableCheckStore};
