import React from 'react';
import {Card} from '@pharos/shared/components/ui';
import {cn} from '@lib/utils';

const ChartCard = React.forwardRef<
  React.ComponentRef<typeof Card>,
  React.ComponentPropsWithoutRef<typeof Card>
>(({className, ...props}, ref) => (
  <Card
    ref={ref}
    className={cn('h-full gap-0 py-2 pb-0 rounded-none shadow-none border-none', className)}
    {...props}
  />
));
ChartCard.displayName = 'ChartCard';

export {ChartCard};
