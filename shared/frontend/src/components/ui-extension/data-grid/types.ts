/**
 * DataGrid Types
 *
 * 통합 테이블 컴포넌트 타입 정의
 */

import {ReactNode} from 'react';
import {
  ColumnDef,
  Cell,
  VisibilityState,
  SortingState,
  FilterFn,
  Table,
} from '@tanstack/react-table';

/**
 * Data fetching hook의 반환 타입
 * useList (Refine), useQuery (React Query), fetch 등 어떤 방식이든 이 형태로 변환
 */
export interface UseDataResult<T = unknown> {
  data: T[];
  isLoading: boolean;
  error?: Error;
  refetch?: () => void;
}

/**
 * 체크박스 설정
 */
export interface CheckboxConfig<T = unknown> {
  checkboxKey?: string;
  selectedItems?: string[];
  currentPageData?: T[];
  handleSelectAll?: (checked: boolean, selectableItems: T[]) => void;
  handleSelectOne?: (item: T, checked: boolean) => void;
  isSelectable?: boolean | ((row: T) => boolean);
}

/**
 * Export 설정 (Core의 완전한 버전)
 */
export interface ExportConfig<T = unknown> {
  fileName?: string;
  mapData?: (row: T) => Record<string, unknown>;
  PdfButtonSize?: 'default' | 'sm' | 'lg' | 'icon' | null;
}

/**
 * Pagination 설정
 */
export interface PaginationConfig {
  pageSize?: number;
  initialPageIndex?: number;
  /** 현재 페이지 인덱스 (0-based). 외부에서 변경 시 DataGrid 내부 상태를 동기화합니다. */
  pageIndex?: number;
}

/**
 * DataGrid Props (통합 - 기술부채 제거 완료)
 */
export interface DataGridProps<T = unknown> {
  // ============================================================================
  // 데이터 소스
  // ============================================================================
  useData?: () => UseDataResult<T>;
  data?: T[];

  // ============================================================================
  // 테이블 설정
  // ============================================================================
  columns: ColumnDef<T>[];
  getRowId?: (row: T, index: number) => string;
  getSubRows?: (row: T) => T[] | undefined;

  /**
   * 전역 상태 저장을 위한 고유 키
   * persistState=true일 때 필수
   */
  tableKey?: string;

  /**
   * 전역 상태 저장 활성화 여부 (pagination, searchTerm)
   * true일 경우 tableKey 필수, 없으면 에러 발생
   * false 또는 미설정 시 로컬 상태만 사용
   */
  persistState?: boolean;

  // ============================================================================
  // 필터 (모든 모드에서 사용, table 객체 필요 없으면 무시)
  // ============================================================================
  leftFilters?: (table: Table<T>) => ReactNode[];
  rightFilters?: (table: Table<T>) => ReactNode[];
  columnFilters?: Array<{id: string; value: unknown}>;

  // ============================================================================
  // 검색 & 새로고침
  // ============================================================================
  onRefresh?: () => void;
  isRefreshing?: boolean;

  // ============================================================================
  // 기능 설정
  // ============================================================================
  enablePagination?: boolean;
  paginationConfig?: PaginationConfig;
  pageSizes?: number[];
  paginationVariant?: 'default' | 'numbered';

  /** 서버 사이드 pagination을 위한 전체 데이터 수 */
  totalCount?: number;

  enableColumnResizing?: boolean;
  enableExpandingColumn?: boolean;
  /** Expand toggle을 주입할 컬럼 ID */
  treeColumnId?: string;
  useTableSorting?: boolean;
  initialSorting?: SortingState;

  // ============================================================================
  // 체크박스 & Export
  // ============================================================================

  /** 체크박스 설정 */
  checkboxConfig?: CheckboxConfig<T>;

  /** Export 설정 */
  exportConfig?: ExportConfig<T>;

  // ============================================================================
  // UI 설정
  // ============================================================================

  /** 테이블 높이 (숫자 또는 CSS 문자열) */
  tableHeight?: string | number | 'flex';

