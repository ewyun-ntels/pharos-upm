/**
 * Chart Panel Example
 *
 * withChartData HOC를 사용한 순수 렌더링 패널 예시.
 * - chartData는 HOC가 주입 (데이터 fetching/loading/error 처리 불필요)
 * - 컴포넌트는 렌더링만 담당
 */

import React from 'react';
import type { WithChartDataProps } from '@pharos/core/panel-registry';

export interface ChartPanelOptions {
  /** 차트 제목 */
  title?: string;
  /** 차트 색상 */
  color?: string;
  /** 범례 표시 여부 */
  showLegend?: boolean;
  /** 데이터 포인트 표시 */
  showDataPoints?: boolean;
}

/**
 * 차트 패널 컴포넌트 (순수 렌더링)
 *
 * 등록부에서 withChartData(ChartPanel)로 감싸 사용.
 */
export function ChartPanel({ chartData, options, title, dataProvider }: WithChartDataProps<ChartPanelOptions>) {
  const {
    color = '#3b82f6',
    showLegend = true,
  } = options || {};

  const hasData = chartData?.chartMetric && chartData.chartMetric.length > 0;
  const dataPoints = chartData?.chartMetric?.length || 0;
  const uniqueMetrics = chartData?.uniqueKeys || [];

  return (
    <div
      style={{
        height: '100%',
        padding: '20px',
        display: 'flex',
        flexDirection: 'column',
        gap: '16px',
      }}
    >
      {/* 헤더 */}
      <div style={{ borderBottom: `2px solid ${color}`, paddingBottom: '12px' }}>
        <h3 style={{ fontSize: '18px', fontWeight: '600', color, margin: 0 }}>
          {title}
        </h3>
      </div>

      {/* 컨텐츠 영역 */}
      <div
        style={{
          flex: 1,
          backgroundColor: '#f9fafb',
          borderRadius: '8px',
          padding: '24px',
          border: `2px solid ${color}33`,
          display: 'flex',
          flexDirection: 'column',
          gap: '16px',
        }}
      >
        {/* 데이터 상태 표시 */}
        <div style={{ textAlign: 'center' }}>
          <div
            style={{
              width: '80px',
              height: '80px',
              margin: '0 auto 16px',
              borderRadius: '50%',
              backgroundColor: `${color}20`,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: '40px',
            }}
          >
            {hasData ? '📈' : '📊'}
          </div>
          <h4 style={{ fontSize: '18px', fontWeight: '600', color, margin: 0 }}>
            {hasData ? '데이터 로드 완료' : '데이터 없음'}
          </h4>
        </div>

        {/* 데이터 통계 */}
        {hasData && (
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))',
              gap: '12px',
              marginTop: '16px',
            }}
          >
            <StatCard label="데이터 포인트" value={dataPoints} color={color} />
            <StatCard label="메트릭 개수" value={uniqueMetrics.length} color={color} />
            <StatCard label="쿼리 개수" value={dataProvider.chartQuery.length} color={color} />
          </div>
        )}

        {/* 범례 (옵션) */}
        {hasData && showLegend && uniqueMetrics.length > 0 && (
          <div
            style={{
              marginTop: '16px',
              padding: '12px',
              backgroundColor: 'white',
              borderRadius: '6px',
              border: '1px solid #e5e7eb',
            }}
          >
            <div style={{ fontSize: '12px', fontWeight: '600', color: '#6b7280', marginBottom: '8px' }}>
              메트릭 목록
            </div>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
              {uniqueMetrics.slice(0, 10).map((metric: string, index: number) => (
                <span
                  key={index}
                  style={{
                    fontSize: '11px',
                    padding: '4px 8px',
                    backgroundColor: `${color}15`,
                    color,
                    borderRadius: '4px',
                    border: `1px solid ${color}30`,
                  }}
                >
                  {metric}
                </span>
              ))}
              {uniqueMetrics.length > 10 && (
                <span style={{ fontSize: '11px', padding: '4px 8px', color: '#6b7280' }}>
                  +{uniqueMetrics.length - 10} more
                </span>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function StatCard({ label, value, color }: { label: string; value: number; color: string }) {
  return (
    <div
      style={{
        padding: '12px',
        backgroundColor: 'white',
        borderRadius: '6px',
        border: `1px solid ${color}30`,
        textAlign: 'center',
      }}
    >
      <div style={{ fontSize: '10px', color: '#6b7280', marginBottom: '4px' }}>{label}</div>
      <div style={{ fontSize: '24px', fontWeight: '700', color }}>{value.toLocaleString()}</div>
    </div>
  );
}
