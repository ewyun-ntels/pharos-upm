### Field 컴포넌트

> **💡 코드 참고**: `shared/frontend/src/components/ui-extension/form-field.tsx`

---

#### 1. 컴포넌트 아키텍처 (Prop-Driven)

모든 Field 컴포넌트(`FieldInput`, `FieldTextarea`, `FieldSelect`, `FieldSwitch`, `FieldCheckbox`)는 각자의 네이티브 엘리먼트(`Input`, `Textarea`, `select`, `Switch`, `Checkbox`)의 기본 속성을 모두 상속받으며, 추가로 다음의 확장 Prop을 지원합니다.

| Prop | 타입 | 설명 |
|---|---|---|
| `label` | `string` | 폼 필드의 제목 (라벨) |
| `description` | `string` | 라벨 하단에 표시되는 보조 안내 문구 |
| `required` | `boolean` | 필수 입력 여부 (라벨 옆에 `*` 표시) |
| `error` | `string \| boolean` | 에러 발생 여부 및 에러 메시지. `string` 전달 시 하단에 붉은색 텍스트 표출 (현재 `FieldInput`, `FieldTextarea` 적용) |
| `containerClassName` | `string` | 폼 요소 전체를 감싸는 최상단 부모 `div`의 Tailwind 클래스. 레이아웃 변경 시 사용 |

---

#### 2. 에러 핸들링 (shadcn 호환)

`FieldInput` 및 `FieldTextarea` 컴포넌트는 `error` prop을 통해 shadcn 스타일의 에러 UI를 네이티브하게 지원합니다.
- `error` 값이 존재하면 부모 컨테이너에 `data-invalid={true}` 속성이 부여됩니다.
- 내부 `<Input>` 또는 `<Textarea>` 엘리먼트에는 `aria-invalid={true}` 속성이 부여되어, 자동으로 `border-destructive` 및 붉은색 포커스 링 디자인이 적용됩니다.

---

#### 3. 레이아웃 제어 (세로/가로 배치)

기본적으로 모든 Field 컴포넌트는 **세로 배치(`flex-col`)**로 설계되어 있습니다.
`FieldSwitch` 등에서 **가로 배치(Row)**가 필요한 경우, `layout` 속성 대신 `containerClassName`을 활용해 Tailwind 유틸리티 클래스로 유연하게 덮어쓸 수 있습니다.

---

#### 4. 코드 사용 예시

```tsx
import { FieldInput, FieldTextarea, FieldSwitch, FieldCheckbox } from '@pharos/shared/components/ui-extension';

// 1. 일반 텍스트 입력 (에러 핸들링 포함)
<FieldInput
  label="Username"
  required
  description="영문자, 숫자 조합으로 입력해주세요."
  type="text"
  id="username"
  error={errors.username ? errors.username.message : undefined}
  {...register('username')}
/>

// 2. 스위치 토글 (가로 배치)
<FieldSwitch
  label="알림 설정"
  description="이메일로 알림을 수신합니다."
  checked={true}
  onCheckedChange={setAlerts}
  containerClassName="flex-row items-center justify-between gap-4"
/>

// 3. 체크박스
<FieldCheckbox
  label="이용 약관 동의"
  description="서비스 이용 약관에 동의합니다."
  id="terms"
  checked={agree}
  onCheckedChange={setAgree}
/>
```
