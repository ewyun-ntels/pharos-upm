import React from 'react';
import {render, screen} from '@testing-library/react';
import {PieCardChart as PromPieChart} from './pieCard';
import type {Action} from '@pharos/shared/types/dashboard';

/**
 * PromPieChart 핵심 기능 테스트
 *
 * 목적: pluginContext 구조 검증 및 주요 에러 케이스 테스트
 * 상세 렌더링/계산 로직은 E2E 테스트에서 검증
 */

// Mock data
let mockIsLoading: boolean;
let mockError: any;
let mockChartData: any;

// Mock useDashboardQuery
const mockUseDashboardQuery = jest.fn(() => ({
  data: mockChartData,
  isLoading: mockIsLoading,
  isFetching: mockIsLoading,
  isSuccess: !mockIsLoading && !mockError,
  isError: !!mockError,
  error: mockError,
}));

const mockToChartQueryRequests = jest.fn((queries) => queries);

// Mock Refine hooks (pluginContext 구조 검증용)
const mockUseList = jest.fn();
const mockUseOne = jest.fn();
const mockUseCreate = jest.fn();
const mockUseUpdate = jest.fn();
const mockUseDelete = jest.fn();
const mockUseCustom = jest.fn();
const mockUseTranslation = jest.fn(() => ({
  translate: (_: string) => 'Chart data not exists.',
  changeLocale: jest.fn(),
  getLocale: jest.fn(() => 'en'),
}));

jest.mock('@/lib/data-provider', () => ({
  ...jest.requireActual('@/lib/data-provider'),
  useTranslation: () => ({translate: (_: string) => 'Chart data not exists.'}),
}));

jest.mock('@features/dashboard/hooks/use-dashboard-store', () => ({
  useDashboardStore: jest.fn((selector: any) =>
    selector({
      setFilter: jest.fn(),
      getFilter: jest.fn(() => undefined),
    }),
  ),
}));

// Minimal Recharts mock
jest.mock('recharts', () => ({
  __esModule: true,
  PieChart: ({children}: any) => <div data-testid="piechart">{children}</div>,
  Pie: () => <div data-testid="pie" />,
  Legend: () => <div data-testid="legend" />,
}));

jest.mock('@pharos/shared/components/ui/chart', () => ({
  __esModule: true,
  ChartContainer: ({children}: any) => <div>{children}</div>,
  ChartTooltip: () => null,
  ChartTooltipContent: () => null,
}));

jest.mock('@pharos/shared/components/charts/utils/tootlipFormatter', () => ({
  PromFormatter: () => null,
}));

jest.mock('@pharos/shared/components/charts/utils/legendFormatter', () => ({
  formatLegendLabel: (key: string) => key,
  PromLegend: () => null,
}));

jest.mock('@pharos/shared/components/charts/utils/tableLegendFormatter', () => ({
  PromLegendTable: () => null,
}));

jest.mock('@pharos/shared/components/ui-extension', () => ({
  LoadingIndicator: () => <div className="loading">Loading...</div>,
  PanelLoadingBar: () => <div className="panel-loading-bar" />,
}));

const baseProps = {
  permission: 'editor' as Action,
  refetchInterval: false as const,
  args: {} as any,
  dataProvider: {
    chartQuery: [
      {
        datasourceName: 'test',
        label: 'test',
        query: 'test query',
      },
    ],
    dataProviderName: 'test',
    resource: 'test',
  },
  options: {
    legendAlign: 'bottom' as const,
    legendEnabled: true,
    legendRule: 'default',
    showLast: false,
  },
  pluginContext: {
    hooks: {
      useDashboardQuery: mockUseDashboardQuery,
      toChartQueryRequests: mockToChartQueryRequests,
      useList: mockUseList,
      useOne: mockUseOne,
      useCreate: mockUseCreate,
      useUpdate: mockUseUpdate,
      useDelete: mockUseDelete,
      useCustom: mockUseCustom,
      useTranslation: mockUseTranslation,
    },
  },
};

describe('PromPieChart - Core Functionality', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockIsLoading = false;
    mockError = null;
    mockChartData = undefined;
  });

  it('shows error when pluginContext is missing', () => {
    const {container} = render(<PromPieChart {...baseProps} pluginContext={undefined as any} />);

    expect(container.textContent).toContain('useDashboardQuery not available in pluginContext');
  });

  it('shows loading state', () => {
    mockIsLoading = true;

    const {container} = render(<PromPieChart {...baseProps} />);

    expect(container.querySelector('.loading')).toBeInTheDocument();
  });

  it('shows no data message when chartMetric is empty', () => {
    mockChartData = {
      chartType: 'timeseries',
      uniqueKeys: [],
      chartMetric: [],
    };

    render(<PromPieChart {...baseProps} />);

    expect(screen.getByText('Chart data not exists.')).toBeInTheDocument();
  });

  it('uses panel mode for viewer even when saved chart query is empty', () => {
    mockChartData = {
      chartType: 'timeseries',
      uniqueKeys: [],
      chartMetric: [],
    };

    render(
      <PromPieChart
        {...baseProps}
        permission="viewer"
        dashboardId="dashboard-1"
        id="panel-1"
        kind="panels"
        dataProvider={{
          ...baseProps.dataProvider,
          chartQuery: [{datasourceName: '', label: '', query: ''}],
        }}
      />,
    );

    expect(mockToChartQueryRequests).not.toHaveBeenCalled();
    expect(mockUseDashboardQuery).toHaveBeenCalledWith(expect.objectContaining({
      dashboardId: 'dashboard-1',
      queries: [expect.objectContaining({
        dashboardId: 'dashboard-1',
        queryName: 'query-0',
        id: 'panel-1',
        kind: 'panels',
        query: '',
        datasourceName: '',
      })],
    }));
  });

  it('shows unsupported message for non-timeseries data', () => {
    mockChartData = {
      chartType: 'table',
      uniqueKeys: [],
      chartMetric: [{}],
    };

    const {container} = render(<PromPieChart {...baseProps} />);

    expect(container.textContent).toContain('Unsupported chart data($table)');
  });
});
