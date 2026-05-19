'use client';

import React, {useState, createContext, useContext, useMemo, useCallback} from 'react';
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '../../../ui';
import {cn} from '../../../../lib';
import type {TableSelectContextType, OptionData} from './types';
import {Title} from '../../../ui-extension';

const TableSelectContext = createContext<TableSelectContextType | null>(null);

interface TableSelectTemplateProps {
  title?: string;
  description?: string;
  children: React.ReactNode;
  defaultValue?: string;
  className?: string;
  onSelectionChange?: (value: string) => void;
  selectPlaceholder?: string;
  selectLabel?: string;
  selectWidth?: string;
  leftCustomItems?: React.ReactNode;
  rightCustomItems?: React.ReactNode;
}

/**
 * TableSelectTemplate
 * 
 * Select 드롭다운으로 옵션을 선택하고, 선택된 옵션에 따라 다른 컨텐츠를 표시하는 템플릿
 * 
 * @example
 * ```tsx
 * <TableSelectTemplate
 *   title="Alert Editor"
 *   description="Configure alert rules and notifications"
 *   selectLabel="Editor Type"
 *   selectPlaceholder="Select editor type"
 *   defaultValue="rule"
 *   onSelectionChange={(value) => console.log(value)}
 * >
 *   <Options>
 *     <Option value="rule" label="Alert Rules">
 *       <AlertRuleEditor />
 *     </Option>
 *     <Option value="notification" label="Notifications">
 *       <NotificationEditor />
 *     </Option>
 *   </Options>
 * </TableSelectTemplate>
 * ```
 */
