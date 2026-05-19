/**
 * DropdownMenu 기반 DataTable Row Menu (간단 버전)
 * 
 * 기존 dropdown-menu 컴포넌트를 활용한 간단한 wrapper
 * - Radix UI의 Portal, 외부 클릭 처리 자동
 * - 테이블 리렌더 시에도 메뉴 유지
 */

import * as React from 'react';
import { MoreVertical } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
} from './dropdown-menu';
import { Button } from './button';

interface TableRowMenuProps {
  children: React.ReactNode;
  align?: 'start' | 'center' | 'end';
}

/**
 * 테이블 행의 메뉴 (DropdownMenu wrapper)
 * 
 * @example
 * <TableRowMenu>
 *   <TableRowMenuItem onClick={(e) => handleEdit(row)}>Edit</TableRowMenuItem>
 *   <TableRowMenuItem onClick={(e) => handleDelete(row)}>Delete</TableRowMenuItem>
 * </TableRowMenu>
 */
export function TableRowMenu({ children, align = 'end' }: TableRowMenuProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
          <MoreVertical className="h-4 w-4" />
          <span className="sr-only">Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align={align}>
        {children}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

/**
 * 메뉴 아이템 (DropdownMenuItem wrapper)
 */
export function TableRowMenuItem({ 
  children, 
  variant,
  onClick,
  ...props 
}: React.ComponentProps<typeof DropdownMenuItem>) {
  return (
    <DropdownMenuItem variant={variant} onClick={onClick} {...props}>
      {children}
    </DropdownMenuItem>
  );
}

/**
 * 메뉴 구분선 (DropdownMenuSeparator wrapper)
 */
export function TableRowMenuSeparator() {
  return <DropdownMenuSeparator />;
}
