import React, {useCallback, useLayoutEffect, useState} from 'react';
import {IconButton} from '@pharos/shared/components/ui-extension';
import {ChevronLeft, ChevronRight} from '@pharos/shared/components';
import {useTabStore} from '@components/controlBar/controlTabStore';
import useTerminalStore from '@components/terminal/terminalStore';
import {Tabs} from './controlTab';
import {defaultEditorSize} from '../layout/constants';

interface TabsSectionProps {
  onResize: (ySize: number, enable: boolean) => void;
  checkIsOpen: () => boolean;
  // 탭 라인 숨김 여부
  hideTabLine?: boolean;
  scrollRef: React.RefObject<HTMLDivElement | null>;
}

const TabsSection = ({onResize, checkIsOpen, hideTabLine, scrollRef}: TabsSectionProps) => {
  const tabStore = useTabStore();
  const removeChannel = useTerminalStore((state) => state.removeChannel); // terminal : remove Channel
  const [showScrollButtons, setShowScrollButtons] = useState(false);

  const scrollLeft = () => {
    if (scrollRef.current) {
      scrollRef.current.scrollBy({left: -100, behavior: 'smooth'});
    }
  };

  const scrollRight = () => {
    if (scrollRef.current) {
      scrollRef.current.scrollBy({left: 100, behavior: 'smooth'});
    }
  };

  const checkScrollButtonsVisibility = useCallback(() => {
    if (scrollRef.current) {
      const scrollWidth = scrollRef.current.scrollWidth;
      const clientWidth = scrollRef.current.clientWidth;

      // TOOD 하단과 같은 이유로 인하여 +1 추가
      // setShowScrollButtons(scrollWidth > clientWidth + 1);에서 clientWidth + 1을 사용하여 버튼이 정상적으로 표시되는 이유는 미세한 오차를 보정하여 확대된 상태에서 발생하는 비교 오류를 회피하기 때문입니다.
      //
      // 배경: 브라우저 확대 시 발생하는 오차
      // 브라우저에서 확대가 발생하면 다음과 같은 일이 일어날 수 있습니다:
      //
      // scrollWidth와 clientWidth의 미세한 차이: 확대된 상태에서는 렌더링 방식에 따라 scrollWidth와 clientWidth 사이에 미세한 오차가 발생할 수 있습니다. 이 오차는 브라우저가 확대 비율에 맞춰 CSS 픽셀을 소수점 이하로 처리하면서 발생하는 것으로, 1픽셀 이하의 차이가 생길 수 있습니다.
      // 정확한 스크롤 상태 판단이 어려움: scrollWidth와 clientWidth가 거의 동일하게 계산될 경우, 스크롤이 필요하지 않음에도 불구하고 스크롤이 필요한 것처럼 판별될 수 있습니다. 예를 들어, 확대된 상태에서 scrollWidth와 clientWidth가 1픽셀 정도 차이가 나는 경우, scrollWidth > clientWidth 비교만으로는 스크롤 버튼을 정확히 표시하지 못하게 됩니다.
      // +1을 추가한 이유와 효과
      // 오차 보정: scrollWidth > clientWidth + 1로 비교할 때, 이 +1 픽셀은 확대 비율에 따른 미세한 오차를 허용하여 정상적인 판단을 도와줍니다. 만약 이 오차가 실제로 스크롤이 필요할 만큼 크다면 scrollWidth는 clientWidth + 1을 초과하게 되어 스크롤 버튼을 표시하도록 하게 됩니다.
      // 확대 상태에서도 정확한 스크롤 판별: 이 추가된 1픽셀은 특히 확대된 상태에서도 정확하게 스크롤이 필요할 때만 버튼을 나타나게 합니다. 미세한 크기 차이를 보정하여 scrollWidth와 clientWidth가 거의 같을 때도 불필요한 스크롤 버튼 표시를 방지합니다.
      // 따라서 scrollWidth > clientWidth + 1 방식은 확대 상태에서의 불필요한 버튼 표시를 줄이고, 실제로 스크롤이 필요할 때만 버튼이 나타나도록 보정하는 데 효과적입니다.
      setShowScrollButtons(scrollWidth > clientWidth + 1);
    }
  }, [scrollRef]);

  useLayoutEffect(() => {
    checkScrollButtonsVisibility();
    window.addEventListener('resize', checkScrollButtonsVisibility);

    return () => {
      window.removeEventListener('resize', checkScrollButtonsVisibility);
    };
  }, [checkScrollButtonsVisibility, tabStore.tabData.size]);

  return (
    <>
      {tabStore.tabData.size !== 0 && showScrollButtons && (
        <IconButton
          variant="ghost"
          className="h-8 w-8 m-[2px] mx-1"
          onClick={scrollLeft}
          icon={<ChevronLeft />}
        />
      )}
      <div
        ref={scrollRef}
        style={{height: 36}}
        className="w-full flex overflow-x-scroll overflow-y-hidden scrollbar-hide"
      >
        {Array.from(tabStore.tabData.keys()).map((tabId) => {
          const data = tabStore.tabData.get(tabId);
          const tabTypeData = data?.type ?? 'editor';

          return (
            <Tabs
              tabLineHide={hideTabLine}
              selected={tabStore.lastSelectId === tabId}
              onClick={() => {
                tabStore.setLastSelectId(tabId);
                checkIsOpen() && onResize(defaultEditorSize, true);
              }}
              onCancelClick={() => {
                // Terminal 일 경우 store 삭제
                if (data?.type === 'terminal')
                  removeChannel(`${data.pod}-${data.namespace}-${data.selectContainer}`);

                tabStore.remove(tabId);
                if (tabStore.tabData.size === 0) {
                  onResize(0, false);
                }
              }}
              tabId={tabId}
              key={tabId}
              label={data?.label}
              tabType={tabTypeData} // 탭 타입별 아이콘 삽입을 위한 타입
            />
          );
        })}
      </div>
      {tabStore.tabData.size !== 0 && showScrollButtons && (
        <IconButton
          variant="ghost"
          className="h-8 w-8 m-[2px] mx-1"
          onClick={scrollRight}
          icon={<ChevronRight />}
        />
      )}
    </>
  );
};

export {TabsSection};