export function TableSelectTemplate({
  title,
  description,
  children,
  defaultValue,
  className,
  onSelectionChange,
  selectPlaceholder = 'Select an option',
  selectLabel,
  selectWidth = '200px',
  leftCustomItems,
  rightCustomItems,
}: TableSelectTemplateProps) {
  const [optionsData, setOptionsDataState] = useState<Record<string, OptionData>>({});
  const [selectedValue, setSelectedValueState] = useState<string>('');
  const [disabledOptions, setDisabledOptionsState] = useState<Set<string>>(new Set());

  const optionValues = useMemo(() => Object.keys(optionsData), [optionsData]);

  // onSelectionChange를 ref로 관리
  const onSelectionChangeRef = React.useRef(onSelectionChange);
  
  React.useEffect(() => {
    onSelectionChangeRef.current = onSelectionChange;
  }, [onSelectionChange]);

  // 초기값 설정 (마운트 시 한 번만)
  const isInitializedRef = React.useRef(false);
  
  React.useEffect(() => {
    // 이미 초기화되었으면 무시
    if (isInitializedRef.current) {
      return;
    }

    // 옵션이 로드되고, 아직 선택값이 없으면 초기화
    if (optionValues.length > 0 && !selectedValue) {
      const initialValue = defaultValue || optionValues[0];
      setSelectedValueState(initialValue);
      onSelectionChangeRef.current?.(initialValue);
      isInitializedRef.current = true;
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [optionValues.length, selectedValue, defaultValue]);

  const setSelectedValue = useCallback((value: string) => {
    if (!disabledOptions.has(value)) {
      setSelectedValueState(value);
      onSelectionChangeRef.current?.(value);
    }
  }, [disabledOptions]);

  const setOptionData = useCallback((value: string, label: string, data: React.ReactNode) => {
    setOptionsDataState((prev) => {
      const current = prev[value];
      if (current && current.label === label && current.content === data) {
        return prev;
      }
      return {
        ...prev,
        [value]: {label, content: data},
      };
    });
  }, []);

  const setOptionDisabled = useCallback((value: string, disabled: boolean) => {
    setDisabledOptionsState((prev) => {
      const newSet = new Set(prev);
      if (disabled) {
        newSet.add(value);
      } else {
        newSet.delete(value);
      }
      return newSet;
    });
  }, []);

  const contextValue: TableSelectContextType = useMemo(
    () => ({
      selectedValue,
      setSelectedValue,
      optionsData,
      setOptionData,
      disabledOptions,
      setOptionDisabled,
    }),
    [selectedValue, setSelectedValue, optionsData, setOptionData, disabledOptions, setOptionDisabled]
  );

  return (
    <TableSelectContext.Provider value={contextValue}>
      <div className={cn('w-full', className)}>
        <section className="w-full flex flex-col">
          {/* Header Section */}
          {(title || optionValues.length > 0 || leftCustomItems || rightCustomItems) && (
            <header className="flex flex-col w-full gap-3 pb-2">
              {/* Title and Description */}
              {(title || description) && (
                <div className="flex flex-col gap-1">
                  {title && <Title>{title}</Title>}
                  {description && <p className="text-sm text-muted-foreground">{description}</p>}
                </div>
              )}

              {/* Controls Row: Left Items + Select + Right Items */}
              <div className="flex items-center justify-between w-full gap-4">
                {/* Left Custom Items + Select (왼쪽 정렬) */}
                <div className="flex items-center gap-2">
                  {leftCustomItems}

                  {optionValues.length > 0 && (
                    <Select key={selectedValue} value={selectedValue} onValueChange={setSelectedValue}>
                      <SelectTrigger style={{width: selectWidth}} className="h-8">
                        <SelectValue placeholder={selectPlaceholder} />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          {selectLabel && <SelectLabel>{selectLabel}</SelectLabel>}
                          {optionValues.map((value) => {
                            const isDisabled = disabledOptions.has(value);
                            const optionData = optionsData[value];
                            const displayLabel = optionData?.label || value;
                            return (
                              <SelectItem key={value} value={value} disabled={isDisabled}>
                                {displayLabel}
                              </SelectItem>
                            );
                          })}
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  )}
                </div>

                {/* Right Custom Items (오른쪽 정렬) */}
                {rightCustomItems && <div className="flex items-center gap-2">{rightCustomItems}</div>}
              </div>
            </header>
          )}

          {/* Body: Render selected option content */}
          <div className="w-full">
            {/* Register children (Options component) - hidden */}
            <div style={{display: 'none'}}>{children}</div>

            {/* Render selected content */}
            {selectedValue && optionsData[selectedValue] && optionsData[selectedValue].content}
          </div>
        </section>
      </div>
    </TableSelectContext.Provider>
  );
}

interface OptionsProps {
  children: React.ReactNode;
}

/**
 * Options wrapper component
 * TableSelectTemplate의 자식으로 사용
 */
export function Options({children}: OptionsProps) {
  return <>{children}</>;
}

interface OptionProps {
  value: string;
  label: string;
  children: React.ReactNode;
  disabled?: boolean;
}

/**
 * Option component
 * Select의 각 옵션과 해당 옵션 선택 시 렌더링될 컨텐츠를 정의
 * 
 * @param value - Select의 option value (고유 식별자)
 * @param label - Select에 표시될 레이블
 * @param children - 해당 옵션이 선택되었을 때 렌더링될 컨텐츠
 * @param disabled - 옵션 비활성화 여부
 */
export function Option({value, label, children, disabled = false}: OptionProps) {
  const context = useContext(TableSelectContext);

  React.useEffect(() => {
    if (context) {
      const current = context.optionsData[value];
      if (!current || current.label !== label || current.content !== children) {
        // Only update if data actually changed
        context.setOptionData(value, label, children);
        context.setOptionDisabled(value, disabled);
      }
    }
  }, [value, label, context, children, disabled]);

  return null;
}

/**
 * Hook to access TableSelect context
 * 자식 컴포넌트에서 현재 선택된 값 등에 접근 가능
 */
export function useTableSelectContext() {
  const context = useContext(TableSelectContext);
  if (!context) {
    throw new Error('useTableSelectContext must be used within TableSelectTemplate');
  }
  return context;
}
