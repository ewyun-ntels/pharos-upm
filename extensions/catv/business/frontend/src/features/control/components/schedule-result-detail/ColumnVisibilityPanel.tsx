'use client';

import type {Table} from '@tanstack/react-table';
import {SlidersHorizontal} from 'lucide-react';
import {
  Button,
  Checkbox,
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@pharos/shared/components/ui';

interface ColumnVisibilityPanelProps<T> {
  table: Table<T>;
  /** 기본 컬럼 그룹 ID 목록 (토글 가능, basicLabel 섹션으로 표시) */
  basicColumnIds?: string[];
  /** 기본 컬럼 그룹 레이블 */
  basicLabel?: string;
  /** 동적 컬럼 그룹 레이블 */
  dynamicLabel?: string;
}

export function ColumnVisibilityPanel<T>({
  table,
  basicColumnIds,
  basicLabel = '기본 컬럼',
  dynamicLabel = '정보 컬럼',
}: ColumnVisibilityPanelProps<T>) {
  const toggleableColumns = table.getAllColumns().filter((col) => col.getCanHide());

  const basicColumns = basicColumnIds
    ? toggleableColumns.filter((col) => basicColumnIds.includes(col.id))
    : [];
  const dynamicColumns = toggleableColumns.filter((col) => !basicColumnIds?.includes(col.id));

  if (dynamicColumns.length === 0 && basicColumns.length === 0) return null;

  const visibleDynamicCount = dynamicColumns.filter((col) => col.getIsVisible()).length;
  const allDynamicVisible =
    dynamicColumns.length > 0 && visibleDynamicCount === dynamicColumns.length;

  const visibleBasicCount = basicColumns.filter((col) => col.getIsVisible()).length;
  const allBasicVisible = basicColumns.length > 0 && visibleBasicCount === basicColumns.length;

  // 트리거 배지 표시: basic + dynamic 중 하나라도 숨김이 있으면 배지 표시
  const totalToggable = basicColumns.length + dynamicColumns.length;
  const totalVisible = visibleBasicCount + visibleDynamicCount;
  const someHidden = totalVisible < totalToggable;

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" size="sm" className="h-8 hidden lg:flex items-center gap-1.5">
          <SlidersHorizontal className="h-3.5 w-3.5" />
          컬럼
          {someHidden && (
            <span className="ml-0.5 rounded bg-primary px-1 py-0 text-[10px] font-semibold text-primary-foreground leading-4">
              {totalVisible}/{totalToggable}
            </span>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-130 p-4">
        <div className="space-y-3">
          {/* 헤더 */}
          <div className="flex items-center justify-between">
            <h4 className="text-sm font-semibold">컬럼 표시 설정</h4>
          </div>

          {/* 기본 컬럼 섹션 (토글 가능) */}
          {basicColumns.length > 0 && (
            <div>
              <div className="flex items-center justify-between mb-1.5">
                <p className="text-xs text-muted-foreground">
                  {basicLabel} ({visibleBasicCount}/{basicColumns.length} 선택됨)
                </p>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-6 px-2 text-xs"
                  onClick={() =>
                    basicColumns.forEach((col) => col.toggleVisibility(!allBasicVisible))
                  }
                >
                  {allBasicVisible ? '전체 해제' : '전체 선택'}
                </Button>
              </div>
              <div className="grid grid-cols-3 gap-1">
                {basicColumns.map((col) => (
                  <label
                    key={col.id}
                    className="flex cursor-pointer items-center gap-1.5 rounded px-1 py-0.5 hover:bg-muted/50"
                  >
                    <Checkbox
                      checked={col.getIsVisible()}
                      onCheckedChange={(checked) => col.toggleVisibility(!!checked)}
                      className="h-3.5 w-3.5"
                    />
                    <span className="truncate font-mono text-xs" title={col.id}>
                      {col.id}
                    </span>
                  </label>
                ))}
              </div>
            </div>
          )}

          {/* 구분선 */}
          {basicColumns.length > 0 && dynamicColumns.length > 0 && <div className="border-t" />}

          {/* 동적 컬럼 */}
          {dynamicColumns.length > 0 && (
            <div>
              <div className="flex items-center justify-between mb-1.5">
                <p className="text-xs text-muted-foreground">
                  {dynamicLabel} ({visibleDynamicCount}/{dynamicColumns.length} 선택됨)
                </p>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-6 px-2 text-xs"
                  onClick={() =>
                    dynamicColumns.forEach((col) => col.toggleVisibility(!allDynamicVisible))
                  }
                >
                  {allDynamicVisible ? '전체 해제' : '전체 선택'}
                </Button>
              </div>
              <div className="max-h-64 overflow-y-auto pr-1">
                <div className="grid grid-cols-3 gap-1">
                  {dynamicColumns.map((col) => (
                    <label
                      key={col.id}
                      className="flex cursor-pointer items-center gap-1.5 rounded px-1 py-0.5 hover:bg-muted/50"
                    >
                      <Checkbox
                        checked={col.getIsVisible()}
                        onCheckedChange={(checked) => col.toggleVisibility(!!checked)}
                        className="h-3.5 w-3.5"
                      />
                      <span className="truncate font-mono text-xs" title={col.id}>
                        {col.id}
                      </span>
                    </label>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}
