# TableTabsTemplate - 확장 가능한 탭 테이블 시스템

React 기반의 모듈화된 테이블 템플릿 시스템입니다. TanStack Table과 TanStack Query를 완벽하게 지원하며, 단일 테이블부터 멀티탭 테이블까지 유연하게 구성할 수 있습니다.

## 배경

기존의 단일 파일 구조에서 발생하는 복잡도와 확장성 문제를 해결하기 위해 설계되었습니다:

### 문제점

- 모든 테이블 로직이 한 곳에 집중되어 복잡도 증가
- 새로운 테이블 타입 추가 시 기존 코드에 영향
- 탭별 독립적인 상태 관리의 어려움
- 재사용 가능한 컴포넌트 부족

### 해결책

- **모듈화된 구조**: 테이블 타입별 독립된 컴포넌트
- **공통 로직 추상화**: shared 디렉토리에 재사용 가능한 컴포넌트와 훅
- **독립적 상태 관리**: 각 탭의 검색, 페이지네이션, 정렬 상태 격리
- **확장성**: 새로운 테이블 타입을 기존 코드 영향 없이 추가 가능

## 사용법

### 1. 기본 import

```tsx
import {
    TableTabsTemplate,
    Tabs,
    Tab,
    TableBasic,
    TablePagination,
    TableInfinity,
} from '@components/template/table-tabs';
```

### 2. 단일 테이블 만들기

탭 네비게이션 없이 깔끔한 단일 테이블 인터페이스를 만들 수 있습니다:

```tsx
function UserManagement() {
    const [users] = useState(generateUsers(50));

    return (
        <TableTabsTemplate
            name="User Management"
            description="Single table interface without tab navigation"
            showTabs={false}  // 탭 네비게이션 숨김
        >
            <Tabs>
                <Tab name="Users">
                    <TableBasic
                        data={users}
                        columns={userColumns}
                        tableKey="single-users-table"
                        useSearch={true}
                        searchableColumns={['name', 'email']}
                        useTableSorting={true}
                        rowCursor={true}
                    />
                </Tab>
            </Tabs>
        </TableTabsTemplate>
    );
}
```

### 3. 멀티탭 테이블 만들기

여러 테이블을 탭으로 구성하여 관리할 수 있습니다:

```tsx
function AdminDashboard() {
    const [users] = useState(generateUsers(15));
    const [products] = useState(generateProducts(25));
    const [orders, setOrders] = useState(generateOrders(8));

    const handlePageChange = (page: number) => {
        console.log(`Page changed to: ${page}`);
    };

    const loadMoreOrders = async () => {
        // 무한 스크롤 로직
        const nextBatch = await fetchMoreOrders();
        setOrders(prev => [...prev, ...nextBatch]);
    };

    return (
        <TableTabsTemplate
            name="Admin Dashboard"
            description="Multi-tab interface with different table types"
            defaultTab="Users"
            onTabChange={(tabName) => console.log(`Switched to: ${tabName}`)}
            showTabs={true}  // 기본값: true
        >
            <Tabs>
                <Tab name="Users">
                    <TableBasic
                        data={users}
                        columns={userColumns}
                        tableKey="users-table"
                        useSearch={true}
                        leftCustomItems={<UserFilters/>}
                        rightCustomItems={<UserActions/>}
                        searchableColumns={['name', 'email']}
                        useTableSorting={true}
                        rowCursor={true}
                    />
                </Tab>

                <Tab name="Products">
                    <TablePagination
                        data={products}
                        columns={productColumns}
                        tableKey="products-table"
                        onPageChange={handlePageChange}
                        onPageSizeChange={(size) => console.log(`Page size: ${size}`)}
                        useSearch={true}
                        leftCustomItems={<ProductFilters/>}
                        rightCustomItems={<ProductActions/>}
                    />
                </Tab>

                <Tab name="Orders">
                    <TableInfinity
                        data={orders}
                        columns={orderColumns}
                        tableKey="orders-table"
                        onLoadMore={loadMoreOrders}
                        hasNextPage={hasNextPage}
                        isLoading={isLoading}
                        useSearch={true}
                        searchableColumns={['productName', 'status']}
                    />
                </Tab>

                <Tab name="Disabled Tab" disabled={true}>
                    <TableBasic
                        data={[]}
                        columns={userColumns}
                        tableKey="disabled-table"
                        emptyCustomMessage="This tab is disabled"
                    />
                </Tab>
            </Tabs>
        </TableTabsTemplate>
    );
}
```

### 4. TanStack Query와 함께 사용하기

TanStack Query의 자동/수동 새로고침 기능을 완벽하게 지원합니다:

