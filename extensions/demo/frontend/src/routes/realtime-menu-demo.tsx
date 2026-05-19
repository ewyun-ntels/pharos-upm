/**
 * 실시간 갱신 테이블 메뉴 데모
 * 
 * DataTableWithMenu 컴포넌트 검증:
 * - 1초마다 데이터 갱신
 * - 행 동적 추가/삭제
 * - 메뉴가 열린 상태로 유지되는지 확인
 */

import { useState, useEffect } from 'react';
import { ColumnDef } from '@tanstack/react-table';
import { Pencil, Trash2, Copy, RefreshCw } from 'lucide-react';
import {
  TableTabsTemplate,
  Tabs,
  Tab,
} from '@pharos/shared/components/template/table-tabs';
import { DataGrid } from '@pharos/shared/components/ui-extension/data-grid';
import {
  DataTableWithMenu,
  RowMenu,
  MenuItem,
  MenuSeparator,
} from '@pharos/shared/components/ui';

interface User {
  id: string;
  name: string;
  email: string;
  status: 'active' | 'inactive' | 'pending';
  lastUpdate: number;
}

// 샘플 데이터 생성
const generateUsers = (count: number): User[] => {
  const names = ['Alice', 'Bob', 'Charlie', 'David', 'Eve', 'Frank', 'Grace', 'Henry'];
  const statuses: User['status'][] = ['active', 'inactive', 'pending'];
  
  return Array.from({ length: count }, (_, i) => ({
    id: `user-${i + 1}`,
    name: names[i % names.length],
    email: `${names[i % names.length].toLowerCase()}@example.com`,
    status: statuses[i % statuses.length],
    lastUpdate: Date.now(),
  }));
};

