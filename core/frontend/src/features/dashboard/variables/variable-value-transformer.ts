/**
 * VariableValueTransformer
 * 
 * 변수 값을 쿼리에서 사용할 수 있는 형식으로 변환
 * - 배열 → 'val1','val2' 
 * - 특수 함수: abc:quota → quota(abc)
 * - 이스케이프 처리
 */

/**
 * SQL 단일쿼트로 감싸진 값을 제거 ('.*' → .*)
 * regex 포맷에서 customAllValue 등이 SQL 포맷으로 저장된 경우 처리
 * ewyun-20260515
 */
function stripSqlQuotes(s: string): string {
  if (s.length >= 2 && s.startsWith("'") && s.endsWith("'")) {
    return s.slice(1, -1);
  }
  return s;
}

/**
 * 배열을 datasource 포맷에 맞게 변환
 * - sql (기본): 'val1','val2'  → SQL IN 절용
 * - regex:      val1|val2     → Prometheus 정규식 OR용
 * ewyun-20260515: Prometheus 변수값이 'pod' 처럼 싱글쿼트로 감싸져 regex 매칭 실패하던 문제 수정
 */
function transformArrayValue(values: unknown[], format: 'sql' | 'regex' = 'sql'): string {
  if (format === 'regex') {
    return values.map(v => stripSqlQuotes(String(v))).join('|');
  }
  return values.map(v => `'${String(v)}'`).join(',');
}

/**
 * 특수 함수 처리 (미래 확장용)
 * 예: "myvalue:quota" → "quota(myvalue)"
 */
function applySpecialFunctions(value: string, functions?: string[]): string {
  if (!functions || functions.length === 0) return value;
  
  // 예: ['quota', 'upper'] → upper(quota(value))
  return functions.reduceRight((acc, fn) => `${fn}(${acc})`, value);
}

/**
 * 변수 값을 쿼리 args로 변환
 * 
 * @param value - 변수 값 (string | string[] | number)
 * @param options - 변환 옵션
 * @returns 변환된 값
 */
export function transformVariableValue(
  value: unknown,
  options?: {
    /** 특수 함수 목록 (예: ['quota', 'upper']) */
    functions?: string[];
    /** 이스케이프 여부 */
    escape?: boolean;
    /** 배열 직렬화 포맷: 'sql'(기본) → 'val1','val2' / 'regex' → val1|val2 */
    format?: 'sql' | 'regex';
  }
): string | number {
  // null/undefined
  if (value === null || value === undefined) {
    return '';
  }

  // 배열 → datasource 포맷에 맞게 변환
  if (Array.isArray(value)) {
    const transformed = transformArrayValue(value, options?.format);
    return options?.functions
      ? applySpecialFunctions(transformed, options.functions)
      : transformed;
  }
  
  // 숫자 → 그대로
  if (typeof value === 'number') {
    return value;
  }
  
  // 문자열 → regex 포맷이면 SQL 쿼트 제거 후 반환
  const strValue = options?.format === 'regex'
    ? stripSqlQuotes(String(value))
    : String(value);
  return options?.functions
    ? applySpecialFunctions(strValue, options.functions)
    : strValue;
}

/**
 * VariableMeta Map을 변환된 args Map으로 변환
 */
export function transformVariableMetasToArgs(
  filterMetas: Map<string, { value?: unknown }>,
  varNames: string[]
): Map<string, string | number> {
  const args = new Map<string, string | number>();
  
  varNames.forEach(varName => {
    const meta = filterMetas.get(varName);
    if (meta?.value !== undefined) {
      const transformed = transformVariableValue(meta.value);
      args.set(varName, transformed);
    }
  });
  
  return args;
}
