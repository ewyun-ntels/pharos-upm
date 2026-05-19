import React, {PropsWithChildren, useState, useCallback} from 'react';
import {Resizable, ResizeCallback} from 're-resizable';
import {create} from 'zustand';
import {persist} from 'zustand/middleware';
import {useWindowDimensions} from '@lib/interanl/window-demensions';
const defaultContentYSize: number = 0;
const defaultSanpContentGap = 20;

type ContentLayout = {
  contentStopYSize: number;
  set: (size: number) => void;
};

const useContentLayout = create<ContentLayout>()(
  persist(
    (set) => ({
      contentStopYSize: defaultContentYSize,
      set: (size) => set({contentStopYSize: size}),
    }),
    {name: 'n-component-template-content-layout'},
  ),
);

type ContentStopYSize = {
  contentYSize: number;
  enable: boolean;
};

type ContentProps = PropsWithChildren<{
  // headerSize: number; //부모 layout에서 받음
}>;

const Content: React.FC<ContentProps> = ({children}) => {
  const {height} = useWindowDimensions();
  const contentLayout = useContentLayout();

  // TODO enable을 통합해서 관리할지는 추가 파악 필요
  const [, setContentYSize] = useState<ContentStopYSize>({
    contentYSize: contentLayout.contentStopYSize,
    enable: true,
  });

  const handleResize = useCallback<ResizeCallback>(
    (_event, _direction, _elementRef, delta) => {
      setContentYSize({
        contentYSize: contentLayout.contentStopYSize - delta.height,
        enable: true,
      });
    },
    [contentLayout],
  );

  const handleResizeStop = useCallback<ResizeCallback>(
    (_event, _direction, _elementRef, delta) => {
      contentLayout.set(contentLayout.contentStopYSize - delta.height);
    },
    [contentLayout],
  );

  return (
    <div>
      <Resizable
        snap={{y: [0 /* 최상단 */, height /* 최하단 */]}}
        snapGap={defaultSanpContentGap}
        boundsByDirection
        size={{height: height - contentLayout.contentStopYSize}}
        defaultSize={{width: '100%', height: height - defaultContentYSize}}
        onResize={handleResize}
        onResizeStop={handleResizeStop}
        minHeight={height}
        maxHeight={height}
        enable={{
          top: false,
          right: false,
          //TODO: bottom: contentYSize.enable, 추후 필요시 교체. 주석 처리 이유 : <Control> 컴포넌트를 삭제하면서 resizable이 필요없게 되어 bottom: false로 변경하여 기능을 비활성함
          bottom: false,
          left: false,
          topRight: false,
          bottomRight: false,
          bottomLeft: false,
          topLeft: false,
        }}
      >
        <div style={{width: '100%', height: '100%'}}>
          <div className={'content-wrap pb-0 h-full flex flex-col'}>{children}</div>
        </div>
      </Resizable>
      {/* <div className={'h-full sticky z-20'}>
        <Control
          height={contentYSize.contentYSize}
          onResize={(ySize, enable) => {
            setContentYSize({contentYSize: ySize, enable: enable});
            contentLayout.set(ySize);
          }}
        />
      </div> */}
    </div>
  );
};

export {Content, defaultContentYSize};
