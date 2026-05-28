import { useState } from 'react';
import { Popover, PopoverContent, PopoverTrigger, Checkbox, Skeleton } from '@pharos/shared/components/ui';
import { ChevronDown, Check } from 'lucide-react';
import { cn } from '@pharos/shared/lib';

interface MultiSelectFilterProps {
  label: string;
  options: string[];
  selected: string[];
  onChange: (selected: string[]) => void;
  isLoading?: boolean;
  className?: string;
}

export function MultiSelectFilter({
  label,
  options,
  selected,
  onChange,
  isLoading = false,
  className,
}: MultiSelectFilterProps) {
  const [open, setOpen] = useState(false);
  const isAll = selected.length === 0;

  const triggerText = isAll
    ? 'All'
    : selected.length === 1
      ? selected[0]
      : `${selected.length}개 선택`;

  function toggleAll() {
    onChange([]);
  }

  function toggleItem(item: string) {
    if (selected.includes(item)) {
      onChange(selected.filter(s => s !== item));
    } else {
      onChange([...selected, item]);
    }
  }

  if (isLoading) {
    return <Skeleton className="h-8 w-44 rounded-md" />;
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          className={cn(
            'flex h-8 min-w-[11rem] items-center justify-between gap-1 rounded-md border border-input bg-background px-3 text-sm ring-offset-background',
            'hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
            className,
          )}
        >
          <span className="truncate">{triggerText}</span>
          <ChevronDown className="h-4 w-4 shrink-0 opacity-50" />
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-52 p-1" align="end">
        <div className="mb-0.5 px-1 py-0.5 text-[10px] font-medium uppercase tracking-wide text-muted-foreground">
          {label}
        </div>

        {/* All */}
        <label className="flex cursor-pointer items-center gap-2 rounded px-2 py-1.5 text-sm hover:bg-accent">
          <Checkbox
            checked={isAll}
            onCheckedChange={toggleAll}
            id="multi-select-all"
          />
          <span className="font-medium">전체</span>
          {isAll && <Check className="ml-auto h-3.5 w-3.5 text-primary" />}
        </label>

        {options.length > 0 && <div className="my-1 border-t" />}

        <div className="max-h-52 overflow-y-auto">
          {options.map(opt => (
            <label
              key={opt}
              className="flex cursor-pointer items-center gap-2 rounded px-2 py-1.5 text-sm hover:bg-accent"
            >
              <Checkbox
                checked={selected.includes(opt)}
                onCheckedChange={() => toggleItem(opt)}
              />
              <span className="truncate">{opt}</span>
            </label>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  );
}
