### Edit 페이지 레이아웃

> **💡 코드 참고**: `shared/frontend/src/components/ui-extension/form-field.tsx`

---

#### 1. 페이지 전체 구조

Edit 페이지는 스크롤 가능한 컨텐츠 영역 안에 섹션 단위로 Form 필드를 배치하는 구조입니다.

```
main (w-full h-full)
└── ScrollArea / main (overflow-y-auto)
    ├── PageHeader
    └── section (flex flex-col w-full px-5 pt-5 pb-5)
        └── form container (max-w-3xl space-y-6)
            ├── [Section A] (space-y-4)
            │   ├── 섹션 제목 (Title variant="h3")
            │   └── fields (space-y-4)
            │       └── <FieldInput> / <FieldSwitch> 등 ... (반복)
            ├── [Section B] (space-y-4) ...반복
            └── ButtonGroup (flex justify-start gap-2 mt-6)
```

---

#### 2. 섹션 구성

섹션은 `Title variant="h3"`으로 제목을 표기하고, 그 아래 `space-y-4`로 필드를 나열합니다.

```tsx
{/* 섹션 A */}
<div className="space-y-4">
  <Title variant="h3">Section Title</Title>
  <div className="space-y-4">
    <FieldInput
      label="Field Name"
      required
      id="field-name"
    />
    <FieldSwitch
      label="Enable Feature"
      containerClassName="flex-row items-center justify-between gap-4"
    />
    {/* ...필드 반복 */}
  </div>
</div>

{/* 섹션 B */}
<div className="space-y-4">
  <Title variant="h3">Another Section</Title>
  ...
</div>
```



#### 3. Form 필드 구성 (Prop-Driven)

섹션 내의 각 폼 필드는 단일 컴포넌트(`FieldInput`, `FieldTextarea`, `FieldSelect`, `FieldSwitch`, `FieldCheckbox`)에 prop(`label`, `required`, `error` 등)을 전달하여 구성합니다.
가로 배치(Row)가 필요한 경우 `containerClassName` 속성을 통해 커스텀 Tailwind 클래스를 주입합니다. (자세한 내용은 `form.md` 참고)



#### 4. ButtonGroup

버튼 영역은 `ButtonGroup`으로 표현하며, 영역(레이아웃)만 제공합니다. 버튼의 구성은 사용자 몫입니다.

- 기본 배치: 좌측 정렬 (`flex justify-start gap-2 mt-6`)
- Cancel + Save가 있을 경우 **Cancel → Save** 순서로 배치합니다.

```
ButtonGroup (flex justify-start gap-2 mt-6)
├── Cancel  variant="outline"
└── Save    variant="default"
```
