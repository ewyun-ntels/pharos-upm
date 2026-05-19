/**
 * Example Extension Panel
 */

'use client';

import React from 'react';
import type { PanelEditorOptionsProps, PanelProps } from '@pharos/core/panel-registry';
import { ChartDataWrapper } from '@pharos/core/panel-registry';

export interface ExampleChartOptions {
  title?: string;
  showLegend?: boolean;
  chartColor?: string;
  lineWidth?: number;
  showDataPoints?: boolean;
}

type ExampleChartPanelProps = PanelProps<ExampleChartOptions>;

export function ExampleChartPanel(props: ExampleChartPanelProps) {
  const {
    title = 'Example Chart',
    showLegend = true,
    chartColor = '#3b82f6',
    lineWidth = 2,
    showDataPoints = false,
  } = props.options || {};

  const primaryDataProvider = Array.isArray(props.dataProvider)
    ? props.dataProvider[0]
    : props.dataProvider;

  // 데이터 소스 미설정 체크
  if (!primaryDataProvider?.chartQuery || primaryDataProvider.chartQuery.length === 0) {
    return (
      <div style={{ padding: '16px', height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <div style={{ textAlign: 'center', color: '#6b7280' }}>
          <div style={{ fontSize: '48px', marginBottom: '8px' }}>📊</div>
          <div style={{ fontSize: '14px' }}>No data source configured</div>
        </div>
      </div>
    );
  }

  return (
    <ChartDataWrapper
      pluginContext={props.pluginContext}
      dataProvider={primaryDataProvider}
      args={props.args || new Map()}
      refetchInterval={props.refetchInterval}
      dashboardId={props.dashboardId}
      id={props.id}
    >
      {(chartData) => {
        const dataSourceCount = primaryDataProvider.chartQuery.length;
        const resourceName = primaryDataProvider.resource || 'metrics';
        const hasData = chartData?.chartMetric && chartData.chartMetric.length > 0;
        const dataPointCount = chartData?.chartMetric?.length || 0;
        const uniqueMetrics = chartData?.uniqueKeys || [];

        return (
          <div style={{ padding: '16px', height: '100%', display: 'flex', flexDirection: 'column', gap: '16px' }}>
            <div style={{ borderBottom: `${lineWidth}px solid ${chartColor}`, paddingBottom: '8px' }}>
              <h3 style={{ fontSize: '18px', fontWeight: '600', color: chartColor, margin: 0 }}>
                {props.title || title}
              </h3>
            </div>

            <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', backgroundColor: '#f9fafb', borderRadius: '8px', padding: '32px', border: `2px solid ${chartColor}` }}>
              <div style={{ textAlign: 'center' }}>
                <div style={{ width: '120px', height: '120px', margin: '0 auto 16px', borderRadius: '50%', backgroundColor: `${chartColor}20`, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '48px' }}>
                  {hasData ? '📈' : '📊'}
                </div>
                <h4 style={{ fontSize: '20px', fontWeight: '600', color: chartColor, marginBottom: '8px' }}>
                  Example Extension Panel
                </h4>
                <p style={{ fontSize: '14px', color: '#6b7280', marginBottom: '16px' }}>
                  {hasData ? 'Data loaded from sample-provider' : 'No data available'}
                </p>
                <div style={{ display: 'inline-block', padding: '8px 16px', backgroundColor: 'white', border: `1px solid ${chartColor}`, borderRadius: '8px', fontSize: '12px' }}>
                  <div style={{ color: '#6b7280' }}>Data Sources: <strong style={{ color: chartColor }}>{dataSourceCount}</strong></div>
                  <div style={{ color: '#6b7280', marginTop: '4px' }}>Resource: <strong style={{ color: chartColor }}>{resourceName}</strong></div>
                  {hasData && (
                    <>
                      <div style={{ color: '#6b7280', marginTop: '4px' }}>Data Points: <strong style={{ color: chartColor }}>{dataPointCount}</strong></div>
                      <div style={{ color: '#6b7280', marginTop: '4px' }}>Metrics: <strong style={{ color: chartColor }}>{uniqueMetrics.join(', ')}</strong></div>
                    </>
                  )}
                </div>
              </div>
            </div>

            {showLegend && (
              <div style={{ padding: '12px', backgroundColor: '#eff6ff', border: '1px solid #bfdbfe', borderRadius: '8px', display: 'flex', alignItems: 'center', gap: '8px' }}>
                <span style={{ fontSize: '16px' }}>ℹ️</span>
                <div style={{ flex: 1 }}>
                  <p style={{ fontSize: '12px', color: '#1e40af', fontWeight: '600', margin: 0 }}>Extension Panel Info</p>
                  <p style={{ fontSize: '11px', color: '#3b82f6', margin: 0, marginTop: '2px' }}>
                    Line Width: {lineWidth}px • Data Points: {showDataPoints ? 'Enabled' : 'Disabled'}
                  </p>
                  {hasData && chartData && (
                    <p style={{ fontSize: '11px', color: '#3b82f6', margin: 0, marginTop: '4px' }}>
                      Provider: {primaryDataProvider?.dataProviderName || 'default'} •
                      Time Range: {new Date(chartData.startTime || 0).toLocaleTimeString()} - {new Date(chartData.endTime || 0).toLocaleTimeString()}
                    </p>
                  )}
                </div>
              </div>
            )}

            {/* 실제 데이터 표시 (개발용) */}
            {hasData && showDataPoints && chartData?.chartMetric && (
              <div style={{ padding: '12px', backgroundColor: '#f0fdf4', border: '1px solid #86efac', borderRadius: '8px', maxHeight: '200px', overflowY: 'auto' }}>
                <p style={{ fontSize: '12px', fontWeight: '600', color: '#166534', marginBottom: '8px' }}>Sample Data (First 3 entries):</p>
                <pre style={{ fontSize: '10px', color: '#15803d', margin: 0, whiteSpace: 'pre-wrap' }}>
                  {JSON.stringify(chartData.chartMetric.slice(0, 3), null, 2)}
                </pre>
              </div>
            )}
          </div>
        );
      }}
    </ChartDataWrapper>
  );
}

export function ExampleChartPanelOptions({ options = {}, onOptionsChange }: PanelEditorOptionsProps<ExampleChartOptions>) {
  const chartTitle = options.title || 'Example Chart';
  const chartColor = options.chartColor || '#3b82f6';
  const lineWidth = options.lineWidth || 2;
  const showLegend = options.showLegend ?? true;
  const showDataPoints = options.showDataPoints ?? false;

  const handleChange = (field: keyof ExampleChartOptions, value: any) => {
    if (onOptionsChange) {
      onOptionsChange({ ...options, [field]: value });
    }
  };

  return (
    <div style={{ padding: '16px' }}>
      <h4 style={{ fontSize: '16px', fontWeight: '600', marginBottom: '16px' }}>Chart Settings</h4>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
        <div>
          <label htmlFor="chartTitle" style={{ display: 'block', fontSize: '13px', marginBottom: '4px' }}>Title</label>
          <input id="chartTitle" type="text" value={chartTitle} onChange={(e) => handleChange('title', e.target.value)} style={{ width: '100%', padding: '8px', border: '1px solid #d1d5db', borderRadius: '6px' }} />
        </div>
        <div>
          <label htmlFor="chartColor" style={{ display: 'block', fontSize: '13px', marginBottom: '4px' }}>Color</label>
          <input id="chartColor" type="color" value={chartColor} onChange={(e) => handleChange('chartColor', e.target.value)} style={{ width: '100%', height: '40px', border: '1px solid #d1d5db', borderRadius: '6px' }} />
        </div>
        <div>
          <label htmlFor="lineWidth" style={{ display: 'block', fontSize: '13px', marginBottom: '4px' }}>Line Width</label>
          <input id="lineWidth" type="number" min="1" max="10" value={lineWidth} onChange={(e) => handleChange('lineWidth', Number(e.target.value))} style={{ width: '100%', padding: '8px', border: '1px solid #d1d5db', borderRadius: '6px' }} />
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <input id="showLegend" type="checkbox" checked={showLegend} onChange={(e) => handleChange('showLegend', e.target.checked)} />
          <label htmlFor="showLegend">Show Legend</label>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <input id="showDataPoints" type="checkbox" checked={showDataPoints} onChange={(e) => handleChange('showDataPoints', e.target.checked)} />
          <label htmlFor="showDataPoints">Show Data Points</label>
        </div>
      </div>
    </div>
  );
}
