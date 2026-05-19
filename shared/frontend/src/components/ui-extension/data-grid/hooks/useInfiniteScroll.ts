/**
 * useInfiniteScroll
 *
 * IntersectionObserver 기반 무한 스크롤 hook
 * - 스크롤 하단 sentinel 요소 감지
 * - 자동으로 fetchNextPage 호출
 */

import { useEffect, useRef, useCallback } from 'react';

export interface UseInfiniteScrollOptions {
  /** 다음 페이지 존재 여부 */
  hasNextPage: boolean;
  /** 현재 다음 페이지 로딩 중 여부 */
  isFetchingNextPage: boolean;
  /** 다음 페이지 로드 함수 */
  fetchNextPage: () => void;
  /** IntersectionObserver rootMargin (기본: '100px') */
  rootMargin?: string;
  /** 활성화 여부 (기본: true) */
  enabled?: boolean;
}

export interface UseInfiniteScrollReturn {
  /** sentinel 요소에 연결할 ref */
  loadMoreRef: React.RefObject<HTMLDivElement | null>;
}

export function useInfiniteScroll({
  hasNextPage,
  isFetchingNextPage,
  fetchNextPage,
  rootMargin = '100px',
  enabled = true,
}: UseInfiniteScrollOptions): UseInfiniteScrollReturn {
  const loadMoreRef = useRef<HTMLDivElement>(null);

  const handleIntersect = useCallback(
    (entries: IntersectionObserverEntry[]) => {
      const [entry] = entries;

      // 조건: 화면에 보임 && 다음 페이지 있음 && 로딩 중 아님
      if (entry.isIntersecting && hasNextPage && !isFetchingNextPage) {
        fetchNextPage();
      }
    },
    [hasNextPage, isFetchingNextPage, fetchNextPage]
  );

  useEffect(() => {
    const element = loadMoreRef.current;

    // 비활성화 또는 element 없으면 스킵
    if (!enabled || !element) return;

    // 다음 페이지 없으면 observer 불필요
    if (!hasNextPage) return;

    const observer = new IntersectionObserver(handleIntersect, {
      root: null, // viewport 기준
      rootMargin,
      threshold: 0,
    });

    observer.observe(element);

    return () => {
      observer.disconnect();
    };
  }, [enabled, hasNextPage, rootMargin, handleIntersect]);

  return { loadMoreRef };
}
