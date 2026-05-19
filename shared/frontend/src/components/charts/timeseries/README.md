# TimeSeries Chart Library

재사용 가능한 고성능 시계열 차트 컴포넌트입니다. uPlot 기반으로 대용량 데이터를 효율적으로 시각화합니다.

## � 설치 및 사용

### Import
```tsx
import { TimeSeriesChart } from '@pharos/shared/components/charts';
import type { TimeSeriesPanelOptions } from '@pharos/shared/components/charts';
```

### 기본 사용법

```tsx
import { TimeSeriesChart } from '@pharos/shared/components/charts';
import { useTranslation } from '@refinedev/core';
import { convertUnit, UnitType } from '@lib/unitUtils';
import { formatLocalTime24 } from '@lib/formatDate';

function MyChart() {
  const { translate } = useTranslation();

  return (
    <TimeSeriesChart
      chartWidth={800}
      chartHeight={400}
      dataProvider={{
        chartQuery: [
          { query: 'cpu_usage', legend: 'CPU' },
          { query: 'memory_usage', legend: 'Memory' }
        ],
        resource: 'metrics',
        dataProviderName: 'prometheus'
      }}
      options={{
        legendEnabled: true,
        legendAlign: 'right',
        chartType: 'line',
        unit: 'percent'
      }}
      pluginContext={pluginContext}
      formatLocalTime={formatLocalTime24}
      convertUnit={(value, unit) => convertUnit(value, unit as UnitType).toString()}
      translate={translate}
    />
  );
}
```

## 🎯 주요 용도

### 1. Dashboard Panel
```tsx
// core/frontend/src/features/panels/plugins/timeSeries/TimeSeriesCard.tsx
import { TimeSeriesChart } from '@pharos/shared/components/charts';

export function TimeSeriesCard(props: PanelProps) {
  const { translate } = useTranslation();
  
  return (
    <Card>
      <TimeSeriesChart
        {...props}
        chartWidth={chartWidth}
        chartHeight={chartHeight}
        formatLocalTime={formatLocalTime24}
        convertUnit={(v, u) => convertUnit(v, u as UnitType).toString()}
        translate={translate}
      />
    </Card>
  );
}
```

### 2. Alert Rule Preview
```tsx
function AlertPreview({ rule }) {
  return (
    <TimeSeriesChart
      chartWidth={800}
      chartHeight={300}
      dataProvider={{
        chartQuery: [{ query: rule.query, legend: rule.name }]
      }}
      options={{
        thresholds: [
          { value: rule.threshold, color: '#ff9800', type: 'dash' }
        ]
      }}
      pluginContext={pluginContext}
    />
  );
}
```

### 3. Standalone 차트 (Card 없이)
```tsx
function SimpleChart({ data }) {
  return (
    <div className="h-[400px]">
      <TimeSeriesChart
        chartWidth={800}
        chartHeight={400}
        data={data}
        options={{
          legendEnabled: true,
          chartType: 'area',
          fillOpacity: 0.3
        }}
      />
    </div>
  );
}
```

## 🔧 Props

### TimeSeriesChartProps

| Prop | Type | 설명 |
|------|------|------|
| `chartWidth` | `number?` | 차트 너비 (px) |
| `chartHeight` | `number?` | 차트 높이 (px) |
| `dataProvider` | `DataProvider?` | 데이터 제공자 설정 |
| `args` | `ChartQueryArgs?` | 쿼리 인자 (시간 범위, 필터 등) |
| `options` | `TimeSeriesPanelOptions?` | 차트 옵션 |
| `pluginContext` | `PluginContext?` | Plugin 실행 컨텍스트 |
| `formatLocalTime` | `(date: Date) => string` | 날짜 포맷 함수 |
| `convertUnit` | `(value: number, unit?: string) => string` | 단위 변환 함수 |
| `translate` | `(key: string) => string` | i18n 번역 함수 |
| `timeRangeCallback` | `(start, end) => void` | 시간 범위 선택 콜백 |

### TimeSeriesPanelOptions

