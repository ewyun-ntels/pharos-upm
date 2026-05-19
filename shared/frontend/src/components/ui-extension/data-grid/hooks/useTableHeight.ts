/**
 * useTableHeight Hook
 *
 * 테이블 높이 계산 로직
 */

import { RefObject, useState, useEffect, useCallback } from 'react';
import { TABLE_TD_HEIGHT } from '@pharos/shared/components/ui-extension';
import { useWindowDimensions } from '../../../../lib/window-dimensions';

interface UseTableHeightProps {
  tableHeight?: string | number;
  useData?: boolean;
  dataLength: number;
  tableRef: RefObject<HTMLDivElement | null>;
  filterRef: RefObject<HTMLDivElement | null>;
  paginationRef: RefObject<HTMLDivElement | null>;
  tableBodyRef: RefObject<HTMLTableSectionElement | null>;
  /** isLoading 상태 변화 시 re-observe 트리거용 (isLoading false → DataGridTableView 마운트 후 height 재계산) */
  isLoading?: boolean;
}

export function useTableHeight({
  tableHeight: propsTableHeight,
  useData = false,
  dataLength,
  tableRef,
  filterRef,
  paginationRef,
  tableBodyRef,
  isLoading,
}: UseTableHeightProps): string | undefined {
  const { height: windowHeight } = useWindowDimensions();
  // 초기값을 최소 높이로 설정 (header + 1 row)
  const [tableHeight, setTableHeight] = useState<number>(TABLE_TD_HEIGHT * 2);

  // 높이 재계산 함수
  const calculateHeight = useCallback(() => {
    if (!tableRef.current || !tableBodyRef.current) return;

    // 실제 tbody의 scrollHeight만 사용 (row 개수나 wrap 여부와 무관하게 실제 높이)
    let actualRowHeight = tableBodyRef.current.scrollHeight;

    // 스크롤 컨테이너 찾기 (tbody -> table -> scroll div)
    const scrollContainer = tableBodyRef.current.parentElement?.parentElement;

    // 가로 스크롤바 높이 측정 (Windows: ~15-17px, Mac overlay: 0px)
    const hxScrollbarHeight = scrollContainer
      ? scrollContainer.offsetHeight - scrollContainer.clientHeight
      : 0;

    // 실제 thead 높이 측정 (TABLE_TD_HEIGHT 상수 대신 DOM에서 직접 측정하여 오차 제거)
    // grouping header 등 행이 여러 개여도 정확히 처리됨
    const theadEl = tableBodyRef.current.previousElementSibling;
    const headerHeight = theadEl
      ? Math.ceil(theadEl.getBoundingClientRect().height)
      : TABLE_TD_HEIGHT;

    // header + tbody + 실제 가로 스크롤바 높이
    actualRowHeight = headerHeight + actualRowHeight + hxScrollbarHeight;

    const tableTop = tableRef.current.getBoundingClientRect().top || 0;
    const availableHeight = windowHeight - 120 - tableTop; // 120px offset

    // data가 없을 때(No results)는 availableHeight 사용 — tbody scrollHeight이 1줄이라 작게 계산되는 버그 방지
    // data가 있을 때는 actualRowHeight와 availableHeight 중 작은 값 사용
    const result = dataLength === 0 || availableHeight <= actualRowHeight
      ? availableHeight
      : actualRowHeight;
    setTableHeight(result > 0 ? result : TABLE_TD_HEIGHT * 2);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [windowHeight, tableRef, tableBodyRef, dataLength, isLoading]);

  // 윈도우 리사이즈 시 재계산
  useEffect(() => {
    calculateHeight();
  }, [calculateHeight]);

  // tbody DOM이 실제로 변경되거나 크기가 변할 때 재계산
  useEffect(() => {
    const tbody = tableBodyRef.current;
    if (!tbody) return;

    // 초기 계산
    calculateHeight();

    // tbody 크기 변화 감지 (row wrap, 데이터 변경 등)
    const resizeObserver = new ResizeObserver(() => {
      calculateHeight();
    });

    resizeObserver.observe(tbody);
    return () => resizeObserver.disconnect();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [calculateHeight]);

  if (propsTableHeight === 'flex') {
    return 'flex';
  }

  if (!propsTableHeight) {
    return `${tableHeight}px`;
  }

  // 문자열 높이 처리
  if (typeof propsTableHeight === 'string') {
    // CSS 값은 그대로 반환
    if (propsTableHeight.includes('%') ||
        propsTableHeight.includes('px') ||
        propsTableHeight.includes('vh')) {
      return propsTableHeight;
    }

    // 숫자 문자열 파싱
    const heightValue = parseInt(propsTableHeight);
    if (isNaN(heightValue)) return propsTableHeight;

    return useData
      ? `calc(${heightValue}px - 100px)` // Dashboard Panel: Filter + Pagination 공간
      : dataLength !== 0 ? `${heightValue}px` : undefined;
  }

  // 숫자 높이 처리 (동적 계산)
  if (useData) {
    const filterHeight = filterRef.current?.getBoundingClientRect().height || 0;
    const paginationHeight = paginationRef.current?.getBoundingClientRect().height || 0;
    const calculated = propsTableHeight - (filterHeight + paginationHeight + 16); // 8px margin top + 8px bottom padding
    return calculated > 0 ? `${calculated}px` : '100px';
  }

  return dataLength !== 0 ? `${propsTableHeight}px` : undefined;
}
