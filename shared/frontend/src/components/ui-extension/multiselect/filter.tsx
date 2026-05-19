import * as React from 'react';
import {CheckIcon, ChevronDownIcon, Cross2Icon} from '@radix-ui/react-icons';
import {Column} from '@tanstack/react-table';

import {cn} from '../../../lib';
import {Button} from '@pharos/shared/components/ui';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from '@pharos/shared/components/ui';
import {Popover, PopoverContent, PopoverTrigger} from '@pharos/shared/components/ui';
import {useState, useEffect} from 'react';
import {LabelBadge} from '../label-badge';

interface MultiSelectFilterOption {
  label: string;
  value: string;
  icon?: React.ComponentType<{className?: string}>;
  unsetFilter?: (label: string, value: string) => void;
  setFilter?: (label: string, value: string) => void;
}

interface MultiSelectFilterProps<TData, TValue> {
  column?: Column<TData, TValue>;
  title?: string;
  options: MultiSelectFilterOption[];
}

function MultiSelectFilter<TData, TValue>({
  column,
  title,
  options,
}: MultiSelectFilterProps<TData, TValue>) {
  const [selectedValues, setSelectedValues] = useState<Set<string>>(
    new Set(column?.getFilterValue() as string[]),
  );

  useEffect(() => {
    column?.setFilterValue(Array.from(selectedValues));
  }, [selectedValues, column]);

  const handleSelect = (option: MultiSelectFilterOption, isSelected: boolean) => {
    const newSelectedValues = new Set(selectedValues);
    if (isSelected) {
      newSelectedValues.delete(option.value);
      option.unsetFilter?.(option.label, option.value);
    } else {
      newSelectedValues.add(option.value);
      option.setFilter?.(option.label, option.value);
    }
    setSelectedValues(newSelectedValues);
  };

  const handleBadgeClose = (value: string) => {
    const newSelectedValues = new Set(selectedValues);
    newSelectedValues.delete(value);
    setSelectedValues(newSelectedValues);
  };

  const clearFilters = () => {
    setSelectedValues(new Set());
    column?.setFilterValue(undefined);
  };

  // const selectedBadgeAttr: BadgeProps = { variant: "secondary", className: "rounded-sm px-1 font-normal" };

  return (
    <Popover>
      {/* PopoverTrigger (Button) */}
      <div>
        <PopoverTrigger asChild>
          <Button variant="outline" size="sm" className="h-9 text-sm font-normal border w-full">
            {title}
            <ChevronDownIcon className="ml-auto h-4 w-4" />
          </Button>
        </PopoverTrigger>

        {/* Selected Badges below the button */}
        {selectedValues.size > 0 && (
          <div className="mt-3 flex flex-wrap gap-1 items-center">
            {options
              .filter((option) => selectedValues.has(option.value))
              .map((option) => (
                <LabelBadge
                  variant={'secondary'}
                  key={option.value}
                  className={cn(
                    'rounded-sm px-1 font-normal',
                    'flex items-center h-6 px-2 py-1 text-zinc-500 border border-color-border-input bg-secondary rounded-xl shadow-sm',
                  )}
                >
                  {option.label}
                  <button
                    className="ml-1 rounded-full hover:bg-primary/10"
                    onClick={() => handleBadgeClose(option.value)}
                  >
                    <Cross2Icon className="h-4 w-4 text-secondary-foreground hover:text-primary" />
                  </button>
                </LabelBadge>
              ))}
            {/* Clear All button */}
            <Button
              variant="ghost"
              size="sm"
              className="ml-2 h-6 px-2 text-destructive"
              onClick={clearFilters}
            >
              필터 삭제
            </Button>
          </div>
        )}
      </div>

      {/* TODO fasted-filter와 중복되는 로직 제거 (109 line) */}
      <PopoverContent className="w-[200px] p-0" align="start">
        <Command>
          <CommandInput placeholder={title} />
          <CommandList>
            <CommandEmpty>No results found.</CommandEmpty>
            <CommandGroup>
              {options.map((option) => {
                const isSelected = selectedValues.has(option.value);
                return (
                  <CommandItem key={option.value} onSelect={() => handleSelect(option, isSelected)}>
                    <div
                      className={cn(
                        'mr-2 flex h-4 w-4 items-center justify-center rounded-sm border border-primary',
                        isSelected
                          ? 'bg-primary text-primary-foreground'
                          : 'opacity-50 [&_svg]:invisible',
                      )}
                    >
                      <CheckIcon className={cn('h-4 w-4')} />
                    </div>
                    {option.icon && <option.icon className="mr-2 h-4 w-4 text-muted-foreground" />}
                    <span>{option.label}</span>
                    {column?.getFacetedUniqueValues()?.get(option.value) && (
                      <span className="ml-auto flex h-4 w-4 items-center justify-center font-mono text-xs">
                        {column.getFacetedUniqueValues().get(option.value)}
                      </span>
                    )}
                  </CommandItem>
                );
              })}
            </CommandGroup>
            {selectedValues.size > 0 && (
              <>
                <CommandSeparator />
                <CommandGroup>
                  <CommandItem onSelect={clearFilters} className="justify-center text-center">
                    Clear filters
                  </CommandItem>
                </CommandGroup>
              </>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}

export type {MultiSelectFilterProps, MultiSelectFilterOption};

export {MultiSelectFilter};
