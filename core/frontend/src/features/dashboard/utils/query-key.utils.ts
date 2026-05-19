/**
 * dependsOn 객체를 안정적인 문자열로 변환
 * 
 * 문제점:
 * - JSON.stringify({b: 1, a: 2}) !== JSON.stringify({a: 2, b: 1})
 * - 동일한 값이지만 다른 queryKey 생성 → 불필요한 재요청
 * 
 * 해결:
 * - 키를 정렬하여 항상 동일한 문자열 생성
 * 
 * @param dependsOn - 의존성 객체
 * @returns 안정적인 JSON 문자열 또는 undefined
 */
export const serializeDependsOn = (dependsOn?: Record<string, string>): string | undefined => {
  if (!dependsOn || Object.keys(dependsOn).length === 0) {
    return undefined;
  }
  
  // 키를 정렬하여 일관된 문자열 생성
  const sortedKeys = Object.keys(dependsOn).sort();
  const sortedObj = sortedKeys.reduce((acc, key) => {
    acc[key] = dependsOn[key];
    return acc;
  }, {} as Record<string, string>);
  
  return JSON.stringify(sortedObj);
};
