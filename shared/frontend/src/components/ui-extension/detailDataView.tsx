import React from 'react';
import {cn} from '../../lib';

interface DataNameProps extends React.HTMLAttributes<HTMLElement> {}

const DataName = ({className, children}: DataNameProps) => {
  return (
    <span className={cn('text-muted-foreground text-xs leading-4', className)}>{children}</span>
  );
};

const DataValue = ({className, children}: DataNameProps) => {
  return (
    <span className={cn(`sm:text-sm col-span-2 break-all text-balance`, className)}>
      {children}
    </span>
  );
};

interface DataViewProps extends React.HTMLAttributes<HTMLElement> {
  name: string;
}

const DataView = ({name, className, children}: DataViewProps) => {
  return (
    <div className="grid grid-rows-1 sm:grid-cols-3 gap-2 border-b py-2">
      <DataName>{name}</DataName>
      <DataValue className={cn(className)}>{children}</DataValue>
    </div>
  );
};

export {DataView};