  /** 테이블 제목 */
  title?: string;

  /** 표시 이름 */
  displayName?: string;

  /** 테이블 설명 (필터 영역에 표시) */
  description?: string;

  /** 검색 가능 컬럼 */
  searchableColumns?: string[];

  /** 빈 메시지 */
  emptyCustomMessage?: string;

  /** 컬럼 visibility */
  visibilityState?: VisibilityState;

  /** Row cursor */
  rowCursor?: boolean;

  /**
   * 테이블 너비 처리 방식
   * - 'independent': 각 컬럼 독립 px. 컬럼 줄이면 오른쪽 빈공간 발생. 데이터 그리드 표준.
   * - 'spacer': 각 컬럼 독립 px + 끝에 spacer 컬럼으로 빈공간 채움. 목록형 UI 표준. (기본값)
   * - 'fill-last': 마지막 컬럼이 나머지 공간을 모두 차지.
   */
  tableWidthMode?: 'independent' | 'spacer' | 'fill-last';

  /** 클래스명 */
  className?: string;

  /** 로딩 상태 */
  isLoading?: boolean;

  // ============================================================================
  // 이벤트 핸들러
  // ============================================================================

  /** Row 클릭 핸들러 */
  onRowClick?: (row: T) => void;

  /** Cell 클릭 핸들러 */
  dataCellOnClick?: (cell: Cell<T, unknown>) => void;

  /** 데이터 로드 완료 핸들러 */
  onDataLoaded?: (data: T[]) => void;

  /** 테이블 준비 완료 핸들러 */
  onTableReady?: (table: Table<T>) => void;

  /** 페이지 변경 핸들러 */
  onPageChange?: (page: number) => void;

  /** 페이지 사이즈 변경 핸들러 */
  onPageSizeChange?: (pageSize: number) => void;

  /** 현재 페이지 데이터 변경 핸들러 */
  onCurrentPageDataChange?: (data: T[]) => void;

  /**
   * 콜럼 리사이즈 완료 핸들러
   * 사용자가 콜럼을 드래그하여 리사이즈할 때 호출 (debounced).
   * key: column id, value: px 너비
   */
  onColumnSizingChange?: (sizing: Record<string, number>) => void;

  // ============================================================================
  // 고급 설정
  // ============================================================================

  /** Row 클릭 제외 컬럼 */
  excludeRowClickColumns?: string[];

  /** Global filter 함수 */
  globalFilterFn?: FilterFn<T>;

  /** Children (Pagination 등 추가 컨텐츠) */
  children?: ReactNode | ((table: Table<T>) => ReactNode);
}

// ============================================================================
// DataGridInfinity Types
// ============================================================================

/**
 * 무한 스크롤 설정
 */
export interface InfiniteScrollConfig {
  /** 다음 페이지 존재 여부 */
  hasNextPage: boolean;
  /** 다음 페이지 로딩 중 여부 */
  isFetchingNextPage: boolean;
  /** 다음 페이지 로드 함수 */
  fetchNextPage: () => void;
  /** IntersectionObserver rootMargin */
  rootMargin?: string;
}

/**
 * DataGridInfinity Props
 * - 페이지네이션 관련 props 제거
 * - 무한 스크롤 관련 props 추가
 */
export interface DataGridInfinityProps<T = unknown> extends Omit<
  DataGridProps<T>,
  | 'enablePagination'
  | 'paginationConfig'
  | 'pageSizes'
  | 'onPageChange'
  | 'onPageSizeChange'
  | 'children'
> {
  // ============================================================================
  // 무한 스크롤 (필수)
  // ============================================================================

  /** 다음 페이지 존재 여부 */
  hasNextPage: boolean;

  /** 다음 페이지 로딩 중 여부 */
  isFetchingNextPage: boolean;

  /** 다음 페이지 로드 함수 */
  fetchNextPage: () => void;

  /** IntersectionObserver rootMargin (기본: '100px') */
  rootMargin?: string;
}
