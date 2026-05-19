import {ColumnConfig} from '@pharos/shared/hooks/table-columns';
import {DASHBOARD_PROVIDER_NAME, DASHBOARD_RESOURCES} from '@providers/dashboard-provider';
import type { Panel } from '@pharos/shared/types/dashboard';
import type { PanelProps } from '@pharos/core/panel-registry';
import type { TablePanelOptions } from './types';

export const extractColumnsFromQuery = (query: string): string[] => {
  // 템플릿 변수와 쿼리 정규화
  const normalizedQuery = query
    .replace(/\{\{[^}]+}}/g, '1') // 템플릿 변수를 숫자로 치환
    .replace(/\s+/g, ' ') // 연속 공백을 하나로
    .replace(/,\s*/g, ', ') // 쉼표 뒤 공백 정규화
    .trim();

  const upperQuery = normalizedQuery.toUpperCase();

  // CTE가 있는 경우와 없는 경우 구분
  let selectIndex = -1;

  // WITH절이 있는 경우 마지막 SELECT 찾기
  if (upperQuery.includes('WITH ')) {
    let currentIndex = upperQuery.indexOf('WITH ');
    let depth = 0;
    let inString = false;
    let stringChar = '';

    // WITH절 이후의 마지막 SELECT 찾기
    while (currentIndex < upperQuery.length - 6) {
      const char = upperQuery[currentIndex];

      if ((char === "'" || char === '"' || char === '`') && !inString) {
        inString = true;
        stringChar = char;
      } else if (char === stringChar && inString) {
        inString = false;
        stringChar = '';
      }

      if (!inString) {
        if (char === '(') depth++;
        else if (char === ')') depth--;

        // 괄호 밖에서 SELECT 발견
        if (depth === 0 && upperQuery.slice(currentIndex, currentIndex + 7) === 'SELECT ') {
          selectIndex = currentIndex;
        }
      }
      currentIndex++;
    }
  } else {
    // 일반 쿼리에서 괄호 밖의 첫 번째 SELECT 찾기
    let currentIndex = 0;
    let depth = 0;
    let inString = false;
    let stringChar = '';

    while (currentIndex < upperQuery.length - 6) {
      const char = upperQuery[currentIndex];

      if ((char === "'" || char === '"' || char === '`') && !inString) {
        inString = true;
        stringChar = char;
      } else if (char === stringChar && inString) {
        inString = false;
        stringChar = '';
      }

      if (!inString) {
        if (char === '(') depth++;
        else if (char === ')') depth--;

        // 괄호 밖에서 SELECT 발견 (첫 번째만)
        if (depth === 0 && upperQuery.slice(currentIndex, currentIndex + 7) === 'SELECT ' && selectIndex === -1) {
          selectIndex = currentIndex;
          break;
        }
      }
      currentIndex++;
    }
  }

  if (selectIndex === -1) return [];

  // FROM 구간 찾기 (괄호와 문자열 깊이 고려)
  let fromIndex = -1;
  let depth = 0;
  let inString = false;
  let stringChar = '';
  let currentIndex = selectIndex + 7;

  while (currentIndex < upperQuery.length) {
    const char = upperQuery[currentIndex];

    if ((char === "'" || char === '"' || char === '`') && !inString) {
      inString = true;
      stringChar = char;
    } else if (char === stringChar && inString) {
      inString = false;
      stringChar = '';
    }

    if (!inString) {
      if (char === '(') depth++;
      else if (char === ')') depth--;

      if (depth === 0 && upperQuery.slice(currentIndex, currentIndex + 5) === 'FROM ') {
        fromIndex = currentIndex;
        break;
      }
    }
    currentIndex++;
  }

  if (fromIndex === -1) return [];

  // SELECT~FROM 사이 컬럼 정의 추출
  const columnDefinitions = normalizedQuery.substring(selectIndex + 6, fromIndex).trim();

  // 서브쿼리를 고려한 컬럼 분리
  const columns: string[] = [];
  let buf = '';
  depth = 0;
  inString = false;
  stringChar = '';

  for (let i = 0; i < columnDefinitions.length; i++) {
    const char = columnDefinitions[i];

    if ((char === "'" || char === '"' || char === '`') && !inString) {
      inString = true;
      stringChar = char;
      buf += char;
    } else if (char === stringChar && inString) {
      inString = false;
      stringChar = '';
      buf += char;
    } else if (inString) {
      buf += char;
    } else {
      if (char === '(') depth++;
      if (char === ')') depth--;

      if (char === ',' && depth === 0) {
        if (buf.trim()) {
          columns.push(buf.trim());
        }
        buf = '';
      } else {
        buf += char;
      }
    }
  }

  if (buf.trim()) {
    columns.push(buf.trim());
  }

  // 개선된 별칭 추출
  return columns
    .map((col) => {
      const trimmedCol = col.trim();
      if (!trimmedCol) return '';

      // 1. AS 별칭 추출 (따옴표 포함)
      const asQuoteMatch = trimmedCol.match(/\s+as\s+(['"`])(.*?)\1/i);
      if (asQuoteMatch) return asQuoteMatch[2].trim();

      // 2. AS 별칭 추출 (따옴표 없음)
      const asMatch = trimmedCol.match(/\s+as\s+([^\s,'"`()]+)/i);
      if (asMatch) return asMatch[1].trim();

      // 3. 끝에 따옴표로 감싸진 별칭
      const endQuoteMatch = trimmedCol.match(/(['"`])(.*?)\1\s*$/);
      if (endQuoteMatch) return endQuoteMatch[2].trim();

      // 4. 서브쿼리 처리: (SELECT ... FROM ...) AS 별칭
      if (trimmedCol.startsWith('(') && trimmedCol.includes('SELECT')) {
        const subQueryAliasMatch = trimmedCol.match(/\)\s+as\s+(['"`])(.*?)\1/i);
        if (subQueryAliasMatch) return subQueryAliasMatch[2].trim();

        const subQueryAliasNoQuote = trimmedCol.match(/\)\s+as\s+([^\s,'"`()]+)/i);
        if (subQueryAliasNoQuote) return subQueryAliasNoQuote[1].trim();

        // AS 없이 서브쿼리 끝에 별칭
        const subQueryEndAlias = trimmedCol.match(/\)\s+(['"`])(.*?)\1\s*$/);
        if (subQueryEndAlias) return subQueryEndAlias[2].trim();
      }

      // 5. 함수 호출 + 별칭
      const funcAliasMatch = trimmedCol.match(/\)\s+(['"`])(.*?)\1\s*$/);
      if (funcAliasMatch) return funcAliasMatch[2].trim();

      // 6. 테이블.컬럼 형태에서 컬럼명만 추출
      const tableColumnMatch = trimmedCol.match(/(\w+)\.(\w+)(?:\s+|$)/);
      if (tableColumnMatch) return tableColumnMatch[2];

      // 7. 단순 컬럼명
      const simpleColumnMatch = trimmedCol.match(/^(\w+)(?:\s+|$)/);
      if (simpleColumnMatch) return simpleColumnMatch[1];

      // 8. 최후 수단: 특수문자 제거 후 마지막 단어
      return (
        trimmedCol
          .replace(/[`'"()]/g, '')
          .split(/[.\s]/)
          .filter(Boolean)
          .pop() || ''
      );
    })
    .filter(Boolean);
};

export const setParamToChart = async (
  panel: Panel,
): Promise<Partial<PanelProps<TablePanelOptions>>> => {
  // Table 패널은 chartQuery가 정확히 1개만 허용됨 (있는 경우)
  const chartQueries = panel.dataProvider?.chartQuery || [];

  if (chartQueries.length > 1) {
    throw new Error('Table panel only supports a single query. Please use only one chartQuery.');
  }
  // chartQuery가 없어도 허용 (새 패널 생성 시)

  const query = panel.dataProvider?.chartQuery?.[0]?.query || '';
  const datasourceName = panel.dataProvider?.chartQuery?.[0]?.datasourceName || '';

  let columnConfigs: ColumnConfig<any>[] = [];

  if (query) {
    try {
      const extractedColumns = extractColumnsFromQuery(query);
      columnConfigs = extractedColumns.map((col) => ({key: col}));
    } catch (e) {
      console.error('Failed to extract columns from query:', e);
    }
  }

  // extractColumnsFromQuery 실패 시 기존 설정 사용
  if (!columnConfigs || columnConfigs.length === 0) {
    columnConfigs = panel.options?.columnConfigs || [];
  }

  const config = new Map<string, ColumnConfig>(
    panel.options?.columnConfigs?.map((item: ColumnConfig) => [item.key as string, item]),
  );

  columnConfigs = columnConfigs.map((item) => {
    const configItem = config.get(item.key as string);
    if (configItem?.properties) {
      // 저장된 properties 전체(size, badge, colorRules 등)를 유지
      return { ...item, properties: configItem.properties };
    }
    return item;
  });

  // 저장된 columnConfigs에는 있지만 쿼리 파싱 결과에 없는 항목도 보존
  // (쿼리가 바뀌어도 수동으로 추가한 컬럼 설정 유지)
  const parsedKeys = new Set(columnConfigs.map((c) => String(c.key)));
  const extraConfigs = (panel.options?.columnConfigs || []).filter(
    (c: ColumnConfig) => !parsedKeys.has(String(c.key)),
  );
  columnConfigs = [...columnConfigs, ...extraConfigs];

  return {
    dataProvider: {
      chartQuery: [
        {
          query: query,
          datasourceName: datasourceName,
          label: '',
        },
      ],
      dataProviderName: DASHBOARD_PROVIDER_NAME,
      resource: DASHBOARD_RESOURCES.RAW,
    },
    options: {
      columnConfigs: columnConfigs,
      usePagenation: panel.options?.usePagenation || false,
      leftItems: panel.options?.leftItems || [],
      rightItems: panel.options?.rightItems || [],
      checkboxConfig: panel.options?.checkboxConfig,
      columnFilters: panel.options?.columnFilters,
      properties: panel.options?.properties || {},
      columnDataLinks: panel.options?.columnDataLinks || {},
    },
  };
};

const checkSearchItem = (items: {id: string; type: string}[], typeName: string) => {
  return items.some(({id, type}) => id === typeName && type === typeName);
};

export const setDefaultParam = (panelData: PanelProps<TablePanelOptions>) => {
  const optionsData = panelData.options;

  const useSearch = checkSearchItem(optionsData?.leftItems || [], 'search');
  const useExport = checkSearchItem(optionsData?.rightItems || [], 'export');

  return {
    dataProvider: {
      chartQuery: [
        {
          query: panelData?.dataProvider?.chartQuery?.[0]?.query || '',
          datasourceName: panelData?.dataProvider?.chartQuery?.[0]?.datasourceName || '',
          label: panelData?.dataProvider?.chartQuery?.[0]?.label || '',
        },
      ],
      dataProviderName: panelData?.dataProvider?.dataProviderName || '',
      resource: 'posts',
    },
    options: {...optionsData, useSearch: useSearch, useExport: useExport},
  };
};
