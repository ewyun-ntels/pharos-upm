## 1. 페이지 패턴별 레이아웃 규격

**적용 범위**: Dashboard List, User List, Notification List, Role Config

```
구조:
main (flex flex-col w-full h-full)
└── TableTabsTemplate
    ├── header (flex flex-col items-start w-full gap-1 pt-3 pb-2)
    │   ├── title row (flex justify-between items-end px-5)
    │   │   ├── left (flex flex-col gap-1)
    │   │   │   ├── Breadcrumb (mb-1)
    │   │   │   └── Title — 페이지 이름
    │   │   └── [선택적] right (flex items-center gap-2)
    │   │       └── 헤더 우측 button
    │   └── [선택적] Tabs (w-full mt-2 border-b border-border)
    └── content (flex-1 min-h-0 overflow-y-auto)
        └── div (px-5)
                └── DataGrid (flex flex-col)
                    ├── filters (flex flex-col gap-2 pt-3 pb-4)
                    │   └── filter row (flex items-center justify-between gap-2)
                    │       ├── left filters (flex items-center gap-2)
                    │       └── right filters (flex items-center gap-2 ml-auto)
                    ├── Table wrapper (relative)
                    │   ├── border overlay (absolute inset-0 rounded-[var(--radius)] border pointer-events-none z-20)
                    │   └── table (w-full text-sm border-separate border-spacing-0, tableLayout: fixed)
                    │       ├── thead (sticky top-0 z-10)
                    │       │   └── tr (bg-muted/60 [&>th]:border-b [&>th]:border-r)
                    │       │       ├── th (h-8 text-[13px] font-semibold) · th · ...
                    │       │       └── [선택적] spacer th (나머지 공간 채움)
                    │       └── tbody
                    │           ├── tr (bg-background hover:bg-muted/30 [&>td]:border-b)
                    │           │   ├── td (py-1 px-3 border-r overflow-hidden text-ellipsis)
                    │           │   └── [선택적] spacer td
                    │           └── (빈 상태) tr → td (colSpan 전체, py-10 text-center)
                    └── [선택적] Pagination wrapper (mt-3)
                        └── Pagination (flex items-center justify-between)
                        ├── left (flex items-center gap-3)
                        │   ├── [선택적] PageSize Select (w-17.5)
                        │   └── RowCount (text-xs text-muted-foreground)
                        └── right
                            ├── default: (flex items-center gap-4)
                            │   ├── "1 / 5" (text-xs)
                            │   └── nav buttons (flex items-center gap-1)
                            │       └── ◁◁ ◁ ▷ ▷▷ (ghost, h-7 w-7)
                            └── numbered: (flex flex-row items-center gap-1)
                                └── ◁ [1] [2] ... [5] ▷ (w-8 h-8)
```

**선택적 영역 조건:**

| 영역 | 조건 |
|------|------|
| `right` | 헤더에 액션 button이 필요할 때 |
| `tab nav` | 데이터 분류가 2개 이상일 때 |
| `spacer th/td` | `tableWidthMode="spacer"` (기본) |
| `Pagination` | 페이지네이션이 필요한 목록 |

#### 테이블 상단 필터 배치 규칙

```
┌───────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ [왼쪽 필터 그룹]                                              [오른쪽 버튼 그룹]                                │
│ Search > Single Select > Multi/Badge Select > Date Picker    Delete > action Button > main button(primary) │
└───────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

**왼쪽 필터 그룹** (순서: 넓은 범위에서 좁은 범위 순으로 배치)
1. **Search** — 검색
2. **Single Select** — 카테고리/대분류
3. **Multi/Badge Select** — 카테고리/속성/상태 필터
4. **Date Picker** — 시간 범위

**오른쪽 button 그룹** (순서: 왼쪽→오른쪽, 페이지에서 button(Primary)가 제일 오른쪽):
1. **Icon Button** : Delete — `variant="destructive"`(**제일 왼쪽**)
2. **action Button**: 새로고침 — `variant="outline"` Export — `variant="secondary"` 등 (**중앙**)
3. **page에서 main button (Primary)**: 생성/추가 (아이콘+텍스트) — `variant="default"` (**제일 오른쪽**)

---