```tsx
function OptimizedUserTable() {
    const {data, refetch, isRefetching} = useQuery({
        queryKey: ['users'],
        queryFn: fetchUsers,
        refetchInterval: 30000, // 30초마다 자동 새로고침
    });

    // 컬럼 정의를 메모이제이션
    const columns = useMemo<ColumnDef<User>[]>(() => [
        {
            id: 'name',
            header: '이름',
            accessorKey: 'name',
        },
        {
            id: 'email',
            header: '이메일',
            accessorKey: 'email',
        },
    ], []);

    // 새로고침 핸들러를 메모이제이션
    const handleRefresh = useCallback(async () => {
        await refetch();
    }, [refetch]);

    // 데이터를 메모이제이션
    const tableData = useMemo(() => data || [], [data]);

    return (
        <TableBasic
            data={tableData}
            columns={columns}
            tableKey="users-table"
            useSearch
            onRefresh={handleRefresh}           // 수동 새로고침
            isRefreshing={isRefetching}         // 로딩 상태 (자동/수동 모두 감지)
            searchableColumns={['name', 'email']}
        />
    );
}
```

**새로고침 동작 방식:**

- **자동 새로고침**: `refetchInterval`로 30초마다 자동 실행
- **수동 새로고침**: 버튼 클릭 시 `refetch()` 실행
- **로딩 상태**: `isRefetching`이 자동/수동 구분 없이 모든 새로고침 상태를 감지
- **성능 최적화**: React.memo, useCallback, useMemo로 불필요한 리렌더 방지

### 5. Sheet 연동하기

테이블 행 클릭 시 Sheet(사이드 패널)를 열 수 있습니다:

```tsx
function UserTableWithSheet() {
    const [users] = useState(generateUsers(15));
    const sheetStore = useSheetStore();

    // User 상세 정보를 Sheet에 표시
    const handleUserClick = (user: User) => {
        sheetStore.setSheet(
            <div className="w-full h-full overflow-auto">
                <div className="p-6">
                    <h2 className="text-2xl font-bold mb-4">User Details</h2>
                    <div className="space-y-4">
                        <div className="border-b pb-2">
                            <p className="text-sm text-gray-500">Name</p>
                            <p className="font-medium">{user.name}</p>
                        </div>
                        <div className="border-b pb-2">
                            <p className="text-sm text-gray-500">Email</p>
                            <p className="font-medium">{user.email}</p>
                        </div>
                        {/* 추가 정보 */}
                    </div>
                </div>
            </div>
        );
        sheetStore.setLayout(500); // Sheet 너비 설정
    };

    const userColumns: ColumnDef<User>[] = [
        {
            accessorKey: 'name',
            header: 'Name',
            cell: ({row}) => (
                <span className="text-blue-600 hover:underline cursor-pointer">
                    {row.original.name}
                </span>
            ),
        },
        // 다른 컬럼들...
    ];

    return (
        <TableBasic
            data={users}
            columns={userColumns}
            tableKey="users-sheet-table"
            dataCellOnClick={(cell) => {
                // 특정 컬럼 클릭 시에만 Sheet 열기
                if (cell.column.id === 'name') {
                    handleUserClick(cell.row.original as User);
                }
            }}
            enableOnClickCloseSheet={true}  // 테이블 외부 클릭 시 Sheet 닫기
            rowCursor={true}
        />
    );
}
```

**Sheet 동작 방식:**

- **Sheet 열기**: `sheetStore.setSheet(content)`로 컨텐츠 설정
- **Sheet 크기**: `sheetStore.setLayout(width)`로 너비 설정 (픽셀)
- **자동 닫기**: `enableOnClickCloseSheet={true}` 설정 시 테이블 외부 클릭으로 자동 닫힘
- **행 변경**: 다른 행 클릭 시 Sheet 내용이 자동으로 변경됨
- **성능**: Sheet Store는 Zustand로 구현되어 효율적인 상태 관리

## 주요 옵션 설명

### TableTabsTemplate 옵션

| 옵션            | 타입                          | 기본값    | 설명            |
|---------------|-----------------------------|--------|---------------|
| `name`        | `string`                    | -      | 페이지 제목        |
| `description` | `string`                    | -      | 페이지 설명        |
| `showTabs`    | `boolean`                   | `true` | 탭 네비게이션 표시 여부 |
| `defaultTab`  | `string`                    | -      | 기본 활성 탭       |
| `onTabChange` | `(tabName: string) => void` | -      | 탭 변경 콜백       |

### Tab 옵션

| 옵션         | 타입        | 기본값     | 설명        |
|------------|-----------|---------|-----------|
| `name`     | `string`  | -       | 탭 이름 (필수) |
| `disabled` | `boolean` | `false` | 탭 비활성화 여부 |

### 공통 테이블 옵션 (CommonTableProps)

