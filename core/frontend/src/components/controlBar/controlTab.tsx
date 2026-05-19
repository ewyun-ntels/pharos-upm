import React from 'react';
import {Button} from '@pharos/shared/components/ui';
import {Menu, Pen, File, XIcon, Terminal} from '@pharos/shared/components';
import {TabType} from '@components/controlBar/controlTabStore';
import {cn} from '@lib/utils';

interface TabsProps {
  selected: boolean;
  onClick: React.MouseEventHandler<HTMLButtonElement>;
  onCancelClick: React.MouseEventHandler<HTMLSpanElement>;
  tabId: string;
  label?: string;
  // 탭 타입 정의
  tabType: TabType;
  // 탭 라인 숨김 여부
  tabLineHide?: boolean;
}

const Tabs = React.memo(
  ({selected, onClick, onCancelClick, tabId, label, tabType, tabLineHide}: TabsProps) => {
    const getIcon = () => {
      switch (tabType) {
        case 'editor':
          return <File className="h-4 w-4" />;
        case 'modifyEditor':
          return <Pen className="h-4 w-4" />;
        case 'terminal':
          return <Terminal className="h-4 w-4" />;
        case 'logView':
          return <Menu className="h-4 w-4" />;
        default:
          return null;
      }
    };

    return (
      <Button
        variant={selected ? 'secondary' : 'ghost'}
        className={cn(
          'cursor-pointer rounded-none border-r shadow-none pr-4 gap-2 items-start align-middle [&>svg]:mt-0.5 [&:last-child]:border-r-0',
          selected && tabLineHide
            ? 'bg-primary/10 hover:bg-primary/10 dark:hover:bg-primary/40 items-start border-r-0'
            : selected
              ? 'bg-primary/20 hover:bg-primary/20 dark:hover:bg-primary/40 items-start border-b-primary border-r-0 border-b-2'
              : 'dark:hover:bg-primary/40 text-secondary-foreground/50',
          // !tabLineHide && "border-b-2"// 디자인이 선택되었을때만 border가 필요해서 주석처리함
        )}
        onClick={onClick}
        value={tabId}
        key={tabId}
      >
        {getIcon()} {/* 탭타입별 아이콘 추가 */}
        {label ? label : tabId}
        <span
          onClick={(e) => {
            e.stopPropagation(); // button과 X icon 이벤트가 동시에 발생하지 않도록 하기 위해
            onCancelClick(e);
          }}
          className={'cursor-pointer h-4 w-4 hover:bg-background/40 rounded-sm'}
        >
          <XIcon />
        </span>
      </Button>
    );
  },
);
Tabs.displayName = 'Tabs';

export {Tabs};
