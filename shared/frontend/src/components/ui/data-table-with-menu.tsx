/**
 * 완전한 라이브러리화된 DataTable + Menu 솔루션
 * 
 * 핵심 기술:
 * 1. Context로 메뉴 상태 관리
 * 2. Portal로 자동 외부 렌더링
 * 3. Compound Components로 선언적 API
 * 4. CSS 기반 위치 계산 (Floating UI 의존성 제거)
 * 
 * 설치 필요: 없음! (React 내장 기능만 사용)
 */

import { createContext, useContext, useState, useRef, useEffect } from 'react';
import { createPortal } from 'react-dom';

// ============================================
// 1. Context로 메뉴 상태 전역 관리
// ============================================

interface MenuContextValue {
  openMenuId: string | null;
  setOpenMenuId: (id: string | null) => void;
}

const MenuContext = createContext<MenuContextValue | null>(null);

function useMenuContext() {
  const context = useContext(MenuContext);
  if (!context) {
    throw new Error('Menu components must be used within DataTableWithMenu');
  }
  return context;
}

// ============================================
// 2. 메뉴 Provider (테이블 래핑)
// ============================================

interface DataTableWithMenuProps {
  children: React.ReactNode;
}

/**
 * ✅ 사용자가 테이블을 이것으로 래핑
 * 
 * @example
 * <DataTableWithMenu>
 *   <Table>...</Table>
 * </DataTableWithMenu>
 */
export function DataTableWithMenu({ children }: DataTableWithMenuProps) {
  const [openMenuId, setOpenMenuId] = useState<string | null>(null);

  return (
    <MenuContext.Provider value={{ openMenuId, setOpenMenuId }}>
      {children}
    </MenuContext.Provider>
  );
}

// ============================================
// 3. RowMenu 컴포넌트 (Compound Component)
// ============================================

type Placement = 'top' | 'bottom' | 'left' | 'right' | 'top-start' | 'top-end' | 'bottom-start' | 'bottom-end';

interface RowMenuProps {
  rowId: string;
  rowData: any;  // 복사된 데이터 (백엔드에서 검증)
  children: React.ReactNode;
  placement?: Placement;
}

/**
 * ✅ 사용자가 테이블 셀에서 사용
 * 
 * @example
 * // rowData를 복사해서 전달 (백엔드에서 삭제 검증 처리)
 * <RowMenu rowId={row.id} rowData={row.original}>
 *   <MenuItem onClick={handleEdit}>Edit</MenuItem>
 * </RowMenu>
 */
export function RowMenu({ rowId, rowData, children, placement = 'bottom-end' }: RowMenuProps) {
  const { openMenuId, setOpenMenuId } = useMenuContext();
  const triggerRef = useRef<HTMLButtonElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);
  const [menuPosition, setMenuPosition] = useState({ top: 0, left: 0 });

  const isOpen = openMenuId === rowId;

  // 메뉴 위치 계산
  useEffect(() => {
    if (isOpen && triggerRef.current && menuRef.current) {
      const triggerRect = triggerRef.current.getBoundingClientRect();
      const menuRect = menuRef.current.getBoundingClientRect();
      
      let top;
      let left;

      // placement에 따라 위치 계산
      switch (placement) {
        case 'bottom-end':
          top = triggerRect.bottom + 4;
          left = triggerRect.right - menuRect.width;
          break;
        case 'bottom-start':
          top = triggerRect.bottom + 4;
          left = triggerRect.left;
          break;
        case 'top-end':
          top = triggerRect.top - menuRect.height - 4;
          left = triggerRect.right - menuRect.width;
          break;
        case 'top-start':
          top = triggerRect.top - menuRect.height - 4;
          left = triggerRect.left;
          break;
        case 'bottom':
          top = triggerRect.bottom + 4;
          left = triggerRect.left + (triggerRect.width - menuRect.width) / 2;
          break;
        default:
          top = triggerRect.bottom + 4;
          left = triggerRect.right - menuRect.width;
      }

      // 화면 밖으로 나가지 않도록 조정
      const maxLeft = window.innerWidth - menuRect.width - 8;
      const maxTop = window.innerHeight - menuRect.height - 8;
      
      setMenuPosition({
        top: Math.min(Math.max(8, top), maxTop),
        left: Math.min(Math.max(8, left), maxLeft),
      });
    }
  }, [isOpen, placement]);

  const handleTriggerClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    setOpenMenuId(isOpen ? null : rowId);
  };

  return (
    <>
      {/* 트리거 버튼 (테이블 내부) */}
      <button
        ref={triggerRef}
        onClick={handleTriggerClick}
        className="p-1 hover:bg-gray-100 rounded"
      >
        <MoreVerticalIcon className="h-4 w-4" />
      </button>

      {/* 메뉴 (Portal로 외부 렌더링) */}
      {isOpen &&
        createPortal(
          <>
            {/* 배경 (외부 클릭 감지용) */}
            <div
              className="fixed inset-0 z-40"
              onClick={() => setOpenMenuId(null)}
            />
            {/* 메뉴 */}
            <div
              ref={menuRef}
              style={{
                position: 'fixed',
                top: `${menuPosition.top}px`,
                left: `${menuPosition.left}px`,
              }}
              className="z-50 bg-white rounded-md shadow-lg border p-1 min-w-[150px]"
              onClick={(e) => e.stopPropagation()}
            >
              {/* ✅ children으로 메뉴 아이템 전달 */}
              <RowMenuContext.Provider value={{ rowData, close: () => setOpenMenuId(null) }}>
                {children}
              </RowMenuContext.Provider>
            </div>
          </>,
          document.body
        )}
    </>
  );
}

