import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@pharos/shared/components/ui';
import { Button } from '@pharos/shared/components/ui';
import { MoreVertical, Settings, Copy, Trash, View } from '@pharos/shared/components';

interface BoxDropdownProps {
  onView: () => void;
  onSettings: () => void;
  onDuplicate: () => void;
  onDelete: () => void;
  onOpenChange: (open: boolean) => void;
  isRow: boolean; // 추가된 prop
}

export const BoxDropdown: React.FC<BoxDropdownProps> = ({
  onView,
  onSettings,
  onDuplicate,
  onDelete,
  onOpenChange,
  isRow, // isRow prop 사용
}) => {
  if (isRow) return null;
  return (
    <DropdownMenu onOpenChange={onOpenChange}>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className="absolute top-0 right-0 z-10 h-8 w-8 p-0 text-muted-foreground cursor-pointer rounded-none"
        >
          <MoreVertical className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-32 p-1 text-sm">
        <DropdownMenuItem onClick={onView} className="gap-2 px-2 py-1.5">
          <View className="w-4 h-4" />
          View
        </DropdownMenuItem>
        <DropdownMenuItem onClick={onSettings} className="gap-2 px-2 py-1.5">
          <Settings className="w-4 h-4" />
          Edit
        </DropdownMenuItem>
        <DropdownMenuItem onClick={onDuplicate} className="gap-2 px-2 py-1.5">
          <Copy className="w-4 h-4" />
          Duplicate
        </DropdownMenuItem>
        <DropdownMenuItem onClick={onDelete} className="gap-2 px-2 py-1.5 text-red-600">
          <Trash className="w-4 h-4" />
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};
