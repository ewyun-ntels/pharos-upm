import { useCallback, useEffect, useRef, useState } from 'react';
import { ColumnDef, ColumnSizingState } from '@tanstack/react-table';

interface UseFlexColumnSizingOptions<T> {
  columns: ColumnDef<T>[];
  columnVisibility?: Record<string, boolean>;
  containerRef: React.RefObject<HTMLElement | null>;
  isLoading?: boolean;
}

interface UseFlexColumnSizingResult {
  columnSizing: Record<string, number>;
  handleColumnSizingChange: (updater: ColumnSizingState | ((old: ColumnSizingState) => ColumnSizingState)) => void;
}

export function useFlexColumnSizing<T>({
  columns,
  columnVisibility = {},
  containerRef,
  isLoading,
}: UseFlexColumnSizingOptions<T>): UseFlexColumnSizingResult {
  const [columnSizing, setColumnSizing] = useState<Record<string, number>>({});
  const userResizedColumns = useRef<Set<string>>(new Set());

  // refs로 최신값 유지 — calculate를 안정적으로 유지하기 위함
  const columnsRef = useRef(columns);
  columnsRef.current = columns;

  const columnVisibilityRef = useRef(columnVisibility);
  columnVisibilityRef.current = columnVisibility;

  // deps 없이 안정적인 함수 — 항상 ref에서 최신값을 읽음
  const calculate = useCallback((containerWidth: number) => {
    const cols = columnsRef.current;
    const visibility = columnVisibilityRef.current;

    const visibleCols = cols.filter((col) => visibility[col.id as string] !== false);
    const flexCols = visibleCols.filter((col) => col.meta?.flex !== undefined && col.id);
    if (flexCols.length === 0) return;

    const fixedWidth = visibleCols
      .filter((col) => col.meta?.flex === undefined)
      .reduce((sum, col) => sum + (typeof col.size === 'number' ? col.size : 150), 0);

    const remaining = Math.max(containerWidth - fixedWidth, 0);
    const totalFlex = flexCols.reduce((sum, col) => sum + (col.meta?.flex ?? 1), 0);

    setColumnSizing((prev) => {
      const next = { ...prev };
      let changed = false;
      flexCols.forEach((col) => {
        const id = col.id as string;
        if (!userResizedColumns.current.has(id)) {
          const flex = col.meta?.flex ?? 1;
          const minSize = typeof col.minSize === 'number' ? col.minSize : 100;
          const newSize = Math.max(Math.floor((remaining * flex) / totalFlex), minSize);
          if (prev[id] !== newSize) {
            next[id] = newSize;
            changed = true;
          }
        }
      });
      // 실제로 변경이 없으면 동일 참조 반환 → 불필요한 재렌더 방지
      return changed ? next : prev;
    });
  }, []); // 빈 deps — refs로 값 읽으므로 재생성 불필요

  // 마운트 시 + 컨테이너 크기 변화 시 flex 컬럼 재계산
  useEffect(() => {
    if (!containerRef.current) return;
    calculate(containerRef.current.offsetWidth);

    const observer = new ResizeObserver((entries) => {
      const width = entries[0]?.contentRect.width;
      if (width) calculate(width);
    });

    observer.observe(containerRef.current);
    return () => observer.disconnect();
  }, [calculate, containerRef]);

  // visibility 변경 시 flex 재계산 — JSON.stringify로 실제 변경 여부 판단
  const prevVisibilityKeyRef = useRef('');
  useEffect(() => {
    const key = JSON.stringify(columnVisibility);
    if (key === prevVisibilityKeyRef.current) return;
    prevVisibilityKeyRef.current = key;

    if (!containerRef.current) return;
    const rafId = requestAnimationFrame(() => {
      if (containerRef.current) calculate(containerRef.current.offsetWidth);
    });
    return () => cancelAnimationFrame(rafId);
  }, [columnVisibility, calculate, containerRef]);

  // 데이터 로딩 완료 후 flex 재계산 (isLoading true → false 전환 시점 감지)
  const prevLoadingRef = useRef(isLoading);
  useEffect(() => {
    const wasLoading = prevLoadingRef.current;
    prevLoadingRef.current = isLoading;

    if (!wasLoading || isLoading || !containerRef.current) return;

    const rafId = requestAnimationFrame(() => {
      if (containerRef.current) calculate(containerRef.current.offsetWidth);
    });
    return () => cancelAnimationFrame(rafId);
  }, [isLoading, calculate, containerRef]);

  const handleColumnSizingChange = useCallback(
    (updater: ColumnSizingState | ((old: ColumnSizingState) => ColumnSizingState)) => {
      setColumnSizing((prev) => {
        const next = typeof updater === 'function' ? updater(prev) : updater;
        Object.keys(next).forEach((id) => {
          if (prev[id] !== next[id]) userResizedColumns.current.add(id);
        });
        return next;
      });
    },
    [],
  );

  return { columnSizing, handleColumnSizingChange };
}
