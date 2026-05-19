import React, {useState, useEffect, useRef} from 'react';
import {Search, X} from 'lucide-react';
import {Input} from '../ui';
import {Table} from '@tanstack/react-table';
import {IconButton} from './icon-button';
import {cn} from '../../lib';

interface SearchInputOption {
  setValue?: (value: string) => void;
  defaultSearchValues?: string[];
  searchValues?: string;
  searchValue?: string;
  handleKeyUp?: React.KeyboardEventHandler<HTMLInputElement>;
  placeholder?: string;
  showClearButton?: boolean;
  onSearchChange?: (searchTerm: string) => void;
  size?: 'hsmall' | 'default';
}

interface SearchInputProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'size'>, SearchInputOption {
  table?: Table<any>;
  size?: 'hsmall' | 'default';
}

const SearchInput = React.forwardRef<HTMLInputElement, SearchInputProps>(
  (
    {
      table,
      placeholder,
      defaultSearchValues,
      searchValues,
      searchValue,
      setValue,
      handleKeyUp,
      showClearButton = true,
      onSearchChange,
      size = 'default',
      ...props
    },
    ref,
  ) => {
    const [localSearchValue, setLocalSearchValue] = useState<string>(searchValue || '');
    const isComposingRef = useRef(false);

    // 외부에서 searchValue가 변경될 때 로컬 상태 업데이트
    // IME 조합 중에는 업데이트 생략 (한글 등 입력 중단 방지)
    useEffect(() => {
      if (searchValue !== undefined && !isComposingRef.current) {
        setLocalSearchValue(searchValue);
        table && table.setGlobalFilter(searchValue);
      }
    }, [searchValue, table, defaultSearchValues]);

    const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
      const targetValue = event.target.value;
      setLocalSearchValue(targetValue);

      table && table.setGlobalFilter(targetValue);
      setValue && setValue(targetValue);
      onSearchChange && onSearchChange(targetValue);
    };

    const handleClear = () => {
      setLocalSearchValue('');
      table && table.setGlobalFilter('');
      setValue && setValue('');
      onSearchChange && onSearchChange('');
    };

    const handleKeyDown: React.KeyboardEventHandler<HTMLInputElement> = (event) => {
      if (event.key === 'Enter') {
        table && table.setGlobalFilter(localSearchValue);
        setValue && setValue(localSearchValue);
        onSearchChange && onSearchChange(localSearchValue);
      } else if (event.key === 'Escape') {
        handleClear();
      }

      // 부모에서 전달된 handleKeyUp이 있을 경우 호출
      if (handleKeyUp) {
        handleKeyUp(event);
      }
    };

    // searchHeight : filter tpye은 h-8, editor tpye은 h-9
    const heightClass = size === 'hsmall' ? 'h-8' : 'h-9';
    const iconTopClass = size === 'hsmall' ? 'top-2' : 'top-2.5';

    return (
      <div className={cn('search-input relative', heightClass)}>
        <Search className={cn('absolute left-2.5 w-4 h-4 text-muted-foreground', iconTopClass)} />
        <Input
          autoComplete="off"
          className={cn('w-full sm:w-64 pl-8 bg-background', heightClass)}
          value={localSearchValue || searchValues || ''}
          placeholder={placeholder ?? 'Search'}
          ref={ref}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown} // handleKeyDown을 기본으로하고, 만약 부모에서 전달된 게 있으면 handleKeyUp으로 되게 함
          onCompositionStart={() => {
            isComposingRef.current = true;
          }}
          onCompositionEnd={() => {
            isComposingRef.current = false;
          }}
          {...props}
        />
        {showClearButton && localSearchValue && (
          <IconButton
            onClick={handleClear}
            variant="ghost"
            icon={<X />}
            className={cn(
              'absolute right-2.5 p-0 rounded-full w-5 h-5 top-1/2 -translate-y-1/2 transform z-100',
            )}
          >
            초기화
          </IconButton>
        )}
      </div>
    );
  },
);

SearchInput.displayName = 'SearchInput';

export {SearchInput};
