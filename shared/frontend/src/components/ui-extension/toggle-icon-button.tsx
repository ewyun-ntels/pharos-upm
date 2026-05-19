import * as React from 'react';
import {cn} from '../../lib';
import {IconButton} from './icon-button';

interface ToggleIconButtonProps {
  icon: React.ReactNode;
  size?: 'icon-xs' | 'icon';
  tooltip: string;
  toggled?: boolean;
  toggledClassName?: string;
  unToggledClassName?: string;
  onClick?: (e: React.MouseEvent<HTMLButtonElement>) => void;
  className?: string;
  disabled?: boolean;
}

export function ToggleIconButton({
  icon,
  size = 'icon',
  tooltip,
  toggled = false,
  toggledClassName,
  unToggledClassName,
  onClick,
  className,
  disabled,
}: ToggleIconButtonProps) {
  const [isToggled, setIsToggled] = React.useState(toggled);

  const handleClick = (e: React.MouseEvent<HTMLButtonElement>) => {
    setIsToggled((prev) => !prev);
    onClick?.(e);
  };

  const defaultToggledClassName = 'bg-primary/20 text-primary hover:bg-primary/30';
  const defaultUnToggledClassName = 'text-foreground';

  return (
    <IconButton
      variant="ghost"
      onClick={handleClick}
      icon={icon}
      size={size}
      className={cn(
        isToggled
          ? (toggledClassName ?? defaultToggledClassName)
          : (unToggledClassName ?? defaultUnToggledClassName),
        className,
      )}
      disabled={disabled}
    >
      {tooltip}
    </IconButton>
  );
}
