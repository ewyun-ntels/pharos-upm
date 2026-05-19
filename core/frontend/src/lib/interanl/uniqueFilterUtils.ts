import {Table} from '@tanstack/react-table';
import {FacetedFilterOption} from '@pharos/shared/components/ui-extension';

const getUniqueFilter = (table: Table<any>, columnName: string) => {
  const iterableIterator = table.getColumn(columnName)?.getFacetedUniqueValues().keys();
  let options: FacetedFilterOption[] = [];
  if (iterableIterator) {
    Array.from(iterableIterator).forEach((value) => {
      options.push({label: value, value: value});
    });
  }
  return options;
};

export {getUniqueFilter};
