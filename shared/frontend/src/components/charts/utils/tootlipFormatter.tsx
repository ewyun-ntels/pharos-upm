import {Square} from '@pharos/shared/components';
import {NameType, Payload, ValueType} from 'recharts/types/component/DefaultTooltipContent';

type formatterProps = {
  value: ValueType;
  name: NameType;
  payload: Payload<ValueType, NameType>;
  color?: string;
  unitFn?: (value: any) => string;
};

const PromFormatter = ({value, name, payload, color, unitFn}: formatterProps) => {
  const colorVal = color ? color : payload.color;
  return (
    <div className={'w-full flex justify-between z-50'}>
      <div style={{color: colorVal}} className={'flex'}>
        <Square fill={colorVal} className={'h-4 w-4 p-1'} />
        {name}
      </div>
      <div className={'pl-1'}>{unitFn ? unitFn(value) : value.toLocaleString()}</div>
    </div>
  );
};

export {PromFormatter};
