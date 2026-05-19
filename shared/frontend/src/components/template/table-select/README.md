# TableSelectTemplate

Select 드롭다운으로 옵션을 선택하고, 선택된 옵션에 따라 다른 컨텐츠(Table 등)를 표시하는 재사용 가능한 템플릿 컴포넌트입니다.

## 🎯 주요 특징

- ✅ **Select 기반 네비게이션**: Tab 대신 Select 드롭다운 사용
- ✅ **유연한 Body 주입**: Children으로 어떤 컴포넌트든 주입 가능
- ✅ **왼쪽/오른쪽 커스텀 아이템**: 검색, 필터, 버튼 등 자유롭게 배치
- ✅ **옵션 비활성화**: 특정 옵션 disable 가능
- ✅ **Context API**: 자식 컴포넌트에서 선택 상태 접근 가능
- ✅ **타입 안전성**: TypeScript 완벽 지원

## 📁 파일 구조

```
table-select/
├── index.ts                           # Public exports
├── shared/
│   ├── TableSelectTemplate.tsx        # 메인 템플릿 컴포넌트
│   └── types.ts                       # TypeScript types
└── README.md                          # 이 파일
```

## 🚀 기본 사용법

### 1. 가장 간단한 예제

```tsx
import {TableSelectTemplate, Options, Option} from '@components/template/table-select';
import {TableBasic} from '@components/template/table-tabs';

export default function SimpleExample() {
  return (
    <TableSelectTemplate
      title="User Management"
      selectPlaceholder="Select user type"
    >
      <Options>
        <Option value="admin" label="Administrators">
          <TableBasic data={adminUsers} columns={userColumns} tableKey="admin-users" />
        </Option>
        <Option value="regular" label="Regular Users">
          <TableBasic data={regularUsers} columns={userColumns} tableKey="regular-users" />
        </Option>
      </Options>
    </TableSelectTemplate>
  );
}
```

### 2. Alert Editor 예제 (실제 사용 케이스)

```tsx
import {TableSelectTemplate, Options, Option} from '@components/template/table-select';
import {AlertRuleEditor} from '@features/alert/components/alert-rule-editor';
import {NotificationEditor} from '@features/alert/components/notification-editor';
import {Button} from '@components/ui/button';
import {Plus, Save} from 'lucide-react';

export default function AlertEditorPage() {
  const leftItems = (
    <Button variant="outline" size="sm" className="h-8">
      <Save className="w-4 h-4 mr-2" />
      Save Draft
    </Button>
  );

  const rightItems = (
    <>
      <Button variant="default" size="sm" className="h-8">
        <Plus className="w-4 h-4 mr-2" />
        Create New
      </Button>
      <Button variant="secondary" size="sm" className="h-8">
        Import
      </Button>
    </>
  );

  return (
    <TableSelectTemplate
      title="Alert Configuration"
      description="Configure alert rules and notification settings"
      selectLabel="Editor Type"
      selectPlaceholder="Select editor"
      selectWidth="180px"
      defaultValue="Alert Rules"
      onSelectionChange={(value) => console.log('Selected:', value)}
      leftCustomItems={leftItems}
      rightCustomItems={rightItems}
    >
      <Options>
        <Option value="rule" label="Alert Rules">
          <AlertRuleEditor />
        </Option>
        <Option value="notification" label="Notifications">
          <NotificationEditor />
        </Option>
        <Option value="history" label="Alert History" disabled={true}>
          <div>Coming soon...</div>
        </Option>
      </Options>
    </TableSelectTemplate>
  );
}
```

### 3. 고급 예제 - 필터와 검색

```tsx
import {TableSelectTemplate, Options, Option} from '@components/template/table-select';
import {TableBasic} from '@components/template/table-tabs';
import {Input} from '@components/ui/input';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@components/ui/select';
import {Button} from '@components/ui/button';
import {useState} from 'react';

export default function AdvancedExample() {
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');

  const leftItems = (
    <>
      <Input
        placeholder="Search..."
        value={searchTerm}
        onChange={(e) => setSearchTerm(e.target.value)}
        className="w-48 h-8"
      />
      <Select value={statusFilter} onValueChange={setStatusFilter}>
        <SelectTrigger className="w-32 h-8">
          <SelectValue placeholder="Status" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All</SelectItem>
          <SelectItem value="active">Active</SelectItem>
          <SelectItem value="inactive">Inactive</SelectItem>
        </SelectContent>
      </Select>
    </>
  );

  const rightItems = (
    <>
      <Button variant="outline" size="sm" className="h-8">
        Export
      </Button>
      <Button variant="default" size="sm" className="h-8">
        Create
      </Button>
    </>
  );

  return (
    <TableSelectTemplate
      title="Dashboard Management"
      description="Manage your dashboards by category"
      selectLabel="Category"
      selectPlaceholder="Select category"
      selectWidth="200px"
      leftCustomItems={leftItems}
      rightCustomItems={rightItems}
    >
      <Options>
        <Option value="personal" label="Personal">
          <TableBasic data={personalDashboards} columns={columns} tableKey="personal" />
        </Option>
        <Option value="team" label="Team">
          <TableBasic data={teamDashboards} columns={columns} tableKey="team" />
        </Option>
        <Option value="public" label="Public">
          <TableBasic data={publicDashboards} columns={columns} tableKey="public" />
        </Option>
      </Options>
    </TableSelectTemplate>
  );
}
```

## 📚 API Reference

