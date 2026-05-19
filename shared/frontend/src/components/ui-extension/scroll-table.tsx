import * as React from 'react';
import * as ScrollAreaPrimitive from '@radix-ui/react-scroll-area';

import {cn} from '../../lib';

/**
 * ScrollTable - DataTable 전용 테이블 컴포넌트
 *
 * 왜 이 컴포넌트가 필요한가?
 * 1. shadcn Table은 <div wrapper + overflow-x-auto>를 포함해서 ScrollArea와 이중 스크롤 충돌
 * 2. DataTable은 sticky header, column resizing을 위해 border-separate 필요
 * 3. ui/table.tsx는 shadcn 원본이므로 업데이트 추적을 위해 수정 금지
 *
 * asViewport 옵션:
 * - ScrollArea의 Viewport와 table을 통합해서 불필요한 wrapper 제거
 * - DataTable이 ScrollAreaPrimitive.Root와 함께 사용할 때만 true로 설정
 * - 일반 사용 시에는 false (기본값) - 순수 table 태그로 동작
 */
interface ScrollTableProps extends React.HTMLAttributes<HTMLTableElement> {
  /**
   * ScrollArea의 Viewport로 감쌀지 여부
   * true: ScrollAreaPrimitive.Viewport로 감싸서 스크롤 영역으로 사용
   * false (기본): 일반 table 태그로 사용
   */
  asViewport?: boolean;
}

const ScrollTable = React.forwardRef<HTMLTableElement, ScrollTableProps>(
  ({asViewport = false, className, ...props}, ref) => {
    const table = (
      <table
        ref={ref}
        className={cn('w-full text-sm border-separate border-spacing-0', className)}
        {...props}
      />
    );

    if (asViewport) {
      return (
        <ScrollAreaPrimitive.Viewport className="h-full w-full rounded-[inherit]">
          {table}
        </ScrollAreaPrimitive.Viewport>
      );
    }

    return table;
  },
);
ScrollTable.displayName = 'ScrollTable';

export default ScrollTable;
