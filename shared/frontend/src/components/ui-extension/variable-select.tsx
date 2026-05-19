import * as React from 'react';
import {CheckIcon} from '@radix-ui/react-icons';
import {useState, useEffect} from 'react';
import {ChevronDown, XIcon, RefreshCw} from 'lucide-react';
import {
  Badge,
  Button,
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '../ui';
import {cn} from '../../lib';

interface VariableSelectOption {
  label: string;
  value: string;
  icon?: React.ComponentType<{className?: string}>;
}

interface VariableSelectProps {
  label?: string;
  options: VariableSelectOption[];
  value?: string | string[]; // Single or multiple values
  onChange?: (value: string | string[]) => void;
  multiple?: boolean; // Enable multi-select
  searchable?: boolean; // Enable search (default: auto based on options length)
  placeholder?: string;
  className?: string;
  disabled?: boolean;
  align?: 'start' | 'center' | 'end';
  showLabel?: boolean; // Show label inside button with divider (Grafana style)
  isLoading?: boolean; // Show loading spinner
  displayMode?: 'count' | 'value-single' | 'value-all' | 'value-threshold'; // How to display selected values in multi-select
  // - 'count': Always show count (e.g., "3")
  // - 'value-single': Show value if 1, count if 2+ (e.g., "server-1" or "3")
  // - 'value-all': Show all values comma-separated (e.g., "server-1, server-2, server-3")
  // - 'value-threshold': Show values up to threshold, then count (e.g., "server-1" or "5")
  displayThreshold?: number; // Threshold for 'value-threshold' mode (default: 1)
  showAllOption?: boolean; // Include "All" option that selects all values (only for multi-select)
  defaultAllSelected?: boolean; // When showAllOption=true, default to "All" selected (only on initial render, only for multi-select)
  customAllValue?: string; // Custom value to use when "All" is selected (e.g., "*" or ".*"), instead of all individual values
  autoSelectFirstOption?: boolean; // Auto-select first option if no value provided (default: false for multi-select, can be used for single-select)
}

interface FilterItems {
  selectedSortedOptions: VariableSelectOption[];
  unselectedSortedOptions: VariableSelectOption[];
}

const FilterItem = (
  option: VariableSelectOption,
  isSelected: boolean,
  handleSelect: (value: string, isSelected: boolean) => void,
  showCheckbox: boolean,
) => {
  return (
    <CommandItem
      key={option.value}
      onSelect={() => handleSelect(option.value, isSelected)}
      className="aria-selected:bg-tranparent hover:bg-accent cursor-pointer"
    >
      {showCheckbox && (
        <div
          className={cn(
            'flex h-4 w-4 items-center justify-center rounded-sm border border-primary cursor-pointer',
            isSelected
              ? 'bg-primary text-primary-foreground'
              : 'opacity-50 [&_svg]:invisible cursor-pointer',
          )}
        >
          <CheckIcon className={cn('h-4 w-4')} />
        </div>
      )}
      {option.icon && <option.icon className="mr-2 h-4 w-4 text-muted-foreground" />}
      <span className="cursor-pointer">{option.label}</span>
      {isSelected && !showCheckbox && (
        <CheckIcon className="ml-auto h-4 w-4" />
      )}
    </CommandItem>
  );
};

const VariableSelect = React.forwardRef<HTMLDivElement, VariableSelectProps>(
  (
    {
      label = 'Select',
      options,
      value,
      onChange,
      multiple = false,
      searchable,
      placeholder,
      className,
      align = 'start',
      disabled = false,
      showLabel = true, // Default: show label inside (Grafana style)
      isLoading = false,
      displayMode = 'value-threshold', // Default: show values up to threshold
      displayThreshold = 1, // Default: 1 (show only 1 value, then count)
      showAllOption = false,
      defaultAllSelected = false,
      customAllValue,
      autoSelectFirstOption = false,
    },
    _ref,
  ) => {
    // Initialize selected values based on single/multiple mode
    const [selectedValues, setSelectedValues] = useState<Set<string>>(() => {
      // Priority 1: If value is provided, use it
      if (value) {
        if (Array.isArray(value)) return new Set(value);
        return new Set([value]);
      }
      
      // Priority 2: For multi-select with showAllOption + defaultAllSelected
      // Set internal state to include "All" for UI, but onChange will send actual values
      if (!value && showAllOption && multiple && defaultAllSelected) {
        const allValues = options.map(opt => opt.value);
        return new Set(['All', ...allValues]);
      }
      
      // Priority 3: Auto-select first option if enabled and options exist
      if (!value && autoSelectFirstOption && options.length > 0) {
        // For multi-select, select first as array; for single-select, just first value
        return new Set([options[0].value]);
      }
      
      // Default: empty
      return new Set();
    });

    const [popoverIsOpen, setPopoverIsOpen] = useState<boolean>(false);
    const [prevPopoverIsOpen, setPrevPopoverIsOpen] = useState<boolean>(false);
    const [filterItems, setFilterItems] = useState<FilterItems | undefined>(undefined);
    const [isInitialized, setIsInitialized] = useState<boolean>(false);

    // Add "All" option if enabled
    const allOptions = React.useMemo(() => {
      if (showAllOption && multiple) {
        return [{label: 'All', value: 'All'}, ...options];
      }
      return options;
    }, [showAllOption, multiple, options]);

    // Notify parent of initial default selection (only once)
    useEffect(() => {
      if (!isInitialized && !value) {
        let shouldNotify = false;
        let initialSelection: string | string[] | null = null;
        
        // Case 1: defaultAllSelected for multi-select with showAllOption
        if (defaultAllSelected && showAllOption && multiple) {
          const allValues = options.map(opt => opt.value);
          
          // If customAllValue is provided, use it; otherwise send all values
          if (customAllValue !== undefined && customAllValue !== '') {
            initialSelection = [customAllValue];
          } else if (allValues.length === 0) {
            return;
          } else {
            initialSelection = allValues;
          }
          shouldNotify = true;
        }
        // Case 2: autoSelectFirstOption when options exist
        else if (autoSelectFirstOption && options.length > 0) {
          initialSelection = multiple ? [options[0].value] : options[0].value;
          shouldNotify = true;
        }
        
        if (shouldNotify && initialSelection && onChange) {
          onChange(initialSelection);
        }
        setIsInitialized(true);
      }
    }, [isInitialized, defaultAllSelected, showAllOption, multiple, autoSelectFirstOption, value, options, onChange, customAllValue]);

    // Handle cascade: when options change and autoSelectFirstOption is enabled
    useEffect(() => {
      // Only trigger if:
      // 1. autoSelectFirstOption is enabled
      // 2. Options exist
      // 3. Not the initial render (isInitialized is true)
      if (!isInitialized || !autoSelectFirstOption || options.length === 0) {
        return;
      }

      if (multiple && showAllOption && defaultAllSelected && value === undefined) {
        return;
      }

      const optionValues = new Set(options.map(opt => opt.value));
      
      // Check if current value(s) are valid in the new options
      let needsReset = false;
      
      if (!value || (Array.isArray(value) && value.length === 0)) {
        // No value selected
        needsReset = true;
      } else if (Array.isArray(value)) {
        if (showAllOption && customAllValue && value.length === 1 && value[0] === customAllValue) {
          return;
        }

        // Multi-select: check if ALL values are invalid (not just one)
        const allValuesInvalid = value.every(v => !optionValues.has(v));
        if (allValuesInvalid) {
          needsReset = true;
        }
      } else {
        // Single-select: check if value is invalid
        if (!optionValues.has(value)) {
          needsReset = true;
        }
      }
      
      if (needsReset && onChange) {
        const initialSelection = multiple ? [options[0].value] : options[0].value;
        onChange(initialSelection);
      }
    }, [options, isInitialized, autoSelectFirstOption, multiple, showAllOption, defaultAllSelected, customAllValue, onChange, value]);

    // Sync with external value changes
    // selectedValues는 deps에서 제외: 내부 상태(handleSelect) 변경 시 effect가 실행되면
    // value prop이 아직 갱신되지 않아 stale value로 "All" 선택을 되돌리는 race condition 발생.
    // eslint-disable-next-line react-hooks/exhaustive-deps
    useEffect(() => {
      // Distinguish between undefined (not yet set, waiting for auto-select) and empty array (explicitly cleared)
      if (value === undefined) {
        // Value is undefined - waiting for auto-select from useDashboardQuery
        // Don't clear internal selection yet
        return;
      }

      if (!value || (Array.isArray(value) && value.length === 0)) {
        setSelectedValues(new Set());
        return;
      }

      let newSet: Set<string>;

      if (Array.isArray(value)) {
        // Multi-select: check if value is customAllValue
        if (showAllOption && customAllValue && value.length === 1 && value[0] === customAllValue) {
          // Value is customAllValue - select "All" + all options in UI
          const allValues = options.map(opt => opt.value);
          newSet = new Set(['All', ...allValues]);
        } else if (showAllOption && options.length > 0) {
          // Check if all options are selected
          const allValues = options.map(opt => opt.value);
          const allSelected = allValues.every(v => value.includes(v));

          if (allSelected) {
            // All options selected - also show "All" in UI
            newSet = new Set(['All', ...value]);
          } else {
            // Partial selection - exclude "All"
            newSet = new Set(value.filter(v => v !== 'All'));
          }
        } else {
          newSet = new Set(value);
        }
      } else {
        newSet = new Set([value]);
      }

      setSelectedValues(newSet);
    }, [value, showAllOption, customAllValue, options, multiple]); // selectedValues 제외: race condition 방지

    const handleSelect = (selectValue: string, isSelected: boolean) => {
      if (disabled) return;

      let newSelectedValues: Set<string>;

      if (multiple) {
        // Handle "All" option
        if (showAllOption && selectValue === 'All') {
          const allNonAllValues = options.map(opt => opt.value);
          
          if (isSelected) {
            // "All" is currently selected -> deselect everything
            newSelectedValues = new Set();
          } else {
            // "All" is not selected -> select all (including "All" option itself)
            newSelectedValues = new Set(['All', ...allNonAllValues]);
          }
        } else {
          // Regular option selected
          newSelectedValues = new Set(selectedValues);
          if (isSelected) {
            // Deselecting an option -> remove it and remove "All"
            newSelectedValues.delete(selectValue);
            newSelectedValues.delete('All');
          } else {
            // Selecting an option -> add it
            newSelectedValues.add(selectValue);
            
            // Check if all options (except "All") are now selected
            const allNonAllValues = options.map(opt => opt.value);
            const selectedNonAll = Array.from(newSelectedValues).filter(v => v !== 'All');
            
            if (showAllOption && selectedNonAll.length === allNonAllValues.length) {
              // All options selected -> also select "All"
              newSelectedValues.add('All');
            }
          }
        }
      } else {
        // Single-select mode
        newSelectedValues = new Set([selectValue]);
        setPopoverIsOpen(false); // Close on select for single mode
      }

      setSelectedValues(newSelectedValues);

      // Call onChange with appropriate format
      if (onChange) {
        if (multiple) {
          // If "All" is selected and customAllValue is provided, use it instead
          if (newSelectedValues.has('All') && customAllValue !== undefined && customAllValue !== '') {
            onChange([customAllValue]);
          } else if (newSelectedValues.has('All')) {
            // "All" is selected but no customAllValue - send all actual values (excluding "All")
            const allNonAllValues = options.map(opt => opt.value);
            onChange(allNonAllValues);
          } else {
            // Send only selected values (excluding "All" if it somehow remains)
            const valuesToSend = Array.from(newSelectedValues).filter(v => v !== 'All');
            onChange(valuesToSend);
          }
        } else {
          onChange(newSelectedValues.size > 0 ? Array.from(newSelectedValues)[0] : '');
        }
      }
    };

    const clearSelection = (e?: React.MouseEvent) => {
      if (disabled) return;
      if (e) e.stopPropagation();
      setSelectedValues(new Set());
      if (onChange) {
        onChange(multiple ? [] : '');
      }
      setPopoverIsOpen(false);
    };

    useEffect(() => {
      if (popoverIsOpen !== prevPopoverIsOpen) {
        const selectedSortedOptions: VariableSelectOption[] = [];
        const unselectedSortedOptions: VariableSelectOption[] = [];
        const optionValuesSet = new Set(allOptions.map((option) => option.value));

        // loop가 많이 돌기는 하지만 options의 길이가 많이 길지 않아서 큰 문제는 없을 것으로 판단됨
        selectedValues.forEach((value) => {
          if (!optionValuesSet.has(value)) {
            selectedSortedOptions.push({label: value, value});
          }
        });

        allOptions.forEach((option) => {
          if (selectedValues.has(option.value)) {
            selectedSortedOptions.push(option);
          } else {
            unselectedSortedOptions.push(option);
          }
        });
        setFilterItems({selectedSortedOptions, unselectedSortedOptions});
        setPrevPopoverIsOpen(popoverIsOpen);
      }
    }, [popoverIsOpen, prevPopoverIsOpen, allOptions, selectedValues]);

    // Auto-enable search if options > 9
    const shouldShowSearch = searchable ?? allOptions.length > 9;

    // Render button content
    const renderButtonContent = () => {
      // Single select display
      if (!multiple) {
        if (selectedValues.size === 0) {
          return (
            <span className="text-muted-foreground truncate">
              {placeholder || (showLabel ? '' : label)}
            </span>
          );
        }
        const selected = options.find(opt => selectedValues.has(opt.value));
        return (
          <span className="truncate font-normal">
            {selected?.label || ''}
          </span>
        );
      }

      // Multi select display
      if (selectedValues.size === 0) {
        return (
          <span className="text-muted-foreground truncate">
            {placeholder || (showLabel ? '' : label)}
          </span>
        );
      }

      // Special case: If "All" is selected, only show "All" badge
      if (showAllOption && selectedValues.has('All')) {
        return (
          <Badge variant="secondary" className="font-normal">
            All
          </Badge>
        );
      }

      // Display based on mode
      if (displayMode === 'value-single' && selectedValues.size === 1) {
        // Show single value as badge
        const selected = options.find(opt => selectedValues.has(opt.value));
        return (
          <Badge variant="secondary" className="font-normal">
            {selected?.label || ''}
          </Badge>
        );
      } else if (displayMode === 'value-all') {
        // Show all values as individual badges
        const selectedOptions = Array.from(selectedValues)
          .map(val => options.find(opt => opt.value === val))
          .filter(Boolean);
        return (
          <div className="flex gap-1 flex-wrap">
            {selectedOptions.map((opt) => (
              <Badge key={opt!.value} variant="secondary" className="font-normal">
                {opt!.label}
              </Badge>
            ))}
          </div>
        );
      } else if (displayMode === 'value-threshold') {
        // Show values up to threshold as badges, THEN remaining count badge
        const selectedOptions = Array.from(selectedValues)
          .filter(val => val !== 'All') // Filter out "All" from display
          .map(val => options.find(opt => opt.value === val))
          .filter(Boolean);
        
        if (selectedOptions.length <= displayThreshold) {
          // Within threshold: show all as individual badges
          return (
            <div className="flex gap-1 flex-wrap">
              {selectedOptions.map((opt) => (
                <Badge key={opt!.value} variant="secondary" className="font-normal">
                  {opt!.label}
                </Badge>
              ))}
            </div>
          );
        } else {
          // Over threshold: show first N badges + remaining count badge
          const visibleOptions = selectedOptions.slice(0, displayThreshold);
          const remainingCount = selectedOptions.length - displayThreshold;
          return (
            <div className="flex gap-1 flex-wrap">
              {visibleOptions.map((opt) => (
                <Badge key={opt!.value} variant="secondary" className="font-normal">
                  {opt!.label}
                </Badge>
              ))}
              <Badge variant="secondary" className="font-normal">
                +{remainingCount}
              </Badge>
            </div>
          );
        }
      } else {
        // Show count badge
        return (
          <Badge variant="secondary" className="font-normal">
            {selectedValues.size}
          </Badge>
        );
      }
    };

    return (
      <div 
        ref={_ref} 
        className={cn(
          'flex items-center h-8',
          showLabel ? 'rounded-md border border-input bg-background hover:border-input focus-within:ring-1 focus-within:ring-ring' : '',
          disabled && 'opacity-50 cursor-not-allowed',
          className
        )}
      >
        {/* Label area (not clickable) */}
        {showLabel && label && (
          <div className="flex items-center text-[13px] font-semibold gap-1.5 px-3 border-r border-border">
            {label}
            {isLoading && (
              <RefreshCw className="h-3 w-3 animate-spin text-muted-foreground" />
            )}
          </div>
        )}
        
        {/* Value area (clickable button) */}
        <Popover open={popoverIsOpen} onOpenChange={(open) => setPopoverIsOpen(open)}>
          <PopoverTrigger asChild>
            <button
              type="button"
              disabled={disabled}
              className={cn(
                'flex items-center justify-between gap-2 flex-1 px-3 py-1',
                'text-sm bg-transparent hover:bg-accent/50 transition-colors',
                'focus:outline-none focus-visible:ring-0',
                'disabled:cursor-not-allowed disabled:opacity-50'
              )}
            >
              <span className="flex-1 truncate text-left">
                {renderButtonContent()}
              </span>
              {multiple && selectedValues.size > 0 && (
                <XIcon
                  className="w-4 h-4 opacity-50 hover:opacity-100 cursor-pointer shrink-0"
                  onClick={clearSelection}
                />
              )}
              <ChevronDown className="w-4 h-4 opacity-50 shrink-0" />
            </button>
          </PopoverTrigger>
          <PopoverContent
            className="w-auto min-w-50 max-w-100 p-0"
            align={align}
          >
            <Command>
              {shouldShowSearch && (
                <CommandInput placeholder={placeholder || `Search ${label}...`} />
              )}
              <CommandList>
                <CommandEmpty>
                  <span className="text-muted-foreground">No results found.</span>
                </CommandEmpty>
                <CommandGroup>
                  {filterItems?.selectedSortedOptions.map((option) => {
                    const isSelected = selectedValues.has(option.value);
                    return FilterItem(option, isSelected, handleSelect, multiple);
                  })}
                  {filterItems !== undefined &&
                    filterItems.selectedSortedOptions.length > 0 &&
                    filterItems.unselectedSortedOptions.length > 0 && (
                      <CommandSeparator className="mt-1 mb-1" />
                    )}
                  {filterItems?.unselectedSortedOptions.map((option) => {
                    const isSelected = selectedValues.has(option.value);
                    return FilterItem(option, isSelected, handleSelect, multiple);
                  })}
                </CommandGroup>
              </CommandList>
              {multiple && selectedValues.size > 0 && (
                <>
                  <CommandSeparator />
                  <CommandGroup>
                    <CommandItem>
                      <Button
                        onClick={clearSelection}
                        variant="ghost"
                        className="h-6 rounded-none w-full justify-center text-center"
                      >
                        Clear selection
                      </Button>
                    </CommandItem>
                  </CommandGroup>
                </>
              )}
            </Command>
          </PopoverContent>
        </Popover>
      </div>
    );
  },
);
VariableSelect.displayName = 'VariableSelect';

export type {VariableSelectProps, VariableSelectOption};
export {VariableSelect};
