/**
 * 타입 정의
 */

export interface ChartQueryArgs {
  [key: string]: any;
}

export interface ChartQuery {
  query: string;
  label?: string;
  datasourceName?: string;
}

export interface MyAwesomePanelOptions {
  // 색상 설정
  primaryColor?: string;
  secondaryColor?: string;
  
  // 크기 설정
  size?: 'small' | 'medium' | 'large';
  
  // 표시 옵션
  showLabels?: boolean;
  showLegend?: boolean;
  
  // 애니메이션
  animated?: boolean;
  
  // 커스텀 메시지
  title?: string;
  description?: string;
}

export interface MyAwesomePanelProps {
  // 데이터 관련
  args?: ChartQueryArgs;
  query?: ChartQuery;
  resource?: string;
  dataProviderName?: string;
  
  // 플러그인 옵션 (PanelProps와 일관성)
  options?: MyAwesomePanelOptions;
  
  // Dashboard 정보
  id?: string;
  dashboardId?: string;
  location?: string;
  
  // 콜백
  timeRangeCallback?: (start: any, end: any) => void;
  
  // 기타
  refetchInterval?: number | false;
  [key: string]: any;
}

export interface ChartEditorParams {
  type: string;
  title?: string;
  description?: string;
  bgTransparent?: boolean;
  dataProvider?: {
    chartQuery?: ChartQuery;
    dataProviderName?: string;
    resource?: string;
  };
  dateTimeRange?: {
    startTime?: any;
    endTime?: any;
  };
  refreshCount?: number;
  
  // My Awesome Panel 전용 옵션
  myAwesomeOptions?: MyAwesomePanelOptions;
}
