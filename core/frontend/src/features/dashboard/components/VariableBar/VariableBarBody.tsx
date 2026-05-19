'use client';

import {VariableBarBodyProps} from './types';

export const VariableBarBody = ({children}: VariableBarBodyProps) => {
  return <div className="flex flex-col flex-1 min-w-0 px-4 pb-4">{children}</div>;
};
