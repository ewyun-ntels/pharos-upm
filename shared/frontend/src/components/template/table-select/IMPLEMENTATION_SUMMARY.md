# TableSelectTemplate 구현 완료 🎉

## ✅ 완료된 작업

### 1. 새로운 TableSelectTemplate 컴포넌트
```
components/template/table-select/
├── shared/
│   ├── TableSelectTemplate.tsx    # ✅ 메인 템플릿 컴포넌트
│   └── types.ts                   # ✅ TypeScript types
├── index.ts                       # ✅ Public exports
└── README.md                      # ✅ 상세 문서
```

**핵심 기능:**
- ✅ Select 드롭다운 기반 네비게이션
- ✅ 유연한 Children 주입 (TableBasic, TablePagination 등)
- ✅ 왼쪽/오른쪽 커스텀 아이템 배치
- ✅ Context API로 선택 상태 관리
- ✅ 옵션 비활성화 지원
- ✅ TypeScript 완벽 지원

### 2. Demo 페이지 생성
```
app/table-select-demo/
├── page.tsx      # ✅ 3가지 실제 사용 예제
└── layout.tsx    # ✅ BaseLayout 래핑
```

**Demo 구성:**
1. **Alert Configuration Editor** (600px 높이)
   - Alert Rules 탭: 12개 Alert Rule 데이터
   - Notifications 탭: 8개 Notification 데이터
   - Alert History 탭: Coming Soon 플레이스홀더
   - 왼쪽: 검색 + Severity 필터
   - 오른쪽: Save Draft + Create New 버튼

2. **Dashboard Management** (600px 높이)
   - Personal 탭: 15개 개인 대시보드 (페이징)
   - Team 탭: 10개 팀 대시보드
   - Public 탭: 20개 공개 대시보드 (페이징)
   - 왼쪽: 검색 + Owner 필터
   - 오른쪽: Export + Import + New Dashboard 버튼

3. **Simple Example** (400px 높이)
   - Email/Slack/Webhook 타입별 Notification 목록
   - 최소 설정으로 사용 예시

### 3. _refine_context.tsx 등록
```tsx
{
  name: 'table-select-demo',
  list: '/table-select-demo',
  meta: {
    label: 'Table Select Demo',
    icon: <LayoutGrid className={'h-4 w-4'} />,
    dataProviderName: 'dataProvider',
    canDelete: false,
  },
}
```

## 🎨 UI/UX 특징

### 레이아웃 구조
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

### 왼쪽 정렬 배치
- **Left Custom Items**: 검색, 필터 등 (gap-2)
- **Select 드롭다운**: leftItems 오른쪽에 배치
- **Right Custom Items**: 최우측 (justify-between)
- **Responsive**: 양쪽 정렬 유지

## 📊 기능 비교

| Feature | TableTabsTemplate | TableSelectTemplate |
|---------|-------------------|---------------------|
| **Navigation** | Tabs (수평 버튼) | Select (드롭다운) |
| **공간 효율** | 많은 공간 차지 | 컴팩트 |
| **옵션 수** | 3-5개 권장 | 10개 이상 가능 |
| **Use Case** | Dashboard, Reports | Alert Editor, Settings |
| **Mobile** | 스크롤 필요 | 드롭다운으로 해결 |
| **왼쪽 정렬** | Tab만 왼쪽 | Select + Left Items 왼쪽 |
| **커스텀 아이템** | Right만 | Left + Right |

## 🚀 사용 예시

