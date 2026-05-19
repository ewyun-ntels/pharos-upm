import React, {useState, useEffect, useRef} from 'react';
import {X} from 'lucide-react';
import {cn} from '../../lib';

interface VariableInputProps {
  label?: string;
  value?: string;
  onChange?: (value: string) => void;
  placeholder?: string;
  disabled?: boolean;
  className?: string;
}

const VariableInput = React.forwardRef<HTMLDivElement, VariableInputProps>(
  ({label, value, onChange, placeholder, disabled = false, className}, ref) => {
    const [localValue, setLocalValue] = useState(value ?? '');
    const inputRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
      if (value !== undefined) setLocalValue(value);
    }, [value]);

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      const v = e.target.value;
      setLocalValue(v);
      onChange?.(v);
    };

    const handleClear = (e: React.MouseEvent) => {
      e.stopPropagation();
      setLocalValue('');
      onChange?.('');
      inputRef.current?.focus();
    };

    return (
      <div
        ref={ref}
        className={cn(
          'flex items-center h-8',
          'rounded-md border border-input bg-background',
          'hover:border-input focus-within:ring-1 focus-within:ring-ring',
          disabled && 'opacity-50 cursor-not-allowed',
          className,
        )}
      >
        {label && (
          <div className="flex items-center text-[13px] font-semibold gap-1.5 px-3 border-r border-border shrink-0">
            {label}
          </div>
        )}
        <div className="relative flex items-center flex-1 min-w-0">
          <input
            ref={inputRef}
            type="text"
            value={localValue}
            onChange={handleChange}
            placeholder={placeholder}
            disabled={disabled}
            className={cn(
              'w-full px-3 py-1 text-sm bg-transparent',
              'focus:outline-none focus-visible:ring-0',
              'disabled:cursor-not-allowed',
              localValue ? 'pr-7' : '',
            )}
          />
          {localValue && (
            <button
              type="button"
              onClick={handleClear}
              className="absolute right-2 flex items-center justify-center w-4 h-4 text-muted-foreground hover:text-foreground"
              tabIndex={-1}
            >
              <X className="w-3 h-3" />
            </button>
          )}
        </div>
      </div>
    );
  },
);

VariableInput.displayName = 'VariableInput';

export type {VariableInputProps};
export {VariableInput};
