import * as React from 'react';
import { cn } from '../../../lib';
import { Title } from '../../ui-extension';
import { Tabs, TabsList, TabsTrigger } from '../../ui';

// ─── SheetTemplate Context (header border-b 제어)─────────────────────────────────────────────
interface SheetTemplateContextType {
  NoBorder: boolean;
  setNoBorder: (value: boolean) => void;
}

const SheetTemplateContext = React.createContext<SheetTemplateContextType>({
  NoBorder: false,
  setNoBorder: () => {},
});

const SheetTemplate = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => {
  const [NoBorder, setNoBorder] = React.useState(false);
  return (
    <SheetTemplateContext.Provider value={{ NoBorder, setNoBorder }}>
      <div
        ref={ref}
        className={cn('flex flex-col h-full w-full bg-background', className)}
        {...props}
      />
    </SheetTemplateContext.Provider>
  );
});
SheetTemplate.displayName = 'SheetTemplate';

export interface SheetTemplateHeaderProps extends Omit<React.HTMLAttributes<HTMLDivElement>, 'title'> {
  title?: React.ReactNode;
  description?: React.ReactNode;
  // 왼쪽 버튼 영역 (정보, 상태 표시)
  leftInfoGroup?: React.ReactNode;
  // 우측 액션 버튼 영역 (액션 버튼 영역 (Edit, copy,Close)
  rightButtonGroup?: React.ReactNode;
}

const SheetTemplateHeader = React.forwardRef<HTMLDivElement, SheetTemplateHeaderProps>(
  ({ className, title, description, leftInfoGroup, rightButtonGroup, children, ...props }, ref) => {
    const { NoBorder } = React.useContext(SheetTemplateContext);

    if (children) {
      return (
        <div ref={ref} className={cn('flex items-start justify-between px-5 py-4 shrink-0', !NoBorder && 'border-b', className)} {...props}>
          {children}
        </div>
      );
    }

    return (
      <div ref={ref} className={cn('flex items-center justify-between px-4 py-3 shrink-0 gap-4', !NoBorder && 'border-b', className)} {...props}>
        {/* leftInfoGroup: title + leftInfoGroup + description */}
        <div className="flex flex-col gap-1 min-w-0 shrink">
          <div className="flex items-center gap-2 min-w-0">
            {title && <Title variant="h3" className="pb-0">{title}</Title>}
            {leftInfoGroup && <div className="flex items-center gap-2 shrink-0">{leftInfoGroup}</div>}
          </div>
          {description && <p className="text-xs text-muted-foreground truncate">{description}</p>}
        </div>

        {/* RightButtonGroup: 액션 버튼 영역 (Edit, copy,Close) */}
        {rightButtonGroup && (
          <div className="flex items-center gap-2 shrink-0">
            {rightButtonGroup}
          </div>
        )}
      </div>
    );
  }
);
SheetTemplateHeader.displayName = 'SheetTemplateHeader';

export interface SheetTemplateTabItem {
  label: React.ReactNode;
  value: string;
}

export interface SheetTemplateTabsProps extends React.HTMLAttributes<HTMLDivElement> {
  tabs: SheetTemplateTabItem[];
  activeTab: string;
  onTabChange: (value: string) => void;
}

const SheetTemplateTabs = React.forwardRef<HTMLDivElement, SheetTemplateTabsProps>(
  ({ className, tabs, activeTab, onTabChange, ...props }, ref) => {
    const { setNoBorder } = React.useContext(SheetTemplateContext);

    React.useLayoutEffect(() => {
      setNoBorder(true);
      return () => setNoBorder(false);
    }, [setNoBorder]);

    return (
      <div ref={ref} className={cn('w-full border-b shrink-0', className)} {...props}>
        <Tabs value={activeTab} onValueChange={onTabChange} className="w-full">
          <TabsList className="flex h-auto w-full justify-start gap-2 rounded-none bg-transparent p-0 px-4">
            {tabs.map((tab) => (
              <TabsTrigger
                key={tab.value}
                value={tab.value}
                className={cn(
                  'text-sm relative rounded-none px-2 py-2 font-medium transition-all border-b-2 cursor-pointer -mb-[1px]',
                  'border-transparent text-muted-foreground hover:text-foreground',
                  'data-[state=active]:border-primary data-[state=active]:text-foreground data-[state=active]:shadow-none',
                )}
              >
                {tab.label}
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>
      </div>
    );
  }
);
SheetTemplateTabs.displayName = 'SheetTemplateTabs';

const SheetTemplateContent = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn('flex flex-col flex-1 overflow-y-auto px-4 py-4 gap-6', className)}
    {...props}
  />
));
SheetTemplateContent.displayName = 'SheetTemplateContent';

// ─── SheetTemplateGroup (섹션 컨테이너. title만 담당) ──────────────────────────────────
export interface SheetTemplateGroupProps extends Omit<React.HTMLAttributes<HTMLDivElement>, 'title'> {
  title?: React.ReactNode;
}

