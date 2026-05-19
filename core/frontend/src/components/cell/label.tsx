import {BadgeCustom} from '@pharos/shared/components/ui-extension';
import {CellContext} from '@tanstack/react-table';

type label = string[] | {[key: string]: any};

const CellLabel = ({getValue, column}: CellContext<any, unknown>) => {
  const value = getValue() as label;
  let values: string[] = [];
  if (Array.isArray(value)) {
    values = [...value];
  } else if (typeof value === 'object') {
    values = Object.entries(value).map(([key, value]) => `${key}=${value}`);
  }
  const displayValues = values.slice(0, 10);
  const isTruncated = value.length > 10;

  return (
    <div className={'flex flex-wrap gap-1'}>
      {displayValues?.map((displayValue, index) => (
        <BadgeCustom variant={'secondary'} key={`label-${column.id}-${index}`}>
          {displayValue}
        </BadgeCustom>
      ))}
      {isTruncated && <span>...</span>}
    </div>
  );
};

export default CellLabel;
