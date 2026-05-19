import {useRef, useCallback, useEffect} from 'react';

interface UseInfiniteScrollOptions {
  hasNextPage?: boolean;
  isFetching?: boolean;
  fetchNextPage: () => void;
  threshold?: number;
  rootMargin?: string;
}

interface UseInfiniteScrollReturn {
  loadMoreRef: React.RefObject<HTMLDivElement>;
}

export function useInfiniteScroll({
  hasNextPage = false,
  isFetching = false,
  fetchNextPage,
  threshold = 0.1,
  rootMargin = '0px',
}: UseInfiniteScrollOptions): UseInfiniteScrollReturn {
  const loadMoreRef = useRef<HTMLDivElement>(null!);

  const handleObserver = useCallback(
    (entries: IntersectionObserverEntry[]) => {
      const [target] = entries;
      if (target.isIntersecting && hasNextPage && !isFetching) {
        fetchNextPage();
      }
    },
    [fetchNextPage, hasNextPage, isFetching],
  );

  useEffect(() => {
    const element = loadMoreRef.current;
    if (!element) return;

    const option = {
      root: null,
      rootMargin,
      threshold,
    };

    const observer = new IntersectionObserver(handleObserver, option);
    observer.observe(element);

    return () => observer.disconnect();
  }, [handleObserver, rootMargin, threshold]);

  return {
    loadMoreRef,
  };
}
