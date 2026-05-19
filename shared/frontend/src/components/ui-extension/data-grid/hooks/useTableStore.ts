import {create} from 'zustand';

interface TableStorePaginationState {
  pageIndex: number;
  pageSize: number;
}

export interface TableStoreState {
  pagination: TableStorePaginationState;
  searchTerm: string;
}

type TableState = TableStoreState;

type ResetOptions = {
  pagination?: boolean | 'keepPageSize';
  searchTerm?: boolean;
  all?: boolean;
};

interface TableStore {
  tableStates: Record<string, TableState>;
  registeredKeys: Set<string>;
  getTableState: (menuKey: string) => TableState;
  updateTableState: (menuKey: string, updates: Partial<TableState>) => void;
  resetTableState: (menuKey: string, options?: ResetOptions) => void;
  resetAllTableStates: () => void;
  registerTableKey: (menuKey: string) => void;
  unregisterTableKey: (menuKey: string) => void;
}

const defaultPagination: TableStorePaginationState = {
  pageIndex: 0,
  pageSize: 10,
};

const defaultTableState: TableState = {
  pagination: defaultPagination,
  searchTerm: '',
};

/**
 * DataGrid 전용 전역 상태 관리 Store
 * 
 * 여러 DataGrid 인스턴스의 상태(pagination, searchTerm)를 독립적으로 관리합니다.
 * tableKey를 사용하여 각 DataGrid의 상태를 격리합니다.
 * 
 * @example
 * ```tsx
 * const { getTableState, updateTableState } = useTableStore();
 * const state = getTableState('menu-users');
 * updateTableState('menu-users', { searchTerm: 'John' });
 * ```
 */
export const useTableStore = create<TableStore>((set, get) => ({
  tableStates: {},
  registeredKeys: new Set<string>(),

  registerTableKey: (menuKey: string) => {
    const { registeredKeys } = get();
    
    if (process.env.NODE_ENV === 'development') {
      if (registeredKeys.has(menuKey)) {
        console.warn(
          `🔴 [TableStore] Duplicate tableKey detected: "${menuKey}"\n` +
          `Multiple DataGrid instances are using the same tableKey.\n` +
          `This will cause state conflicts. Please use unique tableKey for each table.`
        );
      }
    }
    
    set((state) => ({
      registeredKeys: new Set(state.registeredKeys).add(menuKey),
    }));
  },

  unregisterTableKey: (menuKey: string) => {
    set((state) => {
      const newKeys = new Set(state.registeredKeys);
      newKeys.delete(menuKey);
      return { registeredKeys: newKeys };
    });
  },

  getTableState: (menuKey: string) => {
    const state = get();
    return state.tableStates[menuKey] || defaultTableState;
  },

  updateTableState: (menuKey: string, updates: Partial<TableState>) => {
    set((state) => {
      const currentTableState = state.tableStates[menuKey] || defaultTableState;
      return {
        tableStates: {
          ...state.tableStates,
          [menuKey]: {
            ...currentTableState,
            ...updates,
          },
        },
      };
    });
  },

  resetTableState: (menuKey: string, options: ResetOptions = {}) => {
    const {pagination, searchTerm, all} = options;

    // 특정 메뉴의 전체 상태를 기본값으로 리셋
    if (all) {
      get().updateTableState(menuKey, defaultTableState);
      return;
    }

    const updates: Partial<TableState> = {};

    if (pagination === true) {
      updates.pagination = defaultPagination;
    } else if (pagination === 'keepPageSize') {
      const currentState = get().getTableState(menuKey);
      updates.pagination = {
        pageIndex: 0,
        pageSize: currentState.pagination.pageSize,
      };
    }

    if (searchTerm) {
      updates.searchTerm = '';
    }

    if (Object.keys(updates).length > 0) {
      get().updateTableState(menuKey, updates);
    }
  },

  // 모든 메뉴의 테이블 상태를 제거 (전역 초기화)
  resetAllTableStates: () => {
    set({tableStates: {}});
  },
}));
