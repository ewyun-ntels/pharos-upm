'use client';

import React, {useState, createContext, useContext, useMemo, useCallback} from 'react';
import {Title} from '@pharos/shared/components/ui-extension';

import {cn} from '../../../../lib';
import {X} from 'lucide-react';

interface TableTabsContextType {
  activeTab: string;
  setActiveTab: (tabName: string) => void;
  tabsData: Record<string, React.ReactNode>;
  setTabData: (tabName: string, data: React.ReactNode) => void;
  tabsLabel: Record<string, React.ReactNode>;
  setTabLabel: (tabName: string, label: React.ReactNode) => void;
  removeTab: (tabName: string) => void;
  unregisterTab: (tabName: string) => void;
  disabledTabs: Set<string>;
  setTabDisabled: (tabName: string, disabled: boolean) => void;
  closableTabs: Set<string>;
  setTabClosable: (tabName: string, closable: boolean) => void;
}

const TableTabsContext = createContext<TableTabsContextType | null>(null);

interface TableTabsTemplateProps {
  name?: string;
  description?: string;
  children: React.ReactNode;
  defaultTab?: string;
  className?: string;
  onTabChange?: (tabName: string) => void;
  onTabClose?: (tabName: string) => void;
  showTabs?: boolean;
  activeTab?: string;
  breadcrumb?: React.ReactNode;
  rightFilters?: () => React.ReactNode[];
  /** 탭 전환 시 비활성 탭 컴포넌트를 언마운트하지 않고 유지합니다 (display:none). state가 보존됩니다. */
  keepMounted?: boolean;
}

