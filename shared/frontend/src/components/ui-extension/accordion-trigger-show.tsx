import React from 'react';
import {AccordionTrigger} from '@pharos/shared/components/ui';

export const AccordionTriggerShow = ({children}: {children: React.ReactNode}) => {
  return (
    <AccordionTrigger className="grid grid-cols-[minmax(60px,1fr)_60px_30px] py-0 pb-4 [&_svg>*]:!fill-violet-600">
      <div className="truncate text-left">{children}</div>
      <span className="text-violet-600 ml-auto mr-2 text-xs">Show</span>
    </AccordionTrigger>
  );
};