### 1. Alert Editor (실제 사용)
```tsx
import {TableSelectTemplate, Options, Option} from '@components/template/table-select';
import {AlertRuleEditor} from '@features/alert/components/alert-rule-editor';
import {NotificationEditor} from '@features/alert/components/notification-editor';

export default function AlertEditorPage() {
  return (
    <TableSelectTemplate
      title="Alert Configuration"
      description="Configure alert rules and notifications"
      selectLabel="Editor Type"
      selectPlaceholder="Select editor type"
      selectWidth="180px"
      defaultValue="Alert Rules"
      leftCustomItems={
        <>
          <Input placeholder="Search..." />
          <Select>...</Select>  {/* Severity 필터 */}
        </>
      }
      rightCustomItems={
        <>
          <Button>Save Draft</Button>
          <Button>Create New</Button>
        </>
      }
    >
      <Options>
        <Option value="rule" label="Alert Rules">
          <AlertRuleEditor />
        </Option>
        <Option value="notification" label="Notifications">
          <NotificationEditor />
        </Option>
      </Options>
    </TableSelectTemplate>
  );
}
```

### 2. Notification Editor (실제 사용)
```tsx
import {TableSelectTemplate, Options, Option} from '@components/template/table-select';
import {EmailEditor} from '@features/notification/components/email-editor';
import {SlackEditor} from '@features/notification/components/slack-editor';
import {WebhookEditor} from '@features/notification/components/webhook-editor';

export default function NotificationEditorPage() {
  return (
    <TableSelectTemplate
      title="Notification Settings"
      description="Configure notification channels"
      selectLabel="Channel Type"
      selectPlaceholder="Select channel"
      selectWidth="200px"
      defaultValue="Email"
      rightCustomItems={
        <Button>Test Connection</Button>
      }
    >
      <Options>
        <Option value="email" label="Email">
          <EmailEditor />
        </Option>
        <Option value="slack" label="Slack">
          <SlackEditor />
        </Option>
        <Option value="webhook" label="Webhook">
          <WebhookEditor />
        </Option>
      </Options>
    </TableSelectTemplate>
  );
}
```

### 3. 간단한 사용
```tsx
<TableSelectTemplate
  title="User Management"
  selectPlaceholder="Select user type"
>
  <Options>
    <Option value="admin" label="Administrators">
      <TableBasic data={adminUsers} columns={columns} tableKey="admin" />
    </Option>
    <Option value="regular" label="Regular Users">
      <TableBasic data={regularUsers} columns={columns} tableKey="regular" />
    </Option>
  </Options>
</TableSelectTemplate>
```

## 🎯 API Reference

### TableSelectTemplate Props

| Prop | Type | Default | 설명 |
|------|------|---------|------|
| `title` | `string` | `undefined` | 페이지 제목 |
| `description` | `string` | `undefined` | 페이지 설명 |
| `children` | `React.ReactNode` | **Required** | `<Options>` 컴포넌트 |
| `defaultValue` | `string` | 첫 번째 옵션 | 기본 선택값 |
| `className` | `string` | `undefined` | CSS 클래스 |
| `onSelectionChange` | `(value: string) => void` | `undefined` | 선택 변경 콜백 |
| `selectPlaceholder` | `string` | `'Select an option'` | Select placeholder |
| `selectLabel` | `string` | `undefined` | Select 그룹 레이블 |
| `selectWidth` | `string` | `'200px'` | Select 너비 |
| `leftCustomItems` | `React.ReactNode` | `undefined` | 왼쪽 커스텀 아이템 |
| `rightCustomItems` | `React.ReactNode` | `undefined` | 오른쪽 커스텀 아이템 |

### Option Props

| Prop | Type | Default | 설명 |
|------|------|---------|------|
| `value` | `string` | **Required** | 옵션 고유 식별자 |
| `label` | `string` | **Required** | Select에 표시될 레이블 |
| `children` | `React.ReactNode` | **Required** | 옵션 선택 시 렌더링될 컨텐츠 |
| `disabled` | `boolean` | `false` | 옵션 비활성화 여부 |

## 💡 사용 시나리오

### ✅ TableSelectTemplate 사용하기 좋은 경우

