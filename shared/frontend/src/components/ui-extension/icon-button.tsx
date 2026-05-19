import {Button} from '@pharos/shared/components/ui';
import * as React from 'react';
import {cn} from '../../lib';
import {Tooltip, TooltipContent, TooltipProvider, TooltipTrigger} from './tooltip-custom';

type IconButtonSize = 'icon' | 'icon-xs';
const iconSizeClasses: Record<IconButtonSize, string> = {
  icon: 'w-8 h-8',
  'icon-xs': 'w-6 h-6 p-1 [&_svg]:!size-3.5',
};
interface IconButtonProps extends Omit<React.ComponentPropsWithoutRef<typeof Button>, 'size'> {
  icon: React.ReactNode;
  size?: IconButtonSize;
}

export const IconButton = React.forwardRef<HTMLButtonElement, IconButtonProps>(
  ({className, variant, icon, size = 'icon', asChild = false, children, ...props}, ref) => {
    const iconSizeClass = iconSizeClasses[size];

    return (
      // 아이콘 버튼 툴팁 효과 상위 버튼에 group 클래스명 삽입, 해당 tooltip에 group-hover:visible 삽입
      <Button
        ref={ref}
        className={cn('relative group gap-1 cursor-pointer', iconSizeClass, className)}
        variant={variant}
        size={undefined}
        asChild={asChild}
        {...props}
      >
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <span>{icon}</span>
            </TooltipTrigger>
            <TooltipContent variant="icon" side="bottom" className="mt-0.5 font-normal">
              {children}
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </Button>
    );
  },
);
IconButton.displayName = 'IconButton';
