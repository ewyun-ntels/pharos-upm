import {Check} from '@pharos/shared/components';
import React, {useState, useEffect} from 'react';
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
import {ChevronDownIcon} from '@pharos/shared/components';
import {cn} from '../../../lib';
import {LabelBadge} from '../label-badge';

// Simple isEqual for string arrays
const isEqual = (a: string[], b: string[]) => {
  if (a.length !== b.length) return false;
  return a.every((val, idx) => val === b[idx]);
};

type Option = Record<string, string>;

export interface MultiSelectProps {
  label?: string;
  options: Option[];
  onValueChange: (value: string[]) => void;
  placeholder?: string;
  defaultSelectAll?: boolean;
  useSelectAll?: boolean;
  maxDisplayedLabels?: number;
  initValue?: string[];
  onSelectAllChange?: (isSelectAll: boolean) => void;
}

const MultiSelect: React.FC<MultiSelectProps> = ({
  label,
  options = [],
  onValueChange,
  placeholder = 'Select options',
  defaultSelectAll = false,
  useSelectAll = false,
  maxDisplayedLabels = Infinity,
  initValue = [],
  onSelectAllChange,
}) => {
  const [open, setOpen] = useState(false);
  const [selected, setSelected] = useState<string[]>([]);
  const [isSelectAll, setIsSelectAll] = useState(false);
  const [openedSelected, setOpenedSelected] = useState<string[]>([]);
  const [isInitialized, setIsInitialized] = useState(false);

  // Initialize selection
  useEffect(() => {
    if (isInitialized) return;
    if (options.length === 0) return;

    // 1. Use initValue if provided
    if (initValue && initValue.length > 0) {
      setIsSelectAll(false);
      setSelected(initValue);
      setIsInitialized(true);
      return;
    }

    // 2. Select all if defaultSelectAll is true
    if (defaultSelectAll) {
      const allValues = options.map((option) => option.value);
      setSelected(allValues);
      setIsSelectAll(true);
      onValueChange(allValues);
      if (onSelectAllChange) {
        onSelectAllChange(true);
      }
      setIsInitialized(true);
      return;
    }

    // 3. Default: select first option
    if (options.length > 0) {
      setIsSelectAll(false);
      setSelected([options[0].value]);
      onValueChange([options[0].value]);
      setIsInitialized(true);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [options]);

  const handleSelect = (value: string) => {
    if (isSelectAll) {
      setIsSelectAll(false);
      setSelected([value]);
    } else {
      setSelected((prev) =>
        prev.includes(value) ? prev.filter((item) => item !== value) : [...prev, value],
      );
    }
  };

  const handleSelectAll = () => {
    setSelected(() => options.map((option) => option.value));
    setIsSelectAll(true);
  };

  const renderSelectedContent = () => {
    if (selected.length === options.length) {
      return <LabelBadge variant="secondary">All</LabelBadge>;
    } else if (selected.length > 0) {
      const displayedLabels = selected.slice(0, maxDisplayedLabels)?.map((item) => (
        <LabelBadge
          key={item}
          variant="secondary"
          className="block w-auto max-w-[150px] mr-1 pl-1 pr-1 overflow-hidden text-ellipsis"
        >
          {options.find((option) => option.value === item)?.label}
        </LabelBadge>
      ));

      if (selected.length > maxDisplayedLabels) {
        displayedLabels.push(
          <LabelBadge key="more" variant="secondary">
            +{selected.length - maxDisplayedLabels}
          </LabelBadge>,
        );
      }

      return displayedLabels;
    } else {
      return placeholder;
    }
  };

  const handlePopoverChange = (isOpen: boolean) => {
    setOpen(isOpen);
    if (isOpen) {
      setOpenedSelected(selected);
    } else {
      if (isEqual(openedSelected, selected)) {
        return;
      }
      onValueChange(selected);
      if (onSelectAllChange) {
        onSelectAllChange(isSelectAll);
      }
    }
  };

  return (
    <div className="flex items-center">
      <Popover open={open} onOpenChange={handlePopoverChange}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            size={'sm'}
            className={cn(
              'w-full h-auto min-h-8 p-[5px] hover:bg-transparent shadow-none box-border cursor-pointer',
            )}
          >
            <div className="w-full h-auto flex flex-wrap items-center gap-2">
              <small className="text-[13px] text-muted-foreground font-right w-fit text-nowrap leading-none">
                {label}
              </small>
              {renderSelectedContent()}
            </div>
            <ChevronDownIcon className="text-muted-foreground opacity-50 pointer-events-none size-4 shrink-0 translate-y-0.5 transition-transform duration-200" />
          </Button>
        </PopoverTrigger>
        <PopoverContent align={'start'} className="w-[200px] p-0 z-[150]">
          <Command>
            <CommandInput placeholder="Search" />
            <CommandList className="max-h-[320px]">
              <CommandEmpty>No option found.</CommandEmpty>
              <CommandGroup>
                {useSelectAll && (
                  <>
                    <CommandItem onSelect={handleSelectAll}>
                      <div className="flex items-center">
                        <div className="mr-2 flex h-4 w-4 items-center justify-center">
                          {isSelectAll && <Check className="h-4 w-4" />}
                        </div>
                        Select All
                      </div>
                    </CommandItem>
                    <CommandSeparator />
                  </>
                )}
                {options.map((option) => (
                  <CommandItem key={option.value} onSelect={() => handleSelect(option.value)}>
                    <div className="flex items-center">
                      <div className="mr-2 flex h-4 w-4 items-center justify-center">
                        {!isSelectAll && selected.includes(option.value) && (
                          <Check className="h-4 w-4" />
                        )}
                      </div>
                      {option.label}
                    </div>
                  </CommandItem>
                ))}
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
    </div>
  );
};

export default MultiSelect;
