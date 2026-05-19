import timeDuration from '@lib/timeDuration';
import {CellContext} from '@tanstack/react-table';

type time = string | number | Date;

const CellAge = ({getValue}: CellContext<any, unknown>) => {
  const date = getValue() as time;
  const age = timeDuration(new Date(date), new Date());
  return <div>{age}</div>;
};

export default CellAge;