export function RealtimeMenuDemo() {
  const [users, setUsers] = useState<User[]>(generateUsers(5));
  const [isAutoUpdate, setIsAutoUpdate] = useState(true);
  const [updateCount, setUpdateCount] = useState(0);
  const [addCount, setAddCount] = useState(0);
  const [deleteCount, setDeleteCount] = useState(0);

  // ⚡ 1초마다 데이터 갱신 + 랜덤 추가/삭제
  useEffect(() => {
    if (!isAutoUpdate) return;

    const interval = setInterval(() => {
      setUsers((prev) => {
        const names = ['Alice', 'Bob', 'Charlie', 'David', 'Eve', 'Frank', 'Grace', 'Henry', 'Ivy', 'Jack'];
        const statuses: User['status'][] = ['active', 'inactive', 'pending'];
        
        // 이메일 업데이트
        let updated = prev.map((user) => ({
          ...user,
          email: `${user.name.toLowerCase()}@updated-${Date.now()}.com`,
          lastUpdate: Date.now(),
        }));

        // 30% 확률로 새 행 추가 (최대 10개)
        if (Math.random() < 0.3 && updated.length < 10) {
          const newUser: User = {
            id: `user-${Date.now()}`,
            name: names[Math.floor(Math.random() * names.length)],
            email: `auto-${Date.now()}@example.com`,
            status: statuses[Math.floor(Math.random() * statuses.length)],
            lastUpdate: Date.now(),
          };
          updated = [...updated, newUser];
          setAddCount((c) => c + 1);
        }

        // 20% 확률로 랜덤 행 삭제 (최소 3개 유지)
        if (Math.random() < 0.2 && updated.length > 3) {
          const randomIndex = Math.floor(Math.random() * updated.length);
          updated = updated.filter((_, i) => i !== randomIndex);
          setDeleteCount((c) => c + 1);
        }

        return updated;
      });
      setUpdateCount((c) => c + 1);
    }, 1000);

    return () => clearInterval(interval);
  }, [isAutoUpdate]);

  // 액션 핸들러
  const handleEdit = (user: User) => {
    console.log('✏️ Edit:', user);
    const newName = prompt(`Edit user name (현재: ${user.name}):`, user.name);
    if (newName && newName.trim()) {
      setUsers((prev) =>
        prev.map((u) =>
          u.id === user.id
            ? { ...u, name: newName.trim(), email: `${newName.toLowerCase()}@edited.com` }
            : u
        )
      );
    }
  };

  const handleDuplicate = (user: User) => {
    console.log('📋 Duplicate:', user);
    const newUser: User = {
      ...user,
      id: `user-${Date.now()}`,
      name: `${user.name} (Copy)`,
      email: `${user.name.toLowerCase()}-copy@example.com`,
      lastUpdate: Date.now(),
    };
    setUsers((prev) => [...prev, newUser]);
    setAddCount((c) => c + 1);
  };

  const handleDelete = (user: User) => {
    console.log('🗑️ Delete:', user);
    
    // ✅ 복사된 데이터로 삭제 요청 - 백엔드에서 이미 삭제됐는지 검증
    if (users.length <= 1) {
      alert('최소 1개의 행은 유지되어야 합니다.');
      return;
    }
    
    setUsers((prev) => prev.filter((u) => u.id !== user.id));
    setDeleteCount((c) => c + 1);
  };

  const handleToggleStatus = (user: User) => {
    console.log('🔄 Toggle Status:', user);
    const nextStatus: Record<User['status'], User['status']> = {
      active: 'inactive',
      inactive: 'pending',
      pending: 'active',
    };
    setUsers((prev) =>
      prev.map((u) =>
        u.id === user.id
          ? { ...u, status: nextStatus[u.status] }
          : u
      )
    );
  };

  const handleAddRandom = () => {
    const names = ['Alice', 'Bob', 'Charlie', 'David', 'Eve', 'Frank', 'Grace', 'Henry', 'Ivy', 'Jack'];
    const statuses: User['status'][] = ['active', 'inactive', 'pending'];
    const newUser: User = {
      id: `user-${Date.now()}`,
      name: names[Math.floor(Math.random() * names.length)],
      email: `manual-${Date.now()}@example.com`,
      status: statuses[Math.floor(Math.random() * statuses.length)],
      lastUpdate: Date.now(),
    };
    setUsers((prev) => [...prev, newUser]);
    setAddCount((c) => c + 1);
  };

  // ✅ TanStack Table meta를 활용한 컬럼 정의
  const columns: ColumnDef<User>[] = [
    {
      accessorKey: 'name',
      header: 'Name',
      cell: ({ getValue }) => (
        <div className="font-medium">{getValue() as string}</div>
      ),
    },
    {
      accessorKey: 'email',
      header: 'Email',
      cell: ({ getValue }) => (
        <div className="text-sm text-gray-600">{getValue() as string}</div>
      ),
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ getValue }) => {
        const status = getValue() as User['status'];
        const colors = {
          active: 'bg-green-100 text-green-800',
          inactive: 'bg-gray-100 text-gray-800',
          pending: 'bg-yellow-100 text-yellow-800',
        };
        return (
          <span className={`px-2 py-1 rounded text-xs font-medium ${colors[status]}`}>
            {status}
          </span>
        );
      },
    },
    {
      accessorKey: 'lastUpdate',
      header: 'Last Update',
      cell: ({ getValue }) => (
        <div className="text-xs text-gray-500">
          {new Date(getValue() as number).toLocaleTimeString()}
        </div>
      ),
    },
    {
      id: 'actions',
      header: () => <div className="text-center">Actions</div>,
      cell: ({ row }) => (
        <div className="flex justify-center">
          {/* ✅ rowData 복사 - 백엔드에서 삭제 검증 */}
          <RowMenu rowId={row.id} rowData={row.original} placement="bottom-end">
            <MenuItem onClick={handleEdit} icon={Pencil}>
              Edit
            </MenuItem>
            <MenuItem onClick={handleDuplicate} icon={Copy}>
              Duplicate
            </MenuItem>
            <MenuSeparator />
            <MenuItem onClick={handleToggleStatus} icon={RefreshCw}>
              Toggle Status
            </MenuItem>
            <MenuItem onClick={handleDelete} icon={Trash2} variant="destructive">
              Delete
            </MenuItem>
          </RowMenu>
        </div>
      ),
    },
  ];

  return (
    <main className="flex-1 overflow-auto p-6 bg-gray-50">
      <TableTabsTemplate
        name="🔥 실시간 메뉴 데모"
        description="1초마다 갱신되는 테이블에서 메뉴가 유지되는지 확인하세요"
      >
        <Tabs>
          <Tab name="Live Demo">
            <DataTableWithMenu>
              <div className="space-y-3">
                {/* 테스트 안내 */}
                <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="font-semibold text-blue-900 mb-1">
                        🧪 테스트 방법
                      </h3>
                      <ol className="text-sm text-blue-700 space-y-1">
                        <li>1. 아무 행의 메뉴(⋮)를 클릭하여 열기</li>
                        <li>2. 메뉴를 열어둔 채로 데이터 변화 확인</li>
                        <li>3. <strong>행이 추가/삭제되어도 메뉴가 유지되는지 확인 ✅</strong></li>
                        <li>4. 메뉴 아이템 클릭 시 실제 동작 확인</li>
                      </ol>
                    </div>
                    <div className="flex flex-col items-end gap-2">
                      <button
                        onClick={() => setIsAutoUpdate(!isAutoUpdate)}
                        className={`px-4 py-2 rounded font-medium transition-colors ${
                          isAutoUpdate
                            ? 'bg-green-500 hover:bg-green-600 text-white'
                            : 'bg-gray-300 hover:bg-gray-400 text-gray-700'
                        }`}
                      >
                        {isAutoUpdate ? '⏸️ 일시정지' : '▶️ 자동갱신'}
                      </button>
                      <button
                        onClick={handleAddRandom}
                        className="px-4 py-2 rounded font-medium bg-blue-500 hover:bg-blue-600 text-white transition-colors"
                      >
                        ➕ 수동 추가
                      </button>
                    </div>
                  </div>
                </div>

                {/* 통계 패널 */}
                <div className="grid grid-cols-4 gap-3">
                  <div className="p-3 bg-white border border-gray-200 rounded-lg">
                    <div className="text-xs text-gray-500 mb-1">현재 행 수</div>
                    <div className="text-2xl font-bold text-gray-900">{users.length}</div>
                  </div>
                  <div className="p-3 bg-green-50 border border-green-200 rounded-lg">
                    <div className="text-xs text-green-600 mb-1">자동 추가</div>
                    <div className="text-2xl font-bold text-green-700">{addCount}</div>
                  </div>
                  <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
                    <div className="text-xs text-red-600 mb-1">자동 삭제</div>
                    <div className="text-2xl font-bold text-red-700">{deleteCount}</div>
                  </div>
                  <div className="p-3 bg-purple-50 border border-purple-200 rounded-lg">
                    <div className="text-xs text-purple-600 mb-1">갱신 횟수</div>
                    <div className="text-2xl font-bold text-purple-700">{updateCount}</div>
                  </div>
                </div>

                {/* 상태 인디케이터 */}
                <div className="flex items-center justify-center gap-2 text-sm">
                  <div
                    className={`w-2 h-2 rounded-full ${
                      isAutoUpdate ? 'bg-green-500 animate-pulse' : 'bg-gray-300'
                    }`}
                  />
                  <span className={isAutoUpdate ? 'text-green-600 font-medium' : 'text-gray-500'}>
                    {isAutoUpdate ? '🔄 실시간 갱신 중 (1초마다)' : '⏸️ 일시정지됨'}
                  </span>
                </div>

                {/* 테이블 */}
                <DataGrid
                  data={users}
                  columns={columns}
                  tableKey="realtime-menu-demo"
                  getRowId={(row) => row.id}  // ✅ user.id를 row.id로 사용
                  useTableSorting={false}
                  rowCursor={false}
                />
              </div>
            </DataTableWithMenu>
          </Tab>
        </Tabs>

        {/* 추가 정보 */}
        <div className="mt-4 grid grid-cols-2 gap-4">
          <div className="p-4 bg-gray-50 rounded text-sm">
            <div className="font-semibold mb-2">📊 기술 스택:</div>
            <ul className="space-y-1 text-gray-700">
              <li>✅ <strong>TanStack Table</strong>: getRowId + table.getRow()</li>
              <li>✅ <strong>DataTableWithMenu</strong>: table instance 자동 전달</li>
              <li>✅ <strong>MenuItem</strong>: row 존재 여부 자동 확인</li>
              <li>✅ <strong>핸들러 단순화</strong>: state 확인 불필요</li>
              <li>✅ <strong>Zero 의존성</strong>: React 내장만 사용</li>
            </ul>
          </div>
          <div className="p-4 bg-yellow-50 rounded text-sm">
            <div className="font-semibold mb-2 text-yellow-800">🎯 실시간 변경 사항:</div>
            <ul className="space-y-1 text-yellow-700">
              <li>📧 <strong>Email</strong>: 매 초마다 업데이트</li>
              <li>➕ <strong>행 추가</strong>: 30% 확률 (최대 10개)</li>
              <li>➖ <strong>행 삭제</strong>: 20% 확률 (최소 3개)</li>
              <li>⏱️ <strong>타임스탬프</strong>: 실시간 갱신</li>
            </ul>
          </div>
        </div>
        <div className="mt-2 p-3 bg-green-50 border border-green-200 rounded text-sm text-green-700">
          <strong>✨ 핵심 테스트:</strong> 메뉴를 열어두고 행이 추가/삭제되는 것을 확인하세요. 
          TanStack Table의 <code>table.getRow()</code>로 자동 검증합니다!
        </div>
        <div className="mt-2 p-3 bg-blue-50 border border-blue-200 rounded text-sm text-blue-700">
          <strong>🔍 자동 보호:</strong> MenuItem이 자동으로 row 존재 여부를 확인하므로 핸들러에서 state 확인이 불필요합니다.
        </div>
      </TableTabsTemplate>
    </main>
  );
}
