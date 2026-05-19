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
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '../../../ui';
import {cn} from '../../../../lib';
import type {EditorSelectContextType, OptionData} from './types';

const EditorSelectContext = createContext<EditorSelectContextType | null>(null);

interface EditorSelectTemplateProps {
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
  cardClassName?: string;
  headerClassName?: string;
  contentClassName?: string;
}

/**
 * EditorSelectTemplate
 * 
 * Card 통합형 에디터 템플릿 - Select 드롭다운으로 옵션을 선택하고, 선택된 옵션에 따라 다른 컨텐츠를 표시
 * 업계 표준 패턴(Material-UI, Ant Design, Shadcn/ui)을 따라 Select를 Card 내부(Header)에 배치
 * 
 * @example
 * ```tsx
 * <EditorSelectTemplate
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
 * </EditorSelectTemplate>
 * ```
 */
export function EditorSelectTemplate({
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
}: EditorSelectTemplateProps) {
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

  const contextValue: EditorSelectContextType = useMemo(
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
    <EditorSelectContext.Provider value={contextValue}>
      {/* Title */}
      {title && <CardTitle className="mb-3">{title}</CardTitle>}

      {/* Description */}
      {description && (
        <CardDescription className="mt-1.5">{description}</CardDescription>
      )}

      {(optionValues.length > 0 || leftCustomItems || rightCustomItems) && (
        <div className="flex items-center gap-2">
          {leftCustomItems}

          {optionValues.length > 0 && (
            <Select key={selectedValue} value={selectedValue} onValueChange={setSelectedValue}>
              <SelectTrigger style={{width: selectWidth}} className="h-9">
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

          {rightCustomItems}
        </div>
      )}
      <div style={{display: 'none'}}>{children}</div>

      {/* Render selected content */}
      {selectedValue && optionsData[selectedValue] && optionsData[selectedValue].content}
    </EditorSelectContext.Provider>
  );
}

interface OptionsProps {
  children: React.ReactNode;
}

/**
 * Options wrapper component
 * EditorSelectTemplate의 자식으로 사용
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
  const context = useContext(EditorSelectContext);

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
 * Hook to access EditorSelect context
 * 자식 컴포넌트에서 현재 선택된 값 등에 접근 가능
 */
export function useEditorSelectContext() {
  const context = useContext(EditorSelectContext);
  if (!context) {
    throw new Error('useEditorSelectContext must be used within EditorSelectTemplate');
  }
  return context;
}
