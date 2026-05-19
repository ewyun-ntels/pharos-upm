import React from 'react';

/**
 * React 컴포넌트나 복잡한 데이터 구조에서 검색 가능한 텍스트를 추출하는 재귀 함수
 */
export const extractSearchableText = (content: any, fallbackValue?: any): string => {
  if (content == null) return '';

  if (React.isValidElement(content)) {
    const props = content.props as any;

    if (props.children !== undefined && props.children !== null) {
      if (Array.isArray(props.children)) {
        return props.children.map((child: any) => extractSearchableText(child)).join(' ');
      } else {
        return extractSearchableText(props.children);
      }
    }

    return fallbackValue ? String(fallbackValue) : '';
  }

  if (Array.isArray(content)) {
    return content.map((item: any) => extractSearchableText(item)).join(' ');
  }

  return String(content);
};
