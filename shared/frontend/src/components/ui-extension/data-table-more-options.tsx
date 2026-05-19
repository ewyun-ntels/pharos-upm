import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@pharos/shared/components/ui';
import {Button} from '@pharos/shared/components/ui';
import {MoreVertical, Settings, Copy, Trash, View} from 'lucide-react';
import {cn} from '../../lib';

interface BoxDropdownProps {
  onView?: () => void;
  onEdit?: () => void;
  onDupl?: () => void;
  onDelete?: () => void;
}

export const BoxDropdown: React.FC<BoxDropdownProps> = ({onView, onEdit, onDupl, onDelete}) => {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className="top-0 right-0 z-10 h-8 w-8 p-0 text-muted-foreground cursor-pointer rounded-none"
        >
          <MoreVertical className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-32 p-1 text-sm">
        <DropdownMenuItem onClick={onView} className={cn('gap-2 px-2 py-1.5', !onView && 'hidden')}>
          <View className="w-4 h-4" />
          View
        </DropdownMenuItem>
        <DropdownMenuItem onClick={onEdit} className={cn('gap-2 px-2 py-1.5', !onEdit && 'hidden')}>
          <Settings className="w-4 h-4" />
          Edit
        </DropdownMenuItem>
        <DropdownMenuItem onClick={onDupl} className={cn('gap-2 px-2 py-1.5', !onDupl && 'hidden')}>
          <Copy className="w-4 h-4" />
          Duplicate
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={onDelete}
          className={cn('gap-2 px-2 py-1.5 text-red-600', !onDelete && 'hidden')}
        >
          <Trash className="w-4 h-4" />
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};