export function TableTabsTemplate({
  name,
  description,
  children,
  defaultTab,
  className,
  onTabChange,
  onTabClose,
  showTabs = true,
  activeTab: externalActiveTab,
  breadcrumb,
  rightFilters,
  keepMounted = false,
}: TableTabsTemplateProps) {
  const [tabsData, setTabsDataState] = useState<Record<string, React.ReactNode>>({});
  const [tabsLabel, setTabsLabelState] = useState<Record<string, React.ReactNode>>({});
  const [activeTab, setActiveTabState] = useState<string>('');
  const [disabledTabs, setDisabledTabsState] = useState<Set<string>>(new Set());
  const [closableTabs, setClosableTabsState] = useState<Set<string>>(new Set());

  const tabNames = Object.keys(tabsData);

  // Sync with external activeTab if provided
  React.useEffect(() => {
    if (externalActiveTab && externalActiveTab !== activeTab) {
      setActiveTabState(externalActiveTab);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [externalActiveTab]);

  React.useEffect(() => {
    if (tabNames.length > 0 && !activeTab) {
      const firstTab = defaultTab || tabNames[0];
      setActiveTabState(firstTab);
    }
  }, [tabNames, activeTab, defaultTab]);

  const setActiveTab = useCallback(
    (tabName: string) => {
      if (!disabledTabs.has(tabName)) {
        setActiveTabState(tabName);
        onTabChange?.(tabName);
      }
    },
    [onTabChange, disabledTabs],
  );

  const setTabData = useCallback((tabName: string, data: React.ReactNode) => {
    setTabsDataState((prev) => {
      if (prev[tabName] === data) return prev;
      return {
        ...prev,
        [tabName]: data,
      };
    });
  }, []);

  const setTabLabel = useCallback((tabName: string, label: React.ReactNode) => {
    setTabsLabelState((prev) => {
      if (prev[tabName] === label) return prev;
      return {
        ...prev,
        [tabName]: label,
      };
    });
  }, []);

  const removeTab = useCallback(
    (tabName: string) => {
      onTabClose?.(tabName);
      
      setTabsDataState((prev) => {
        const newState = {...prev};
        delete newState[tabName];
        return newState;
      });
      setTabsLabelState((prev) => {
        const newState = {...prev};
        delete newState[tabName];
        return newState;
      });

      if (activeTab === tabName) {
        const remainingTabs = tabNames.filter((name) => name !== tabName);
        if (remainingTabs.length > 0) {
          const nextTab = remainingTabs[0];
          setActiveTab(nextTab);
        } else {
          setActiveTabState('');
        }
      }
    },
    [activeTab, tabNames, onTabClose, setActiveTab],
  );

  // Tab 컴포넌트 언마운트 시 onTabClose 호출 없이 내부 상태만 정리
  const unregisterTab = useCallback((tabName: string) => {
    setTabsDataState((prev) => {
      if (!(tabName in prev)) return prev;
      const newState = {...prev};
      delete newState[tabName];
      return newState;
    });
    setTabsLabelState((prev) => {
      if (!(tabName in prev)) return prev;
      const newState = {...prev};
      delete newState[tabName];
      return newState;
    });
    setDisabledTabsState((prev) => {
      if (!prev.has(tabName)) return prev;
      const newSet = new Set(prev);
      newSet.delete(tabName);
      return newSet;
    });
    setClosableTabsState((prev) => {
      if (!prev.has(tabName)) return prev;
      const newSet = new Set(prev);
      newSet.delete(tabName);
      return newSet;
    });
  }, []);

  const setTabDisabled = useCallback((tabName: string, disabled: boolean) => {
    setDisabledTabsState((prev) => {
      if (prev.has(tabName) === disabled) return prev;
      const newSet = new Set(prev);
      if (disabled) {
        newSet.add(tabName);
      } else {
        newSet.delete(tabName);
      }
      return newSet;
    });
  }, []);

  const setTabClosable = useCallback((tabName: string, closable: boolean) => {
    setClosableTabsState((prev) => {
      if (prev.has(tabName) === closable) return prev;
      const newSet = new Set(prev);
      if (closable) {
        newSet.add(tabName);
      } else {
        newSet.delete(tabName);
      }
      return newSet;
    });
  }, []);

  const contextValue: TableTabsContextType = useMemo(
    () => ({
      activeTab,
      setActiveTab,
      tabsData,
      setTabData,
      tabsLabel,
      setTabLabel,
      removeTab,
      unregisterTab,
      disabledTabs,
      setTabDisabled,
      closableTabs,
      setTabClosable,
    }),
    [
      activeTab,
      setActiveTab,
      tabsData,
      setTabData,
      tabsLabel,
      setTabLabel,
      removeTab,
      unregisterTab,
      disabledTabs,
      setTabDisabled,
      closableTabs,
      setTabClosable,
    ],
  );

  return (
    <TableTabsContext.Provider value={contextValue}>
      <div className={cn('w-full h-full', className)}>
        <section className="w-full h-full flex flex-col">
          {(tabNames.length > 0 || name) && (
            <header className="flex flex-col items-start w-full gap-1 pt-3 pb-2">
              <div className="w-full flex justify-between items-end px-5">
                <div className="flex flex-col gap-2">
                  {breadcrumb && <div className="mb-1">{breadcrumb}</div>}
                  {name && <Title>{name}</Title>}
                  {description && <p className="text-xs">{description}</p>}
                </div>
                {rightFilters && (
                  <div className="flex items-center gap-2">
                    {rightFilters().map((item, index) => (
                      <React.Fragment key={`right-${index}`}>{item}</React.Fragment>
                    ))}
                  </div>
                )}
              </div>

              {/* Tab Navigation - 왼쪽 정렬, 하단 라인 스타일 */}
              {showTabs && tabNames.length > 0 && (
                <div className="w-full mt-2 border-b border-border">
                  <div className="flex px-5 gap-2">
                    {tabNames.map((tabName) => {
                      const isDisabled = disabledTabs.has(tabName);
                      const isClosable = closableTabs.has(tabName);
                      const isActive = activeTab === tabName;

                      return (
                        <div
                          key={tabName}
                          className={cn(
                            'text-sm relative rounded-none px-2 py-2 font-medium transition-all border-b-2 flex items-center gap-2 group cursor-pointer',
                            isActive
                              ? 'border-primary'
                              : 'border-transparent text-muted-foreground hover:text-foreground',
                            isDisabled && 'opacity-50 cursor-not-allowed hover:bg-transparent',
                          )}
                          onClick={() => !isDisabled && setActiveTab(tabName)}
                        >
                          <span>{tabsLabel[tabName] || tabName}</span>
                          {onTabClose && isClosable && (
                            <div
                              className={cn(
                                'flex items-center justify-center rounded-sm hover:bg-muted p-0.5 transition-all opacity-0 group-hover:opacity-100',
                                isActive && 'opacity-70',
                              )}
                              onClick={(e) => {
                                e.stopPropagation();
                                removeTab(tabName);
                              }}
                            >
                              <X className="h-3.5 w-3.5 hover:text-destructive" />
                            </div>
                          )}
                        </div>
                      );
                    })}
                  </div>
                </div>
              )}
            </header>
          )}

          <div className="flex-1 min-h-0 flex flex-col">
            {/* Register children (Tabs component) */}
            <div style={{display: 'none'}}>{children}</div>

            {/* Render active tab content */}
            {keepMounted
              ? Object.entries(tabsData).map(([tabName, content]) => (
                  <div
                    key={tabName}
                    className={cn('flex-1 min-h-0 overflow-y-auto', tabName !== activeTab && 'hidden')}
                  >
                    {content}
                  </div>
                ))
              : activeTab && <div className="flex-1 min-h-0 overflow-y-auto">{tabsData[activeTab]}</div>}
          </div>
        </section>
      </div>
    </TableTabsContext.Provider>
  );
}

interface TabsProps {
  children: React.ReactNode;
}

export function Tabs({children}: TabsProps) {
  return <>{children}</>;
}

interface TabProps {
  name: string;
  label?: React.ReactNode;
  children: React.ReactNode;
  disabled?: boolean;
  closable?: boolean;
}

export function Tab({name, label, children, disabled = false, closable = false}: TabProps) {
  const context = useContext(TableTabsContext);

  React.useEffect(() => {
    if (context) {
      context.setTabData(name, children);
      context.setTabLabel(name, label || name);
      context.setTabDisabled(name, disabled);
      context.setTabClosable(name, closable);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [name, label, disabled, closable, children]);

  // 언마운트 시 tabsData에서 자신을 제거 (onTabClose 콜백 없이 내부 상태만 정리)
  const nameRef = React.useRef(name);
  nameRef.current = name;
  React.useEffect(() => {
    return () => {
      context?.unregisterTab(nameRef.current);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return null;
}
