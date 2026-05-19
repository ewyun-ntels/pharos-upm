'use client';

import React, {
  createContext,
  PropsWithChildren,
  useEffect,
  useRef,
  useState,
} from 'react';
import {create} from 'zustand';
import {Resizable} from 're-resizable';
import {useWindowDimensions} from '@lib/interanl/window-demensions';
import {Content} from '@components/layout/content';
import {persist} from 'zustand/middleware';
import {Sheet} from '@components/sheet/sheet';
import {cn} from '@lib/utils';
import {useSheetStore} from '../sheet/sheetStore';
import {SidebarInset, SidebarProvider, SidebarTrigger} from '@pharos/shared/components/ui';
import {AppSidebar} from '@/features/menu/components/AppSidebar';
import {useFullscreenStore} from '@features/dashboard/hooks/fullscreenStore';

const defaultSheetXSize = 560;
const sheetMinXSize = 400;
const sheetMaxXSize = 1000;
const defaultSnapSheetGap = 20;

enum MenuType {
  Icon = 'icon',
  Full = 'full',
  Hide = 'hide',
}

type MenuLayout = {
  pin: boolean;
  type: MenuType;
  set: (pin: boolean, type: MenuType) => void;
};

const useMenuLayout = create<MenuLayout>()(
  persist(
    (set) => ({
      pin: true,
      type: MenuType.Full,
      set: (pin, type) => set(() => ({pin, type})),
    }),
    {name: 'n-component-template-menu-layout'},
  ),
);

export {useMenuLayout, MenuType};

export const IsOutsideContext = createContext({
  isOutsideClickEnabled: true,
  fnSetOutsideClickEnabled: (_: boolean) => {},
});

export const Layout: React.FC<PropsWithChildren> = ({children}) => {
  const menuLayout = useMenuLayout();
  const sheetLayout = useSheetStore();
  const {height} = useWindowDimensions();
  const [isResizing, setIsResizing] = useState(false);
  const [isOutsideClickEnabled, setIsOutsideClickEnabled] = useState(true);
  const prevMenuTypeRef = useRef<{type: MenuType; pin: boolean} | null>(null);
  const [showKioskHint, setShowKioskHint] = useState(false);
  const hintTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const {isFullscreen, exitFullscreen, isSidebarVisible} = useFullscreenStore();

  const sidebarOpen = isSidebarVisible && menuLayout.type !== MenuType.Hide;

  const fnSetOutsideClickEnabled = (value: boolean) => {
    setIsOutsideClickEnabled(value);
  };

  // fullscreen 진입/탈출 시 menuLayout 조작
  useEffect(() => {
    if (isFullscreen) {
      prevMenuTypeRef.current = {type: menuLayout.type, pin: menuLayout.pin};
      menuLayout.set(false, MenuType.Hide);
      setShowKioskHint(true);
      if (hintTimeoutRef.current) clearTimeout(hintTimeoutRef.current);
      hintTimeoutRef.current = setTimeout(() => setShowKioskHint(false), 5000);
    } else {
      const restored = prevMenuTypeRef.current || {type: MenuType.Full, pin: true};
      menuLayout.set(restored.pin, restored.type);
      setShowKioskHint(false);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isFullscreen]);

  // ESC 키로 fullscreen 탈출
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') exitFullscreen();
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [exitFullscreen]);

  return (
    <SidebarProvider
      className="h-screen overflow-hidden"
      open={sidebarOpen}
      onOpenChange={(open) => {
        menuLayout.set(false, open ? MenuType.Full : MenuType.Hide);
      }}
    >
      <AppSidebar />

      <SidebarInset className="overflow-hidden bg-background">
        {!sidebarOpen && !isFullscreen && <SidebarTrigger className="absolute top-2 left-2 z-40" />}
        <IsOutsideContext.Provider value={{isOutsideClickEnabled, fnSetOutsideClickEnabled}}>
          <div className="flex-1 overflow-hidden h-full">
            <Content>{children}</Content>
          </div>

          {/* Sheet (오른쪽 패널) */}
          {sheetLayout.sheetStopXSize !== 0 && (
            <Resizable
              style={{right: 0, position: 'absolute'}}
              snap={{x: [sheetMinXSize, sheetMaxXSize]}}
              snapGap={defaultSnapSheetGap}
              size={{
                height: height,
                width: sheetLayout.sheetStopXSize,
              }}
              onResizeStart={() => setIsResizing(true)}
              onResizeStop={(_event, _direction, _elementRef, delta) => {
                sheetLayout.setLayout(delta.width + sheetLayout.sheetStopXSize);
                setIsResizing(false);
              }}
              className={cn('overflow-hidden border border-t-0 border-border bg-card h-full z-50')}
              minWidth={sheetMinXSize}
              maxWidth={sheetMaxXSize}
              enable={{
                top: false,
                right: false,
                bottom: false,
                left: true,
                topRight: false,
                bottomRight: false,
                bottomLeft: false,
                topLeft: false,
              }}
            >
              <Sheet setResizing={isResizing} isOutsideClickEnabled={isOutsideClickEnabled} />
            </Resizable>
          )}
        </IsOutsideContext.Provider>
      </SidebarInset>

      {showKioskHint && (
        <div className="fixed bottom-6 left-1/2 transform -translate-x-1/2 bg-black text-white border text-sm px-4 py-2 rounded-md z-9999 shadow-lg">
          Press ESC to exit full screen.
        </div>
      )}
    </SidebarProvider>
  );
};

export {defaultSheetXSize};
