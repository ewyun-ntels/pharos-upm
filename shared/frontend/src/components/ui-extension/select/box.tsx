import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@pharos/shared/components/ui';
import React, {ReactNode, useEffect, useState} from 'react';
import {cn} from '../../../lib';

export interface SelectOption {
  label: string;
  value: string;
  icon?: React.ReactNode;
  [key: string]: any;
}

export interface SelectBoxProps {
  label?: string | ReactNode;
  options: SelectOption[] | string[];
  labelKey?: string;
  valueKey?: string;
  value?: string;
  onChange: (value: string) => void;
  autoSelectFirstOption?: boolean;
  placeholder?: string;
  size?: 'small' | 'medium' | 'large' | 'full' | 'hsmall' | 'none';
  className?: string;
  name?: string;
  showAllOption?: boolean;
}

const defaultLabelKey = 'label';
const defaultValueKey = 'value';

const sizeClass = {
  small: 'w-40',
  medium: 'w-60',
  large: 'w-80',
  full: 'w-full',
  hsmall: 'h-8',
  none: '',
};

export const SelectBox: React.FC<SelectBoxProps> = React.memo(
  ({
    label,
    options,
    labelKey = defaultLabelKey,
    valueKey = defaultValueKey,
    value,
    onChange,
    autoSelectFirstOption = true,
    size = 'none',
    placeholder = '',
    className,
    name,
    showAllOption = false,
  }: SelectBoxProps) => {
    const [selectedValue, setSelectedValue] = useState(value || '');
    const [processedOptions, setProcessedOptions] = useState<(SelectOption | string)[]>([]);

    // Process options (including showAllOption)
    useEffect(() => {
      let nextOptions = [...options];
      if (showAllOption) {
        const hasAll = nextOptions.some((opt) =>
          typeof opt === 'string' ? opt === 'All' : opt.value === 'All'
        );
        if (!hasAll) {
          nextOptions = [{label: 'All', value: 'All'}, ...nextOptions];
        }
      }
      setProcessedOptions(nextOptions);
    }, [options, showAllOption]);

    // Auto-select first option on mount or when options change
    useEffect(() => {
      if (!value && autoSelectFirstOption && processedOptions.length > 0) {
        const firstOption = processedOptions[0];
        const firstValue = typeof firstOption === 'string' ? firstOption : firstOption.value;
        if (selectedValue !== firstValue) {
          setSelectedValue(firstValue);
          onChange(firstValue);
        }
      }
      // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [processedOptions]);

    // Sync external value changes
    useEffect(() => {
      if (value !== undefined && value !== selectedValue) {
        setSelectedValue(value);
      }
      // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [value]);

    const handleValueChange = (newValue: string) => {
      setSelectedValue(newValue);
      onChange(newValue);
    };

    const renderOptions = () => {
      if (!processedOptions || processedOptions.length === 0) {
        return null;
      }

      return processedOptions.map((option, index) => {
        if (typeof option === 'string') {
          return (
            <SelectItem value={option} key={`${option}-${index}`}>
              {option}
            </SelectItem>
          );
        }

        const optionValue = option[valueKey] || option.value;
        const optionLabel = option[labelKey] || option.label;

        return (
          <SelectItem value={optionValue} key={`${optionValue}-${index}`}>
            <div className="flex items-center gap-2">
              {option.icon && (
                <span className="inline-flex items-center justify-center shrink-0 size-8 [&_svg]:size-8!">
                  {option.icon}
                </span>
              )}
              <span>{optionLabel}</span>
            </div>
          </SelectItem>
        );
      });
    };

    return (
      <div className={cn('flex items-center flex-1')}>
        <Select
          onValueChange={handleValueChange}
          value={selectedValue}
          defaultValue={selectedValue}
          name={name}
        >
          <SelectTrigger
            className={cn(
              'w-fit bg-background flex gap-2 justify-between [&_svg]:size-5 shadow-none h-8',
              sizeClass[size],
              className,
            )}
          >
            {label && (
              <span className="flex items-center text-[13px]">
                {label}
                <span
                  className={
                    typeof label === 'string' ? 'mx-2 h-4 border-r border-border' : 'hidden'
                  }
                />
              </span>
            )}
            {!label && !selectedValue && placeholder && <span>{placeholder}</span>}
            <div className="whitespace-nowrap">
              <SelectValue />
            </div>
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>{renderOptions()}</SelectGroup>
          </SelectContent>
        </Select>
      </div>
    );
  },
);

SelectBox.displayName = 'SelectBox';
