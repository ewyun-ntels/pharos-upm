import {ChartEditorParams, MyAwesomePanelOptions} from './types';

/**
 * setParamToChart
 * 
 * Panel Editor의 Form 데이터를 Panel 컴포넌트에 전달할 데이터로 변환합니다.
 * 
 * @param chart - Panel Editor의 Form 데이터
 * @param extraArgs - 추가 인자 (선택사항)
 * @returns Panel 컴포넌트에 전달할 props
 */
export const setParamToChart = async (
  chart: ChartEditorParams,
  extraArgs?: any,
): Promise<Partial<ChartEditorParams>> => {
  if (!chart) {
    return {};
  }

  // 기본 옵션
  const defaultOptions: MyAwesomePanelOptions = {
    primaryColor: '#3b82f6',
    secondaryColor: '#8b5cf6',
    size: 'medium',
    showLabels: true,
    showLegend: true,
    animated: true,
    title: 'My Awesome Panel',
    description: 'This is an example external plugin',
  };

  // 사용자 정의 옵션과 병합
  const myAwesomeOptions: MyAwesomePanelOptions = {
    ...defaultOptions,
    ...chart.myAwesomeOptions,
  };

  // DataProvider 정보 추출
  const dataProvider = {
    chartQuery: {
      query: chart.dataProvider?.chartQuery?.query || '',
      label: chart.dataProvider?.chartQuery?.label || '',
      datasourceName: chart.dataProvider?.chartQuery?.datasourceName || 'prometheus',
    },
    dataProviderName: chart.dataProvider?.dataProviderName || 'dashboardProvider',
    resource: chart.dataProvider?.resource || 'rechart',
  };

  // Panel 컴포넌트에 전달할 최종 데이터
  return {
    // 기본 정보
    title: chart.title,
    description: chart.description,
    bgTransparent: chart.bgTransparent,
    
    // DataProvider
    dataProvider,
    
    // Query 정보 (새로운 구조)
    query: dataProvider.chartQuery,
    resource: dataProvider.resource,
    dataProviderName: dataProvider.dataProviderName,
    
    // 커스텀 옵션
    chartOptions: myAwesomeOptions,
    myAwesomeOptions, // Form에서 사용
    
    // 추가 인자가 있으면 포함
    ...(extraArgs && {extraArgs}),
  };
};

/**
 * setDefaultParam
 * 
 * 새 패널 생성 시 기본값을 설정합니다.
 * 
 * @param data - 초기 데이터 (선택사항)
 * @returns 기본 설정이 적용된 데이터
 */
export const setDefaultParam = (data?: any) => {
  const defaults: Partial<ChartEditorParams> = {
    type: 'my-awesome-panel',
    title: 'My Awesome Panel',
    description: 'External plugin example',
    bgTransparent: false,
    
    dataProvider: {
      chartQuery: {
        query: 'up{job="prometheus"}',
        label: '',
        datasourceName: 'prometheus',
      },
      dataProviderName: 'dashboardProvider',
      resource: 'rechart',
    },
    
    myAwesomeOptions: {
      primaryColor: '#3b82f6',
      secondaryColor: '#8b5cf6',
      size: 'medium',
      showLabels: true,
      showLegend: true,
      animated: true,
      title: 'My Awesome Panel',
      description: 'This is an example external plugin',
    },
  };

  return {
    ...defaults,
    ...data,
  };
};

/**
 * toFormData (선택사항)
 * 
 * 저장된 Panel 데이터를 Form 데이터로 역변환합니다.
 * Panel 편집 시 사용됩니다.
 * 
 * @param panelData - 저장된 Panel 데이터
 * @returns Form에서 사용할 데이터
 */
export const toFormData = (panelData: any) => {
  if (!panelData) {
    return setDefaultParam();
  }

  return {
    type: panelData.type || 'my-awesome-panel',
    title: panelData.title,
    description: panelData.description,
    bgTransparent: panelData.bgTransparent,
    
    dataProvider: panelData.dataProvider || {
      chartQuery: {
        query: panelData.query?.query || '',
        label: panelData.query?.label || '',
        datasourceName: panelData.query?.datasourceName || 'prometheus',
      },
      dataProviderName: panelData.dataProviderName || 'dashboardProvider',
      resource: panelData.resource || 'rechart',
    },
    
    myAwesomeOptions: panelData.chartOptions || panelData.myAwesomeOptions || setDefaultParam().myAwesomeOptions,
  };
};
