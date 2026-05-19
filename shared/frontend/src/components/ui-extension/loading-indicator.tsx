import {RefreshCw} from 'lucide-react';
import {cn} from '../../lib';

interface LoadingIndicatorProps {
  className?: string;
}

export function LoadingIndicator({className}: LoadingIndicatorProps) {
  return (
    <div
      className={cn(
        'flex h-full items-center justify-center gap-2 text-sm text-muted-foreground',
        className,
      )}
    >
      <RefreshCw className="h-4 w-4 animate-spin" />
      Loading...
    </div>
  );
}
