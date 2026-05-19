import * as React from 'react';
import {CheckIcon} from '@radix-ui/react-icons';
import {Column} from '@tanstack/react-table';

import {useState, useEffect} from 'react';
import {ChevronDown, ChevronsUpDown, XIcon} from 'lucide-react';
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

interface FacetedFilterOption {
  label: string;
  value: string;
  count?: number;
  icon?: React.ComponentType<{className?: string}>;
  unsetFilter?: (label: string, value: string) => void;
  setFilter?: (label: string, value: string) => void;
}

interface FacetedFilterRefProps {
  clearFilters: () => void;
  getSelectedFilters: () => string[];
}

interface FacetedFilterProps<TData, TValue> {
  column?: Column<TData, TValue>;
  title?: string;
  options: FacetedFilterOption[];
  defaultValues?: string[];
  selectedValuesFn?: (value: string[]) => void;
  onChange?: () => void;
  enableFilterStyle?: boolean;
  enableNamespaceFilter?: boolean;
  optionWidth?: boolean;
  className?: string;
  disabled?: boolean;
  align?: 'start' | 'center' | 'end';
}

interface FilterItems {
  selectedSortedOptions: FacetedFilterOption[];
  unselectedSortedOptions: FacetedFilterOption[];
}

const FilterItem = (
  option: FacetedFilterOption,
  isSelected: boolean,
  handleSelect: (value: string, isSelected: boolean) => void,
  column: Column<any, any> | undefined,
) => {
  return (
    // command option의 item 1줄에 대한 레이아웃 : 체크박스, value 포함
    <CommandItem
      key={option.value}
      onSelect={() => handleSelect(option.value, isSelected)}
      className="aria-selected:bg-tranparent hover:bg-accent cursor-pointer"
    >
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
      {option.icon && <option.icon className="mr-2 h-4 w-4 text-muted-foreground" />}
      <span className="cursor-pointer pr-4">{option.label}</span>
      <span className="cursor-pointer ml-auto text-xs font-mono capitalize">
        {option.count?.toLocaleString()}
      </span>
      {column?.getFacetedUniqueValues()?.get(option.value) && (
        <span className="ml-auto flex h-4 w-4 items-center justify-end font-mono text-xs capitalize">
          {column.getFacetedUniqueValues().get(option.value)}
        </span>
      )}
    </CommandItem>
  );
};

