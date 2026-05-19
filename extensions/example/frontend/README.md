# Example Extension - 완전한 예제

이 extension은 **일반 HTML/CSS만 사용**하여 외부 패널을 만드는 방법을 보여줍니다.

## 📦 구성

```
extensions/example/frontend/
├── src/
│   └── panels/
│       └── ExampleChartPanel.tsx    # Chart Panel + Options Panel
└── package.json
```

## 🎯 주요 기능

### 1. Chart Panel (ExampleChartPanel)
- ✅ pluginContext를 통한 데이터 fetching
- ✅ chartOptions를 통한 설정값 적용
- ✅ 일반 HTML/CSS만 사용 (외부 의존성 없음)

### 2. Options Panel (ExampleChartPanelOptions)
- ✅ react-hook-form 통합
- ✅ 일반 HTML input/checkbox/color picker 사용
- ✅ chartOptions로 자동 저장/전달

## 💡 사용 방법

### Chart Panel 컴포넌트

```typescript
export function ExampleChartPanel({
  chartOptions,  // ← Options Panel에서 설정한 값
  pluginContext, // ← core에서 주입한 hooks
  query,
  args,
}: Props) {
  // chartOptions 사용
  const title = chartOptions?.title || 'Default Title';
  const showLegend = chartOptions?.showLegend ?? true;
  
  // pluginContext로 데이터 fetching
  const { loading, chartData } = pluginContext.hooks.useChartData({
    queries: [query],
    args,
    resource: 'metrics',
    dataProviderName: 'dataSourceProvider',
  });

  return (
    <div style={{ padding: '16px' }}>
      <h3 style={{ color: chartOptions?.chartColor }}>
        {title}
      </h3>
      {/* 차트 렌더링 */}
    </div>
  );
}
```

### Options Panel 컴포넌트

```typescript
export function ExampleChartPanelOptions({ context }) {
  const { register } = context;
  
  return (
    <div>
      {/* Title */}
      <input 
        {...register('chartOptions.title')}
        placeholder="Enter title..."
      />
      
      {/* Show Legend */}
      <input 
        type="checkbox"
        {...register('chartOptions.showLegend')}
      />
      
      {/* Color Picker */}
      <input 
        type="color"
        {...register('chartOptions.chartColor')}
      />
    </div>
  );
}
```

## 📊 데이터 흐름

```
1. Options Panel에서 설정
   ↓
   chartOptions = {
     title: "My Chart",
     showLegend: true,
     chartColor: "#3b82f6"
   }
   ↓
2. DB에 자동 저장
   ↓
3. Chart Panel로 자동 전달
   ↓
4. Chart Panel에서 chartOptions 사용
```

## 🎨 스타일링

**Inline Styles 사용** (권장):
```typescript
<div style={{
  padding: '16px',
  backgroundColor: '#f3f4f6',
  borderRadius: '8px',
}}>
  ...
</div>
```

**장점**:
- ✅ 외부 CSS 파일 불필요
- ✅ CSS-in-JS 라이브러리 불필요
- ✅ TypeScript 타입 체크
- ✅ 동적 스타일 적용 가능

## 🔧 사용 가능한 기능

### pluginContext.hooks

```typescript
// 데이터 fetching
const { data } = pluginContext.hooks.useList({...});
const { data } = pluginContext.hooks.useOne({...});
const { data } = pluginContext.hooks.useCustom({...});

// 차트 데이터
const { loading, chartData } = pluginContext.hooks.useChartData({...});

// CRUD
const { mutate } = pluginContext.hooks.useCreate();
const { mutate } = pluginContext.hooks.useUpdate();
const { mutate } = pluginContext.hooks.useDelete();
```

## 📝 타입 정의

```typescript
interface ChartOptions {
  title?: string;
  showLegend?: boolean;
  chartColor?: string;
  lineWidth?: number;
  showDataPoints?: boolean;
}

interface Props {
  chartOptions?: ChartOptions;
  pluginContext?: PluginContext;
  query?: any;
  args?: any;
  resource?: string;
  dataProviderName?: string;
}
```

## 🚀 빌드 및 테스트

```bash
# TypeScript 타입 체크
cd extensions/example/frontend
pnpm tsc --noEmit

# Extension 개발
# (core에서 자동으로 로드됨)
```

## ⚠️ 주의사항

### 사용 가능한 것
- ✅ 일반 HTML 태그 (div, input, select 등)
- ✅ Inline styles
- ✅ react-hook-form (register, watch 등)
- ✅ pluginContext hooks
- ✅ chartOptions props

### 사용 불가능한 것
- ❌ core의 UI 컴포넌트 (InputField, SelectField 등)
- ❌ Path alias (@components, @features 등)
- ❌ core 내부 유틸리티 (cn 함수 등)

## 💡 실전 팁

### 1. 색상 Picker 사용
```typescript
<input 
  type="color" 
  {...register('chartOptions.chartColor')}
/>
```

### 2. 숫자 입력
```typescript
<input 
  type="number"
  min="1"
  max="10"
  {...register('chartOptions.lineWidth', { valueAsNumber: true })}
/>
```

### 3. Checkbox 그룹
```typescript
<input 
  type="checkbox"
  {...register('chartOptions.showLegend')}
/>
```

### 4. Select 옵션
```typescript
<select {...register('chartOptions.unit')}>
  <option value="short">Short</option>
  <option value="bytes">Bytes</option>
  <option value="percent">Percent</option>
</select>
```

## 🎉 결과

이 예제를 실행하면:
1. Options Panel에서 설정 변경
2. 자동으로 DB에 저장
3. Chart Panel에 즉시 반영
4. 완전히 작동하는 Extension 완성!

---

**다음 단계**: 실제 데이터와 더 복잡한 차트 로직 추가하기
