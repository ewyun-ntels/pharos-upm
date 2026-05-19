import '@tanstack/react-table';

declare module '@tanstack/react-table' {
  interface ColumnMeta<TData, TValue> {
    /** flex 비율 (AG Grid 스타일). 지정 시 컨테이너 여백을 비율에 따라 배분 */
    flex?: number;
    cellClassName?: string;
    rowSpan?: number;
    searchByFormatted?: boolean;
  }
}
