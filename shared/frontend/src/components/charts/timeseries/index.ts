/**
 * TimeSeries Chart Library
 * 
 * Pure, standalone time series chart component based on uPlot.
 * Accepts data externally - no data fetching logic inside.
 * 
 * @example
 * ```tsx
 * import { TimeSeriesChart, ChartMetricData } from '@pharos/shared/components/charts';
 * 
 * // Prepare data externally
 * const chartData: ChartMetricData = {
 *   chartMetric: [...],
 *   startTime: 1234567890,
 *   endTime: 1234567900,
 *   step: 10,
 *   chartType: 'line',
 *   uniqueKeys: ['cpu', 'memory']
 * };
 * 
 * // Pure rendering
 * <TimeSeriesChart
 *   data={chartData}
 *   loading={false}
 *   chartWidth={800}
 *   chartHeight={400}
 *   options={{
 *     legendEnabled: true,
 *     chartType: 'line',
 *     unit: 'short'
 *   }}
 *   formatLocalTime={(date) => date.toLocaleString()}
 *   convertUnit={(value, unit) => `${value} ${unit}`}
 *   translate={(key) => i18n.t(key)}
 * />
 * ```
 */

export { TimeSeriesChart, legendAsTooltipPlugin, TIMESERIES_DEFAULT_SYNC_KEY } from './TimeSeries';
export { TimeSeriesCardLayout } from './TimeSeriesCardLayout';
export type {
  TimeSeriesChartProps,
  TimeSeriesPanelOptions,
  ChartMetricData,
  ChartMetric,
  Threshold,
  Scales
} from './types';
export type { TimeSeriesCardLayoutProps } from './TimeSeriesCardLayout';
export { TimeSeriesPanelOptionsSchema } from './types';
export { getSortValue, createUniqueColorManager, getColorByKey } from './utils';