const FacetedFilter = React.forwardRef<FacetedFilterRefProps, FacetedFilterProps<any, any>>(
  (
    {
      column,
      title = 'Filter',
      defaultValues,
      options,
      onChange,
      selectedValuesFn,
      enableFilterStyle,
      enableNamespaceFilter,
      optionWidth,
      className,
      align = 'start',
      disabled = false,
    },
    _ref,
  ) => {
    const [selectedValues, setSelectedValues] = useState<Set<string>>(
      column?.getFilterValue()
        ? new Set(column?.getFilterValue() as string[])
        : new Set(defaultValues),
    );
    const [popoverIsOpen, setPopoverIsOpen] = useState<boolean>(false);
    const [prevPopoverIsOpen, setPrevPopoverIsOpen] = useState<boolean>(false);
    const [filterItems, setFilterItems] = useState<FilterItems | undefined>(undefined);
    const [filterEmpty, setFilterEmpty] = useState<String>(title);

    const setFilterValues = (values: string[]) => {
      selectedValuesFn && selectedValuesFn(values);
    };

    const handleSelect = (value: string, isSelected: boolean) => {
      const newSelectedValues = new Set(selectedValues);
      const option = options.find((option) => option.value === value);

      if (isSelected) {
        newSelectedValues.delete(value);
        option?.unsetFilter?.(option.label, option.value);
      } else {
        newSelectedValues.add(value);
        option?.setFilter?.(option.label, option.value);
      }

      setFilterValues(Array.from(newSelectedValues));
      setSelectedValues(newSelectedValues);

      if (selectedValuesFn) {
        selectedValuesFn(Array.from(newSelectedValues));
      }

      if (onChange) {
        onChange();
      }

      if (disabled) return;
    };

    const clearFilters = (e?: React.MouseEvent) => {
      if (disabled) return;
      if (e) e.stopPropagation();
      setSelectedValues(new Set());
      setFilterValues([]);
      column?.setFilterValue(undefined);
      setPopoverIsOpen(false);
    };

    if (_ref) {
      // eslint-disable-next-line react-hooks/rules-of-hooks
      React.useImperativeHandle(_ref, () => ({
        clearFilters,
        getSelectedFilters: () => Array.from(selectedValues),
      }));
    }

    useEffect(() => {
      if (selectedValues.size === 0) {
        column?.setFilterValue(undefined);
      } else {
        column?.setFilterValue(Array.from(selectedValues));
      }
      if (onChange) {
        onChange();
      }
      // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [selectedValues, column]);

    useEffect(() => {
      if (popoverIsOpen !== prevPopoverIsOpen) {
        const selectedSortedOptions: FacetedFilterOption[] = [];
        const unselectedSortedOptions: FacetedFilterOption[] = [];
        const optionValuesSet = new Set(options.map((option) => option.value));

        // loop가 많이 돌기는 하지만 options의 길이가 많이 길지 않아서 큰 문제는 없을 것으로 판단됨
        selectedValues.forEach((value) => {
          if (!optionValuesSet.has(value)) {
            selectedSortedOptions.push({label: value, value});
          }
        });

        options.forEach((option) => {
          if (selectedValues.has(option.value)) {
            selectedSortedOptions.push(option);
          } else {
            unselectedSortedOptions.push(option);
          }
        });
        setFilterItems({selectedSortedOptions, unselectedSortedOptions});
        setPrevPopoverIsOpen(popoverIsOpen);

        if (filterItems?.unselectedSortedOptions.length == 0) {
          setFilterEmpty(title);
        }
      }
      // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [popoverIsOpen, prevPopoverIsOpen, options, selectedValues]);

    return (
      <div
        className={cn(
          'items-center rounded-md bg-background shadow-sm flex',
          selectedValues.size !== 0 && '',
        )}
      >
        <Popover onOpenChange={(open) => setPopoverIsOpen(open)}>
          <PopoverTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              disabled={disabled}
              className={cn(
                'flex flex-1 justify-between gap-2 h-8 border ',
                enableFilterStyle
                  ? 'border-solid h-9 text-sm font-normal min-w-43.25'
                  : 'border-dashed h-8',
                optionWidth ? 'min-w-35' : '',
                className,
              )}
            >
              <span
                className={cn(
                  'capitalize font-medium',
                  selectedValues.size > 0 && enableNamespaceFilter ? 'text-primary' : '',
                )}
              >
                {title}
              </span>
              {selectedValues.size > 0 && (
                <Badge
                  variant={'secondary'}
                  className={
                    'flex mr-auto gap-1.5 rounded-sm px-1 font-normal bg-primary/20 hover:bg-primary/30'
                  }
                  key={`filter-pop-trigger-${title}`}
                >
                  {' '}
                  {selectedValues.size}
                  {/* Button 컴포넌트 내부에 onClick 이벤트가 안되는 이유는 컴포넌트에 !pointer-events-auto 넣기 */}
                  <XIcon
                    key={`${title}-${selectedValues.size}`}
                    className="w-4 h-4 pointer-events-auto!"
                    onClick={clearFilters}
                  />
                </Badge>
              )}
              {enableFilterStyle ? (
                <ChevronsUpDown className="h-3.25! w-3.25! opacity-60 shrink-0" />
              ) : (
                <ChevronDown className="h-4 w-4 opacity-50 shrink-0" />
              )}
            </Button>
          </PopoverTrigger>
          <PopoverContent
            className={cn(
              'flex flex-1 w-auto min-w-50 max-w-125 p-0',
              enableFilterStyle ? 'min-w-75' : '',
              optionWidth ? 'min-w-35 min-h-9' : '',
              className,
            )}
            align={align}
          >
            <Command>
              {options.length > 9 && (
                <CommandInput placeholder={title} className="placeholder:capitalize" />
              )}
              <CommandList>
                {filterItems?.unselectedSortedOptions.length === 0 &&
                !filterItems?.selectedSortedOptions?.length ? null : (
                  <CommandEmpty>
                    <span className="text-muted-foreground">검색 결과가 없습니다.</span>
                  </CommandEmpty>
                )}
                <CommandGroup>
                  {filterItems?.unselectedSortedOptions.length === 0 &&
                  !filterItems?.selectedSortedOptions?.length ? (
                    <span className="px-2 py-1.5 text-sm font-normal">{filterEmpty}</span>
                  ) : null}

                  {filterItems?.selectedSortedOptions.map((option) => {
                    const isSelected = selectedValues.has(option.value);
                    return FilterItem(option, isSelected, handleSelect, column);
                  })}
                  {filterItems !== undefined &&
                    filterItems.selectedSortedOptions.length > 0 &&
                    filterItems.unselectedSortedOptions.length > 0 && (
                      <CommandSeparator className="mt-1 mb-1" />
                    )}
                  {filterItems?.unselectedSortedOptions.map((option) => {
                    const isSelected = selectedValues.has(option.value);
                    return FilterItem(option, isSelected, handleSelect, column);
                  })}
                </CommandGroup>
              </CommandList>
              {filterItems?.unselectedSortedOptions.length === 0 &&
              !filterItems?.selectedSortedOptions?.length ? null : (
                <>
                  <CommandSeparator />
                  <CommandItem>
                    <Button
                      onClick={clearFilters}
                      variant={'ghost'}
                      className={cn(
                        'h-6 rounded-none w-full justify-center text-center capitalize',
                      )}
                    >
                      선택 해제
                    </Button>
                  </CommandItem>
                </>
              )}
            </Command>
          </PopoverContent>
        </Popover>
      </div>
    );
  },
);
FacetedFilter.displayName = 'FacetedFilter';

export type {FacetedFilterProps, FacetedFilterOption, FacetedFilterRefProps};
export {FacetedFilter};
