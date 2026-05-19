/**
 * Variable Parser — Pongo2 Template Variable Extraction
 * 
 * 백엔드 Pongo2와 100% 호환되는 정규식 기반 경량 파서.
 * variable-system과 panels 모두에서 import해서 사용하는 공유 유틸리티.
 */

const PATTERNS = {
  basic:    /\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*}}/g,
  field:    /\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\.[a-zA-Z0-9_.]+\s*}}/g,
  filter:   /\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\|[^}]+}}/g,
  rawBlock: /\{%\s*raw\s*%}[\s\S]*?\{%\s*endraw\s*%}/g,
  escaped:  /\\\{\\\{[^}]*\\}\\}/g,
};

import {
  QUERY_PARAM_START_TIME,
  QUERY_PARAM_END_TIME,
  QUERY_PARAM_STEP,
  QUERY_PARAM_REFRESH_COUNT,
} from '@lib/query-params';

// 시스템 내장 variable (필터가 아니므로 의존성 목록에서 제외)
const BUILTIN_VARIABLES: Set<string> = new Set([
  QUERY_PARAM_START_TIME,
  QUERY_PARAM_END_TIME,
  QUERY_PARAM_STEP,
  QUERY_PARAM_REFRESH_COUNT,
]);

/**
 * 쿼리 문자열에서 {{ variable }} 참조를 추출합니다.
 * raw 블록, 이스케이프, 내장 variable은 자동 제외됩니다.
 * 
 * @example
 * extractVariables("SELECT * FROM {{ table }} WHERE host = {{ host }}")
 * // => ["host", "table"]
 * 
 * @example
 * extractVariables("{% raw %}{{ not_a_var }}{% endraw %} {{ real_var }}")
 * // => ["real_var"]
 */
export function extractVariables(query: string): string[] {
  if (!query) return [];

  // raw 블록과 이스케이프된 부분 제거
  let clean = query
    .replace(PATTERNS.rawBlock, '')
    .replace(PATTERNS.escaped, '');

  const found = new Set<string>();
  
  // 세 가지 패턴 모두 매칭
  for (const pattern of [PATTERNS.basic, PATTERNS.field, PATTERNS.filter]) {
    for (const match of clean.matchAll(pattern)) {
      const varName = match[1];
      if (!BUILTIN_VARIABLES.has(varName)) {
        found.add(varName);
      }
    }
  }

  return [...found].sort();
}

/**
 * filterMetas의 {{variable}}을 실제 값으로 치환합니다. (UI 표시용)
 * Query 치환은 Backend Pongo2가 처리하므로 이 함수는 title/description 등에만 사용합니다.
 *
 * @example
 * replaceVariables("CPU - {{instance}}", filterMetas)
 * // → "CPU - node1"
 */
export function replaceVariables(
  template: string | undefined,
  filterMetas: Map<string, { value?: unknown }> | undefined,
): string {
  if (!template) return '';
  if (!filterMetas?.size) return template;

  return template.replace(
    PATTERNS.basic,
    (match, varName) => {
      const meta = filterMetas.get(varName);
      if (!meta?.value) return match;

      return Array.isArray(meta.value)
        ? meta.value.join(', ')
        : String(meta.value);
    },
  );
}

/**
 * 쿼리 문법 검증 (템플릿 브레이스 매칭 등)
 */
export function validateSyntax(query: string): { isValid: boolean; errors: string[] } {
  const errors: string[] = [];
  
  const open  = (query.match(/\{\{/g) || []).length;
  const close = (query.match(/}}/g) || []).length;
  
  if (open !== close) {
    errors.push('Unmatched template braces {{ }}');
  }
  
  // ${ } 사용 감지 (Pongo2는 {{ }} 사용)
  if (/\$\{[^}]*}/g.test(query)) {
    errors.push('Use {{ }} instead of ${}');
  }
  
  return { 
    isValid: errors.length === 0, 
    errors 
  };
}