| Option | Type | Default | 설명 |
|--------|------|---------|------|
| `legendEnabled` | `boolean?` | `false` | 범례 표시 |
| `legendAlign` | `'right' \| 'bottom'?` | `'right'` | 범례 위치 |
| `legendAsTable` | `boolean?` | `false` | 테이블 형식 범례 (통계 포함) |
| `showAverage` | `boolean?` | `false` | 평균값 표시 |
| `showLast` | `boolean?` | `false` | 마지막값 표시 |
| `showMax` | `boolean?` | `false` | 최댓값 표시 |
| `showMin` | `boolean?` | `false` | 최솟값 표시 |
| `showTotal` | `boolean?` | `false` | 합계 표시 |
| `chartType` | `'line' \| 'area'?` | `'line'` | 차트 타입 |
| `fillOpacity` | `number?` | `0.1` | Area 차트 투명도 (0-1) |
| `unit` | `string?` | `'short'` | 값 단위 |
| `thresholds` | `Threshold[]?` | - | 임계값 라인 |
| `syncId` | `string?` | - | 다중 차트 커서 동기화 ID |

## 📊 사용 예제

### 1. Area 차트 + 임계값
```tsx
<TimeSeriesChart
  chartWidth={800}
  chartHeight={400}
  dataProvider={dataProvider}
  options={{
    chartType: 'area',
    fillOpacity: 0.3,
    thresholds: [
      { value: '80', color: '#ff9800', type: 'dash' },
      { value: '95', color: '#f44336', type: 'solid' }
    ]
  }}
/>
```

### 2. 테이블 범례 + 통계
```tsx
<TimeSeriesChart
  chartWidth={800}
  chartHeight={400}
  dataProvider={dataProvider}
  options={{
    legendEnabled: true,
    legendAlign: 'right',
    legendAsTable: true,
    showLast: true,
    showMax: true,
    showMin: true,
    showAverage: true
  }}
/>
```

### 3. 다중 차트 동기화
```tsx
const syncId = 'my-sync-group';

<>
  <TimeSeriesChart
    options={{ syncId }}
    dataProvider={cpuData}
  />
  <TimeSeriesChart
    options={{ syncId }}
    dataProvider={memoryData}
  />
</>
```

### 4. 시간 범위 선택
```tsx
const [timeRange, setTimeRange] = useState();

<TimeSeriesChart
  dataProvider={dataProvider}
  timeRangeCallback={(start, end) => {
    setTimeRange({ start, end });
  }}
/>
```

### 5. 반응형 차트
```tsx
const containerRef = useRef<HTMLDivElement>(null);
const [dimensions, setDimensions] = useState({ width: 0, height: 0 });

useEffect(() => {
  const resizeObserver = new ResizeObserver(() => {
    if (containerRef.current) {
      setDimensions({
        width: containerRef.current.offsetWidth,
        height: containerRef.current.offsetHeight
      });
    }
  });
  resizeObserver.observe(containerRef.current);
  return () => resizeObserver.disconnect();
}, []);

<div ref={containerRef} className="w-full h-[400px]">
  <TimeSeriesChart
    chartWidth={dimensions.width}
    chartHeight={dimensions.height}
    dataProvider={dataProvider}
  />
</div>
```

## 🎨 주요 기능

- ✅ **고성능**: uPlot 기반, 대용량 데이터 렌더링
- ✅ **커스터마이징**: 스타일, 범례, 축 등 확장 가능
- ✅ **다양한 차트 타입**: Line, Area
- ✅ **임계값 라인**: 여러 개 설정 가능
- ✅ **대화형 범례**: 정렬, 필터링 지원
- ✅ **다중 차트 동기화**: Cursor sync
- ✅ **테마 지원**: Light/Dark 자동 지원
- ✅ **반응형**: 자동 크기 조정

## 🔄 마이그레이션 (기존 코드에서)

### Before
```tsx
import { TimeSeriesChart } from '@features/panels/plugins/timeSeries/TimeSeries';

<TimeSeriesChart {...props} />
```

