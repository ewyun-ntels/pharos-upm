import * as React from 'react';
import {cva, type VariantProps} from 'class-variance-authority';
import {cn} from '../../lib';

const badgeVariants = cva(
  'badge-custom inline-flex w-fit min-w-11 min-h-5 justify-center items-center rounded-md border px-2.5 py-0.5 text-xs [&>div]:text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2',
  {
    variants: {
      variant: {
        default: 'border-transparent bg-background text-foreground',
        green: 'border-transparent bg-green-600/20 text-green-600 text-nowrap',
        blue: 'border-transparent bg-blue-500/20 text-blue-500 text-nowrap',
        yellow: 'border-transparent bg-yellow-400/30 text-yellow-600 text-nowrap',
        orange: 'border-transparent bg-orange-500/20 text-orange-500 text-nowrap',
        red: 'border-transparent bg-red-600/20 text-red-600 text-nowrap',
        gray: 'border-transparent bg-gray-400/20 text-gray-400 text-nowrap',
        violet: 'border-transparent bg-violet-500/20 text-violet-500 text-nowrap',
        primary:
          'border-transparent bg-violet-500 dark:bg-primary text-primary-foreground text-nowrap',
        notprimary:
          'border-transparent bg-gray-400 dark:bg-gray-400/30 text-primary-foreground text-nowrap',
        secondary: 'border-transparent bg-secondary text-secondary-foreground',
        destructive: 'border-transparent bg-destructive text-destructive-foreground text-nowrap',
        outline: 'bg-transparent border border-border',
      },
      outline: {
        true: 'bg-transparent border border-border',
      },
      circle: {
        true: 'pl-2',
      },
      addColor: {
        color1:
          'bg-transparent border border-border text-blue-500 text-nowrap h-5 min-w-11 justify-center',
        color2:
          'bg-transparent border border-border text-cyan-600 text-nowrap h-5 min-w-11 justify-center',
        color3: 'bg-transparent border border-border text-gray-500',
        color4: 'bg-transparent border border-border text-violet-500',
        color5:
          'bg-transparent border border-border text-green-600 text-nowrap h-5 min-w-11 justify-center',
        color6:
          'bg-transparent border border-border text-slate-600 text-nowrap h-5 min-w-11 justify-center',
        // 추가 커스텀 컬러를 만드세요.
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  },
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {
  isTruncated?: boolean;
  outline?: boolean;
  circle?: boolean;
}

function BadgeCustom({
  className,
  variant,
  outline = false,
  circle = false,
  isTruncated,
  addColor,
  ...props
}: BadgeProps) {
  const ref = React.useRef<any>(null);

  const handleClick = () => {
    if (ref.current.classList.contains('line-clamp-1')) {
      ref.current.classList.remove('line-clamp-1');
    } else {
      ref.current.classList.add('line-clamp-1');
    }
  };

  const isLongText = typeof props.children === 'string' && props.children.length > 20;

  return (
    <div
      className={cn(
        badgeVariants({variant, outline, circle, addColor}),
        className,
        isTruncated && 'text-ellipsis line-clamp-1',
        isLongText && 'break-all whitespace-normal',
      )}
      {...props}
      ref={ref}
      onClick={handleClick}
    >
      {circle && (
        <span
          className={cn(
            'w-[7px] h-[7px] rounded-full mr-1',
            variant === 'green' && 'bg-green-600',
            variant === 'blue' && 'bg-blue-500',
            variant === 'yellow' && 'bg-yellow-500',
            variant === 'orange' && 'bg-orange-500',
            variant === 'red' && 'bg-red-500',
            variant === 'gray' && 'bg-gray-400',
            variant === 'violet' && 'bg-violet-500',
            variant === 'default' && 'bg-foreground',
          )}
        />
      )}
      {props.children}
    </div>
  );
}

export {BadgeCustom, badgeVariants};
