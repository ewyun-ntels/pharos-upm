import { useTheme } from '@pharos/core/providers/theme-provider';
import logoSvg from '../../../../images/logo.svg';
import logoDarkSvg from '../../../../images/logo-dark.svg';

interface LogoProps {
  className?: string;
  width?: number;
  height?: number;
}

export function Logo({ className = '', width, height }: LogoProps) {
  const { resolvedTheme } = useTheme();

  const style = width || height
    ? {
        width: width ? `${width}px` : undefined,
        height: height ? `${height}px` : undefined,
      }
    : undefined;

  return (
    <img
      src={resolvedTheme === 'dark' ? logoDarkSvg : logoSvg}
      alt="Logo"
      className={`object-contain ${className}`}
      style={style}
    />
  );
}
