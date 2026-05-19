import {Dispatch, RefObject, SetStateAction, useEffect, useRef, useState} from 'react';
import {useInfiniteScroll} from '@hooks/use-infinite-scroll';

interface UseClientInfiniteScrollProps<T> {
  allData: T[];
  pageSize: number;
  dependencies?: any[];
  isFetching?: boolean;
}

interface UseClientInfiniteScrollReturn<T> {
  displayData: T[];
  loadMoreRef: RefObject<HTMLDivElement>;
  dataCache: RefObject<T[]>;
  currentPage: number;
  setCurrentPage: Dispatch<SetStateAction<number>>;
  totalItems: number;
}

export function UseClientInfiniteScroll<T>({
  allData,
  pageSize,
  dependencies = [],
  isFetching = false,
}: UseClientInfiniteScrollProps<T>): UseClientInfiniteScrollReturn<T> {
  const [currentPage, setCurrentPage] = useState(0);
  const [displayData, setDisplayData] = useState<T[]>([]);
  const dataCache = useRef<T[]>([]);
  const isInitialLoad = useRef(true);

  useEffect(() => {
    setCurrentPage(0);
    isInitialLoad.current = true;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...dependencies]);

  useEffect(() => {
    if (allData.length > 0) {
      dataCache.current = allData;

      if (isInitialLoad.current) {
        setDisplayData(allData.slice(0, pageSize));
        isInitialLoad.current = false;
      } else {
        setDisplayData(allData.slice(0, (currentPage + 1) * pageSize));
      }
    }
  }, [allData, currentPage, pageSize]);

  const loadNextBatch = () => {
    if ((currentPage + 1) * pageSize < dataCache.current.length) setCurrentPage((prev) => prev + 1);
  };

  const {loadMoreRef} = useInfiniteScroll({
    hasNextPage: (currentPage + 1) * pageSize < dataCache.current.length,
    isFetching,
    fetchNextPage: loadNextBatch,
  });

  return {
    displayData,
    loadMoreRef,
    dataCache,
    currentPage,
    setCurrentPage,
    totalItems: dataCache.current.length,
  };
}