### TableSelectTemplate Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `title` | `string` | `undefined` | 페이지 제목 |
| `description` | `string` | `undefined` | 페이지 설명 (제목 아래 표시) |
| `children` | `React.ReactNode` | **Required** | `<Options>` 컴포넌트 |
| `defaultValue` | `string` | 첫 번째 옵션 | 기본 선택값 |
| `className` | `string` | `undefined` | 추가 CSS 클래스 |
| `onSelectionChange` | `(value: string) => void` | `undefined` | 선택 변경 콜백 |
| `selectPlaceholder` | `string` | `'Select an option'` | Select placeholder |
| `selectLabel` | `string` | `undefined` | Select 그룹 레이블 |
| `selectWidth` | `string` | `'200px'` | Select 너비 |
| `leftCustomItems` | `React.ReactNode` | `undefined` | 왼쪽 커스텀 아이템 |
| `rightCustomItems` | `React.ReactNode` | `undefined` | 오른쪽 커스텀 아이템 |

### Option Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `value` | `string` | **Required** | 옵션 고유 식별자 |
| `label` | `string` | **Required** | Select에 표시될 레이블 |
| `children` | `React.ReactNode` | **Required** | 옵션 선택 시 렌더링될 컨텐츠 |
| `disabled` | `boolean` | `false` | 옵션 비활성화 여부 |

### useTableSelectContext Hook

```tsx
import {useTableSelectContext} from '@components/template/table-select';

function MyComponent() {
  const {
    selectedValue,        // 현재 선택된 값
    setSelectedValue,     // 선택 변경 함수
    optionsData,          // 모든 옵션 데이터
    disabledOptions,      // 비활성화된 옵션 Set
  } = useTableSelectContext();

  return <div>Current: {selectedValue}</div>;
}
```

## 🎨 레이아웃 구조

```
┌─────────────────────────────────────────────────────────────┐
│ Title                                                        │
│ Description                                                  │
│                                                              │
│ [Left Items] [Select ▼]              [Right Items] [Button] │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│                        Body Content                          │
│                    (Injected Children)                       │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 왼쪽 정렬 구조
- **Left Items + Select**: 왼쪽에 함께 배치 (gap-2)
- **Right Items**: 오른쪽에 별도로 배치
- **Responsive**: justify-between으로 양쪽 정렬 유지

## 🆚 TableTabsTemplate vs TableSelectTemplate

| Feature | TableTabsTemplate | TableSelectTemplate |
|---------|-------------------|---------------------|
| Navigation | Tabs (수평 버튼) | Select (드롭다운) |
| 공간 효율 | 많은 공간 차지 | 컴팩트 |
| 옵션 수 | 3-5개 권장 | 10개 이상 가능 |
| Use Case | Dashboard, Reports | Alert Editor, Settings |
| Mobile | 스크롤 필요 | 드롭다운으로 해결 |

## 💡 사용 시나리오

### ✅ TableSelectTemplate 사용하기 좋은 경우

1. **옵션이 5개 이상**: Select가 공간 효율적
2. **Alert/Notification Editor**: 설정 타입 선택
3. **Settings 페이지**: 설정 카테고리 전환
4. **Data Source 선택**: 여러 데이터 소스 중 선택
5. **Report Type 선택**: 다양한 리포트 타입

### ❌ TableTabsTemplate 사용하기 좋은 경우

1. **옵션이 3-5개 이하**: 한눈에 보이는 탭이 직관적
2. **Dashboard**: 데이터 뷰 전환 (Rules, History 등)
3. **User Management**: Admin, User 등 역할별 탭
4. **자주 전환하는 뷰**: 클릭 한 번으로 전환

## 🔧 내부 동작 원리

1. **Registration Phase**: `<Option>` 컴포넌트가 mount될 때 Context에 등록
2. **Context Storage**: `optionsData`에 `{label: children}` 형태로 저장
3. **Selection**: Select로 label 선택 → `selectedValue` 업데이트
4. **Rendering**: `optionsData[selectedValue]` 렌더링

## 🚀 확장 가능성

### 1. 다중 Select 지원 (Multi-Select)
```tsx
// 향후 추가 가능
<TableSelectTemplate mode="multi" defaultValues={['rule', 'notification']}>
```

### 2. Select 위치 커스터마이징
```tsx
// 향후 추가 가능
<TableSelectTemplate selectPosition="right">
```

### 3. 옵션별 아이콘
```tsx
// Option에 icon prop 추가
<Option value="rule" label="Alert Rules" icon={<Bell />}>
```

## 📖 Best Practices

1. **Label은 간결하게**: Select 드롭다운에 표시되므로 짧고 명확하게
2. **defaultValue 지정**: 첫 렌더링 시 깜빡임 방지
3. **leftItems는 필터/검색**: 데이터 조작 관련 UI
4. **rightItems는 액션**: Create, Export 등 주요 버튼
5. **disabled 옵션 활용**: Coming Soon 등 미구현 기능 표시

## 🐛 Troubleshooting

### 선택해도 화면이 바뀌지 않음
- `Option`의 `label`이 고유한지 확인
- Context Provider 안에서 사용하는지 확인

### Select가 보이지 않음
- `Options` 안에 최소 1개의 `Option` 필요
- `children`이 올바르게 전달되었는지 확인

### 타입 에러
- `Option`의 `value`, `label`이 모두 string인지 확인
- `children`이 ReactNode인지 확인

## 📚 Related Components

- **TableTabsTemplate**: Tab 기반 네비게이션
- **TableBasic/TablePagination/TableInfinity**: Body에 주입 가능한 Table 컴포넌트
- **Select** (shadcn): 내부에서 사용하는 Select 컴포넌트