| 옵션                        | 타입                               | 기본값     | 설명                    |
|---------------------------|----------------------------------|---------|-----------------------|
| `tableKey`                | `string`                         | -       | 테이블 고유 키 (상태 격리용)     |
| `useSearch`               | `boolean`                        | `false` | 검색 기능 활성화             |
| `searchValue`             | `string`                         | -       | 외부 검색 값               |
| `onSearchChange`          | `(searchTerm: string) => void`   | -       | 검색 변경 콜백              |
| `searchableColumns`       | `string[]`                       | -       | 검색 가능한 컬럼 ID 배열       |
| `useTableSorting`         | `boolean`                        | `true`  | 정렬 기능 활성화             |
| `rowCursor`               | `boolean`                        | `false` | 행 클릭 시 포인터 커서         |
| `leftCustomItems`         | `React.ReactElement`             | -       | 왼쪽 커스텀 아이템 (필터 등)     |
| `rightCustomItems`        | `React.ReactElement`             | -       | 오른쪽 커스텀 아이템 (액션 버튼 등) |
| `onRefresh`               | `() => void \| Promise<void>`    | -       | 새로고침 콜백               |
| `isRefreshing`            | `boolean`                        | `false` | 새로고침 로딩 상태            |
| `dataCellOnClick`         | `(cell: Cell<any, any>) => void` | -       | 셀 클릭 콜백               |
| `emptyCustomMessage`      | `string`                         | -       | 빈 상태 메시지              |
| `onTableReady`            | `(table: any) => void`           | -       | 테이블 준비 완료 콜백          |
| `onCurrentPageDataChange` | `(data: any[]) => void`          | -       | 현재 페이지 데이터 변경 콜백      |
| `enableOnClickCloseSheet` | `boolean`                        | `true`  | 테이블 외부 클릭 시 Sheet 닫기  |

### TablePagination 전용 옵션

| 옵션                 | 타입                           | 기본값 | 설명           |
|--------------------|------------------------------|-----|--------------|
| `onPageChange`     | `(page: number) => void`     | -   | 페이지 변경 콜백    |
| `onPageSizeChange` | `(pageSize: number) => void` | -   | 페이지 크기 변경 콜백 |

### TableInfinity 전용 옵션

| 옵션            | 타입                    | 기본값     | 설명           |
|---------------|-----------------------|---------|--------------|
| `onLoadMore`  | `() => Promise<void>` | -       | 더 로드 콜백      |
| `hasNextPage` | `boolean`             | `false` | 다음 페이지 존재 여부 |
| `isLoading`   | `boolean`             | `false` | 로딩 상태        |

## 성능 최적화

### 메모이제이션 전략

```tsx
// 컴포넌트 메모이제이션
const TableComponent = memo(BaseTableComponent);

// 컬럼 정의 메모이제이션
const columns = useMemo(() => [...], []);

// 핸들러 메모이제이션
const handleRefresh = useCallback(async () => {
    await refetch();
}, [refetch]);

// 데이터 메모이제이션
const tableData = useMemo(() => data || [], [data]);
```

### 상태 격리

각 테이블은 `tableKey`를 통해 독립적인 상태를 관리합니다:

```tsx
// 각각 독립적인 검색/정렬/페이지네이션 상태
<TableBasic tableKey="users-table" ...
/>
<TableBasic tableKey="products-table" ...
/>
```

## 브라우저 호환성

- **Chrome**: 완전 지원
- **Safari**: 완전 지원 (border-collapse 이슈 해결됨)
- **Firefox**: 완전 지원
- **Edge**: 완전 지원

## 데모 페이지

실제 동작을 확인할 수 있는 데모 페이지:

1. **Table Tabs Demo** (`/table-tabs-demo`): 
   - 멀티탭 테이블
   - Sheet 연동 (Users 탭의 이름 클릭)
   - Disable 기능
   - 외부 클릭으로 Sheet 닫기
2. **Single Table Demo** (`/single-table-demo`): 단일 테이블 인터페이스

## 주요 기능

### Sheet 연동
- ✅ 행 클릭 시 Sheet(사이드 패널) 자동 열기
- ✅ 다른 행 클릭 시 Sheet 내용 자동 변경
- ✅ 테이블 외부 클릭 시 Sheet 자동 닫기
- ✅ `enableOnClickCloseSheet` 옵션으로 동작 제어

### 상태 관리
- ✅ 탭별 독립적인 검색/정렬/페이지네이션 상태
- ✅ `tableKey`로 상태 격리
- ✅ Zustand 기반 Sheet Store

### 성능 최적화
- ✅ React.memo로 불필요한 리렌더 방지
- ✅ useCallback/useMemo 활용
- ✅ TanStack Query 완벽 지원

````