// ============================================
// 4. MenuItem Context
// ============================================

interface RowMenuContextValue {
  rowData: any;  // 복사된 데이터
  close: () => void;
}

const RowMenuContext = createContext<RowMenuContextValue | null>(null);

function useRowMenuContext() {
  const context = useContext(RowMenuContext);
  if (!context) {
    throw new Error('MenuItem must be used within RowMenu');
  }
  return context;
}

// ============================================
// 5. MenuItem 컴포넌트
// ============================================

interface MenuItemProps {
  onClick?: (rowData: any) => void;
  variant?: 'default' | 'destructive';
  disabled?: boolean;
  children: React.ReactNode;
  icon?: React.ComponentType<{ className?: string }>;
}

/**
 * ✅ 선언적으로 메뉴 아이템 정의
 */
export function MenuItem({ onClick, variant = 'default', disabled, children, icon: Icon }: MenuItemProps) {
  const { rowData, close } = useRowMenuContext();

  const handleClick = () => {
    if (disabled) return;
    
    // ✅ 복사된 rowData 사용 (삭제 검증은 백엔드에서 처리)
    onClick?.(rowData);
    close();
  };

  return (
    <button
      onClick={handleClick}
      disabled={disabled}
      className={`
        w-full px-3 py-1.5 text-left rounded flex items-center gap-2
        ${disabled ? 'opacity-50 cursor-not-allowed' : 'hover:bg-gray-100 cursor-pointer'}
        ${variant === 'destructive' ? 'text-red-600' : ''}
      `}
    >
      {Icon && <Icon className="h-4 w-4" />}
      {children}
    </button>
  );
}

// ============================================
// 6. MenuSeparator
// ============================================

export function MenuSeparator() {
  return <div className="h-px bg-gray-200 my-1" />;
}

// ============================================
// 7. 사용 예제
// ============================================

import { Pencil, Trash2, Copy } from 'lucide-react';

function MoreVerticalIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
    </svg>
  );
}

/**
 * ✅ 완성된 사용 예제
 * 
 * 사용자는 이렇게만 쓰면 됨!
 */
export function ExampleUsage() {
  const [data, setData] = useState([
    { id: '1', name: 'Alice', email: 'alice@example.com' },
    { id: '2', name: 'Bob', email: 'bob@example.com' },
  ]);

  // 1초마다 갱신해도 메뉴는 열린 상태 유지!
  useEffect(() => {
    const interval = setInterval(() => {
      setData((prev) =>
        prev.map((item) => ({
          ...item,
          email: `${item.name.toLowerCase()}@${Date.now()}.com`,
        }))
      );
    }, 1000);
    return () => clearInterval(interval);
  }, []);

  const handleEdit = (row: any) => {
    console.log('Edit:', row);
  };

  const handleDuplicate = (row: any) => {
    console.log('Duplicate:', row);
  };

  const handleDelete = (row: any) => {
    console.log('Delete:', row);
  };

  return (
    // ✅ 1. 테이블을 래핑
    <DataTableWithMenu>
      <table className="w-full border">
        <thead>
          <tr>
            <th className="border p-2">Name</th>
            <th className="border p-2">Email</th>
            <th className="border p-2 w-12"></th>
          </tr>
        </thead>
        <tbody>
          {data.map((row) => (
            <tr key={row.id} className="border">
              <td className="border p-2">{row.name}</td>
              <td className="border p-2">{row.email}</td>
              <td className="border p-2">
                {/* ✅ 2. 선언적으로 메뉴 정의 */}
                <RowMenu rowId={row.id} rowData={row}>
                  <MenuItem onClick={handleEdit} icon={Pencil}>
                    Edit
                  </MenuItem>
                  <MenuItem onClick={handleDuplicate} icon={Copy}>
                    Duplicate
                  </MenuItem>
                  <MenuSeparator />
                  <MenuItem onClick={handleDelete} variant="destructive" icon={Trash2}>
                    Delete
                  </MenuItem>
                </RowMenu>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      <div className="mt-4 text-sm text-gray-500">
        ⚡ 1초마다 email 갱신 중... 메뉴를 열어두고 확인하세요!
      </div>
    </DataTableWithMenu>
  );
}

// ============================================
// 8. TypeScript 타입 Export
// ============================================

export type {
  DataTableWithMenuProps,
  RowMenuProps,
  MenuItemProps,
};