### After
```tsx
import { TimeSeriesChart } from '@pharos/shared/components/charts';
import { useTranslation } from '@refinedev/core';
import { convertUnit, UnitType } from '@lib/unitUtils';
import { formatLocalTime24 } from '@lib/formatDate';

function MyComponent(props) {
  const { translate } = useTranslation();

  return (
    <TimeSeriesChart
      {...props}
      formatLocalTime={formatLocalTime24}
      convertUnit={(value, unit) => convertUnit(value, unit as UnitType).toString()}
      translate={translate}
    />
  );
}
```

**변경 사항**:
1. Import 경로 변경: `@pharos/shared/components/charts`
2. Utility 함수 주입 필요: `formatLocalTime`, `convertUnit`, `translate`

## 📝 데이터 형식

`pluginContext.hooks.useChartData`를 통해 다음 형식으로 데이터를 받습니다:

```typescript
{
  loading: false,
  chartData: {
    chartType: 'timeseries',
    startTime: 1234567890,
    endTime: 1234567890,
    step: 60,
    uniqueKeys: ['series1', 'series2'],
    chartMetric: [
      { timestamp: 1234567890, series1: 100, series2: 200 },
      { timestamp: 1234567950, series1: 110, series2: 210 },
      // ...
    ]
  }
}
```

## 🎨 색상 관리

- 17개 사전 정의 색상 팔레트
- Legend label 기반 deterministic 색상 할당 (murmurhash3 사용)
- 동일한 label은 항상 같은 색상 보장

## ⚡ 성능 팁

1. **Tooltip Limit**: 많은 시리즈가 있을 때 `tooltipLimit` 사용
2. **범례 위치**: Bottom 범례는 많은 시리즈에서 높이에 영향
3. **동기화 그룹**: 2-4개 차트 권장
4. **데이터 포인트**: 10k+ 포인트 처리 가능, 매우 큰 데이터는 다운샘플링 고려

## 📦 의존성

```json
{
  "uplot": "^1.6.31",
  "uplot-react": "^1.2.2",
  "murmurhash3js": "^3.0.1",
  "uuid": "^11.0.3"
}
```

## 🏗️ 아키텍처

```
shared/frontend/src/components/charts/timeseries/
├── TimeSeries.tsx          # Main chart component
├── TimeSeriesCardLayout.tsx # Card wrapper (optional)
├── types.ts                # Type definitions
├── utils.ts                # Color & sorting utils
└── index.ts                # Exports
```

**core에서 사용 시**:
```
core/frontend/src/features/panels/plugins/timeSeries/
├── TimeSeriesCard.tsx      # Dashboard panel (Card + Data + Chart)
├── TimeSeriesWithData.tsx  # Data fetching wrapper
└── panel.plugin.ts         # Plugin registration
```

## 🎯 사용 패턴

### Pattern 1: 순수 차트만 (Pure Chart)
```tsx
import { TimeSeriesChart } from '@pharos/shared/components/charts';
<TimeSeriesChart data={myData} options={myOptions} />
```

### Pattern 2: Card + 차트 (Card Layout)
```tsx
import { TimeSeriesCardLayout, TimeSeriesChart } from '@pharos/shared/components/charts';
<TimeSeriesCardLayout title="My Chart">
  <TimeSeriesChart data={myData} options={myOptions} />
</TimeSeriesCardLayout>
```

### Pattern 3: Dashboard Panel (core에서만)
```tsx
import { TimeSeriesCard } from '@features/panels/plugins/timeSeries/TimeSeriesCard';
<TimeSeriesCard {...panelProps} />
```

## 📄 타입 Exports

```typescript
import type {
  TimeSeriesChartProps,
  TimeSeriesPanelOptions,
  ChartQueryArgs,
  DataProvider,
  Threshold,
  Scales
} from '@pharos/shared/components/charts';
```

## 🤝 기여

이 라이브러리는 Pharos 프로젝트의 일부입니다.

---

**더 많은 예제가 필요하면 코드를 확인하세요**:
- Dashboard Panel: `core/frontend/src/features/panels/plugins/timeSeries/`
- Custom Health: `core/frontend/src/features/panels/plugins/custom_health/health.tsx`
- Bar Gauge: `core/frontend/src/features/panels/plugins/barGauge/tableSizeSheet.tsx`