const SheetTemplateGroup = React.forwardRef<HTMLDivElement, SheetTemplateGroupProps>(
  ({ className, title, children, ...props }, ref) => {
    return (
      <div ref={ref} className={cn('flex flex-col gap-2', className)} {...props}>
        {title && <Title variant="h4">{title}</Title>}
        {children}
      </div>
    );
  }
);
SheetTemplateGroup.displayName = 'SheetTemplateGroup';

// ─── SheetTemplateGroupGrid (Grid 레이아웃 영역)─────────────────────────────────────────
export interface SheetTemplateGroupGridProps extends React.HTMLAttributes<HTMLDivElement> {
  cols?: 1 | 2;
}

const SheetTemplateGroupGrid = React.forwardRef<HTMLDivElement, SheetTemplateGroupGridProps>(
  ({ className, cols = 2, children, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(
          'grid gap-x-4',
          cols === 1 ? 'grid-cols-1 gap-y-2' : 'grid-cols-2 gap-y-4',
          className,
        )}
        {...props}
      >
        {children}
      </div>
    );
  }
);
SheetTemplateGroupGrid.displayName = 'SheetTemplateGroupGrid';

// ─── SheetTemplateGroupGridItem (내부 단일 항목 label + value + 선택적 description) ────────
export interface SheetTemplateGroupGridItemProps {
  label: React.ReactNode;
  value: React.ReactNode;
  description?: React.ReactNode;
  cols?: 1 | 2;
}

function SheetTemplateGroupGridItem({ label, value, description, cols = 2 }: SheetTemplateGroupGridItemProps) {
  return (
    <div className={cn('flex', cols === 1 ? 'flex-row items-center justify-between gap-3' : 'flex-col gap-0.5')}>
      <span className="text-[13px] font-medium text-muted-foreground shrink-0">{label}</span>
      <div className={cn('text-sm', cols === 1 && 'text-right')}>
        {value}
        {description && (
          <div className="text-xs text-muted-foreground mt-1 whitespace-pre-line wrap-break-word">
            {description}
          </div>
        )}
      </div>
    </div>
  );
}

// ─── SheetTemplateGroupTable (Table 래퍼 영역) ─────────────────────────────────────────────────────
const SheetTemplateGroupTable = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, children, ...props }, ref) => (
  <div
    ref={ref}
    className={cn('w-full overflow-hidden [&_.scrollbar-container]:min-h-0', className)}
    {...props}
  >
    {children}
  </div>
));
SheetTemplateGroupTable.displayName = 'SheetTemplateGroupTable';

// ─── SheetTemplateGroupTab (Tab 래퍼 영역) ─────────────────────────────────────────────────────
export interface SheetTemplateGroupTabProps extends React.HTMLAttributes<HTMLDivElement> {
  tabs: { label: React.ReactNode; value: string }[];
  activeTab?: string;
  onTabChange?: (value: string) => void;
}

const SheetTemplateGroupTab = React.forwardRef<HTMLDivElement, SheetTemplateGroupTabProps>(
  ({ className, tabs, activeTab, onTabChange, children, ...props }, ref) => {
    return (
      <div ref={ref} className={cn('w-full flex flex-col', className)} {...props}>
        <Tabs value={activeTab} onValueChange={onTabChange} className="w-full">
          <TabsList className="w-full h-8">
            {tabs.map((tab) => (
              <TabsTrigger key={tab.value} value={tab.value} className="w-full h-6 text-[13px]">
                {tab.label}
              </TabsTrigger>
            ))}
          </TabsList>
          {children}
        </Tabs>
      </div>
    );
  }
);
SheetTemplateGroupTab.displayName = 'SheetTemplateGroupTab';

// ─── SheetTemplateGroupChart (Chart 래퍼 영역) ─────────────────────────────────────────────────────
const SheetTemplateGroupChart = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, children, ...props }, ref) => (
  <div ref={ref} className={cn('w-full flex flex-col gap-4', className)} {...props}>
    {children}
  </div>
));
SheetTemplateGroupChart.displayName = 'SheetTemplateGroupChart';

// ─── SheetTemplateFooter ────────────────────────────────────────────────────────
const SheetTemplateFooter = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, children, ...props }, ref) => (
  <div
    ref={ref}
    className={cn('flex items-center justify-start gap-2 px-4 py-3 border-t shrink-0', className)}
    {...props}
  >
    {children}
  </div>
));
SheetTemplateFooter.displayName = 'SheetTemplateFooter';

export {
  SheetTemplate,
  SheetTemplateHeader,
  SheetTemplateContent,
  SheetTemplateTabs,
  SheetTemplateGroup,
  SheetTemplateGroupGrid,
  SheetTemplateGroupGridItem,
  SheetTemplateGroupTable,
  SheetTemplateGroupTab,
  SheetTemplateGroupChart,
  SheetTemplateFooter,
};
