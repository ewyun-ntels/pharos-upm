### 1. Sheet 패턴 (사이드 패널)

**적용**: User Detail Sheet, Dashboard History Sheet

> **💡 코드 참고**: 실제 구현된 상세 코드는 `extensions/demo/frontend/src/routes/table-tabs-demo.tsx`의 `DemoUserSheetTypeA`, `DemoUserSheetTypeB` 컴포넌트 참고

#### 1.1 컴포넌트 계층 구조

```text
<SheetCustom>
  ├── <SheetCustomHeader>
  │   └── title, leftInfoGroup, rightButtonGroup (모두 선택적)
  │
  ├── [선택] <SheetCustomTabs>          ← 탭 사용 시에만 추가 (마운트 시 Header border-b 자동 제거)
  │
  ├── <SheetCustomContent>             ← 탭 있을 시 activeTab 조건부 렌더링
  │   └── <SheetCustomGroup>           ← 섹션 컨테이너 (title만 담당)
  │       ├── <SheetCustomGroupGrid>   ← Grid 레이아웃 영역 (cols=1|2)
  │       │   └── <SheetCustomGroupGridItem (내부 단일 항목 label + value + 선택적 description) />
  │       ├── <SheetCustomGroupTable>  ← Table 래퍼 영역
  │       ├── <SheetCustomGroupChart>  ← Chart 래퍼 영역 (단순 래퍼로 동작. 하위에 차트 배치)
  │       └── <SheetCustomGroupTab>    ← Tab 레이아웃 영역
  │
  └── <SheetCustomFooter>              ← 레이아웃 영역만 제공, 버튼은 children으로 구성
```

#### 1.2 헤더 버튼 배치 규칙

```
┌───────────────────────────────────────────────────────────────┐
│ [타이틀] [leftInfoGroup]                 [rightButtonGroup]   │
│ 제목+(정보, 상태 표시)                    액션 버튼 그룹        │
└───────────────────────────────────────────────────────────────┘
```

**왼쪽 그룹**
1. **타이틀** : 타이틀과 선택적 설명(`flex flex-col gap-1`로 묶음)
2. **leftInfoGroup** : 정보에 대한 내용 또는 상태에 대한 표현 (`<Badge>` 등)

**오른쪽 button 그룹**
1. **rightButtonGroup**: 액션에 대한 버튼 그룹 영역. 
<br> *단 X(Close) icon-button이 존재할 경우 — `variant="ghost"` (**제일 오른쪽**)*

#### 1.3 SheetCustomGroupGrid
키-값(Key-Value) 형태의 텍스트/데이터를 나열할 때 사용합니다.

- **`cols`**: `1` 또는 `2` 지원 (기본값: 2)
- 하위에 `SheetCustomGroupGridItem`을 배치하여 사용합니다.
  - `label`, `value`, `description`(선택) 속성 지원
  - 부모 그리드의 `cols`에 맞춰 `cols={1}`이면 좌우 가로 정렬(`flex-row justify-between`), `cols={2}`이면 상하 세로 정렬(`flex-col`)로 렌더링됩니다.

#### 1.4 SheetCustomGroupTable
`DataGrid` 등 테이블을 스크롤 이슈 없이 시트 내부에 배치할 때 사용합니다.

- 부모 영역의 스크롤에 영향을 주지 않도록 내부에 `overflow-hidden` 및 내부 스크롤바 컨테이너를 위한 `min-h-0` 클래스가 내장되어 있습니다.
- 자식 요소로 커스텀 테이블을 바로 배치하면 됩니다.

#### 1.5 Tab 영역 (SheetCustomGroupTab)
SheetCustomGroup 하위에 탭 인터페이스를 구성할 때 사용합니다. 차트뿐만 아니라 테이블, 그리드 등 다양한 컨텐츠를 탭으로 분리할 수 있습니다.

- `tabs={[...]}`, `activeTab`, `onTabChange` 제공
- 내부적으로 `shadcn/ui`의 `Tabs`를 자동 생성하며, 탭 높이(`h-8`)와 여백이 고정되어 디자인 일관성을 유지합니다.
- 사용자는 내부에 `<TabsContent>`만 작성하여 탭별 내용을 자유롭게 채웁니다.

#### 1.6 Chart 영역 (SheetCustomGroupChart)
차트를 포함하는 단순 래퍼 컴포넌트입니다.

- 단순 래퍼로 동작하며, 하위에 차트 컴포넌트를 바로 배치합니다.
- `gap-4`가 적용되어 있어 타이틀 및 다른 요소들과의 일관된 간격을 제공합니다.

#### 1.7 타입 C: JSON Editor (설정/로그 뷰어)
히스토리 내역의 설정 값 등 JSON 포맷 데이터를 직접 보여줄 때 사용합니다.

- **래퍼 설정**: `flex-1 overflow-hidden` (내부 에디터가 자체 스크롤을 갖도록 함, 패딩 없이 꽉 차게 배치)
- **에디터 설정**: `height="100%"`