1. **Alert Editor**: Alert Rule / Notification 설정 전환
2. **Notification Editor**: Email / Slack / Webhook 타입 전환
3. **Settings 페이지**: General / Security / Notifications 등
4. **Data Source 선택**: 여러 데이터 소스 중 선택
5. **Report Type 선택**: Daily / Weekly / Monthly 등
6. **옵션이 5개 이상**: Select가 공간 효율적

### ❌ TableTabsTemplate 사용하기 좋은 경우

1. **Dashboard**: Rules / History 탭 (자주 전환)
2. **User Management**: Admin / User 탭
3. **옵션이 3-5개 이하**: 한눈에 보이는 탭이 직관적
4. **자주 전환하는 뷰**: 클릭 한 번으로 전환

## 🔄 TableTabs vs TableSelect

### TableTabsTemplate (기존)
```tsx
<TableTabsTemplate name="Dashboard">
  <Tabs>
    <Tab name="Rules">...</Tab>
    <Tab name="History">...</Tab>
  </Tabs>
</TableTabsTemplate>
```
- ✅ 한눈에 모든 탭 보임
- ✅ 클릭 한 번으로 전환
- ❌ 많은 수평 공간 차지
- ❌ 5개 이상은 스크롤

### TableSelectTemplate (신규)
```tsx
<TableSelectTemplate title="Alert Editor">
  <Options>
    <Option value="rule" label="Rules">...</Option>
    <Option value="notif" label="Notifications">...</Option>
  </Options>
</TableSelectTemplate>
```
- ✅ 컴팩트 (Select 드롭다운)
- ✅ 10개 이상도 문제없음
- ✅ Left/Right 커스텀 아이템
- ❌ 클릭 2번 필요 (드롭다운 열기 → 선택)

## 📝 실제 적용 예시

### Alert 페이지에 적용
```tsx
// app/alert/editor/page.tsx
import {TableSelectTemplate, Options, Option} from '@components/template/table-select';

export default function AlertEditorPage() {
  return (
    <TableSelectTemplate
      title="Alert Editor"
      selectLabel="Editor Type"
      selectWidth="180px"
      defaultValue="Alert Rules"
      leftCustomItems={<Input placeholder="Search..." />}
      rightCustomItems={<Button>Create New</Button>}
    >
      <Options>
        <Option value="rule" label="Alert Rules">
          <AlertRuleForm />
        </Option>
        <Option value="notification" label="Notifications">
          <NotificationForm />
        </Option>
      </Options>
    </TableSelectTemplate>
  );
}
```

### Notification 페이지에 적용
```tsx
// app/notification/page.tsx
import {TableSelectTemplate, Options, Option} from '@components/template/table-select';

export default function NotificationPage() {
  return (
    <TableSelectTemplate
      title="Notification Settings"
      selectPlaceholder="Select channel type"
      selectWidth="220px"
    >
      <Options>
        <Option value="email" label="Email">
          <EmailNotificationTable />
        </Option>
        <Option value="slack" label="Slack">
          <SlackNotificationTable />
        </Option>
        <Option value="webhook" label="Webhook">
          <WebhookNotificationTable />
        </Option>
      </Options>
    </TableSelectTemplate>
  );
}
```

## 🎉 결론

TableSelectTemplate이 성공적으로 구현되었습니다!

**주요 성과:**
- ✅ Select 기반 유연한 템플릿
- ✅ 왼쪽 정렬 (Left Items + Select)
- ✅ 오른쪽 커스텀 아이템
- ✅ Body에 Children 주입
- ✅ 3가지 실제 예제 포함 Demo 페이지
- ✅ 상세한 README 문서
- ✅ TypeScript 완벽 지원
- ✅ Context API로 확장 가능

**사용 가능한 곳:**
- Alert Editor ✅
- Notification Editor ✅
- Settings 페이지
- Data Source 선택
- Report Type 선택

**URL:**
- Demo: http://localhost:3000/table-select-demo

이제 Alert Editor와 Notification Editor에 바로 적용할 수 있습니다! 🚀
