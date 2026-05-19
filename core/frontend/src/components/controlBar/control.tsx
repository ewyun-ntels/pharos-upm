import React, {useEffect, useRef} from 'react';
import {IconButton} from '@pharos/shared/components/ui-extension';
import {ChevronDown, ChevronUp, MaximizeIcon, MinimizeIcon, Plus} from '@pharos/shared/components';
import {useWindowDimensions} from '@lib/interanl/window-demensions';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@pharos/shared/components/ui';
import {CreateEvent, useCreateEventStore} from './controlEventStore';
import {TabContent} from '@components/controlBar/controlTabContent';
import {
  getLabel,
  getLastId,
  getNextTabId,
  getNextTypeId,
  useTabStore,
} from '@components/controlBar/controlTabStore';
import {TabsSection} from './controlTabSection';
import {defaultEditorSize, minHeight} from '../layout/constants';

interface ControlProps {
  height: number;
  onResize: (ySize: number, enable: boolean) => void;
}

const Control = ({height, onResize}: ControlProps) => {
  const dimensions = useWindowDimensions();
  const tabStore = useTabStore();
  const scrollRef = useRef<HTMLDivElement>(null);

  const checkIsOpen = (): boolean => height <= minHeight;

  const useEventStore = useCreateEventStore();

  const createHandler = (rcvEvent?: CreateEvent) => {
    const {lastId} = getLastId(tabStore.tabData);

    if (checkIsOpen()) {
      onResize(defaultEditorSize, true);
    }

    let id = 'unknown';
    let lastTypeId: number = 0;
    let label = '';
    switch (rcvEvent?.type) {
      case 'editor':
      case 'modifyEditor':
        id = getNextTabId(lastId);
        lastTypeId = getNextTypeId(rcvEvent.type, tabStore.tabData);
        label = rcvEvent.label ? rcvEvent.label : getLabel(rcvEvent.type, lastTypeId);
        tabStore.setEditor(id, rcvEvent.type, lastTypeId, label, 0, 0, rcvEvent.data);
        break;
      case 'terminal':
        id = getNextTabId(lastId);
        lastTypeId = getNextTypeId(rcvEvent.type, tabStore.tabData);
        label = rcvEvent.label ? rcvEvent.label : getLabel(rcvEvent.type, lastTypeId);
        tabStore.setTerminal(
          id,
          lastTypeId,
          label,
          rcvEvent.namespace,
          rcvEvent.pod,
          rcvEvent.selectContainer,
        );
        break;
      case 'logView':
        id = getNextTabId(lastId);
        lastTypeId = getNextTypeId(rcvEvent.type, tabStore.tabData);
        label = rcvEvent.label ? rcvEvent.label : getLabel(rcvEvent.type, lastTypeId);
        tabStore.setLogView(
          id,
          lastTypeId,
          label,
          rcvEvent.namespace,
          rcvEvent.pod,
          rcvEvent.selectContainer,
          rcvEvent.containers,
        );
        break;
    }

    tabStore.setLastSelectId(id);
  };

  useEffect(() => {
    tabStore.tabData.size === 0 && onResize(0, false);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    const rcvData = useEventStore.eventData.pop();
    rcvData && createHandler(rcvData);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [useEventStore.eventData.length]);

  return (
    <div className={'cli-editor h-full'}>
      <header
        className={
          'flex justify-between bg-background border-t border-b px-6 text-secondary-foreground/70 dark:text-foreground/90'
        }
      >
        <TabsSection
          hideTabLine={height <= minHeight}
          onResize={onResize}
          checkIsOpen={checkIsOpen}
          scrollRef={scrollRef}
        />
        <div
          style={{height: 36}}
          className="flex [&_button:hover]:bg-transparent items-center gap-1"
        >
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <div className="h-8 w-8 flex items-center">
                <IconButton variant={'ghost'} icon={<Plus />}>
                  생성
                </IconButton>
              </div>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-36">
              <DropdownMenuGroup>
                <DropdownMenuItem onClick={() => createHandler({type: 'editor', data: {data: ''}})}>
                  Create resource
                </DropdownMenuItem>
              </DropdownMenuGroup>
            </DropdownMenuContent>
          </DropdownMenu>
          {checkIsOpen() ? (
            <IconButton
              variant={'ghost'}
              onClick={() => {
                if (tabStore.lastSelectId !== null) {
                  onResize(defaultEditorSize, true);
                } else {
                  if (tabStore.tabData.size !== 0) {
                    onResize(defaultEditorSize, true);
                  }
                }
              }}
              icon={<ChevronUp />}
              disabled={tabStore.tabData.size === 0}
            >
              위로
            </IconButton>
          ) : (
            <IconButton
              variant={'ghost'}
              onClick={() => onResize(minHeight, true)}
              icon={<ChevronDown />}
            >
              아래로
            </IconButton>
          )}
          {height === dimensions.height ? (
            <IconButton
              variant={'ghost'}
              onClick={() => {
                // qa 이슈에 따라 클릭시 디폴크 크기로 최소화 하도록 수정
                // onResize(minHeight, true)
                onResize(defaultEditorSize, true);
              }}
              icon={<MinimizeIcon />}
            >
              화면 축소
            </IconButton>
          ) : (
            <IconButton
              variant={'ghost'}
              onClick={() => {
                tabStore.tabData.size !== 0 && onResize(dimensions.height, true);
              }}
              icon={<MaximizeIcon />}
              disabled={tabStore.tabData.size === 0}
            >
              전체 화면
            </IconButton>
          )}
        </div>
      </header>
      {height > minHeight && (
        <div style={{height: height - minHeight}}>
          {tabStore.lastSelectId && (
            <TabContent
              tabId={tabStore.lastSelectId}
              height={height - minHeight}
              onResize={onResize}
            />
          )}
        </div>
      )}
    </div>
  );
};

export {Control};
