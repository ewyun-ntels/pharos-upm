import React, {useRef} from 'react';
import {useSheetStore} from '@components/sheet/sheetStore';
import {useOutsideClick} from 'rooks';
import {useCloseSheet} from '@components/sheet/useCloseSheet';

type SheetProps = {
  setResizing: boolean;
  isOutsideClickEnabled: boolean;
};

const hasUnitElementOpen = () => {
  const elements: NodeListOf<HTMLElement> = document.getElementsByName('units');

  return Array.from(elements).some((element: HTMLElement) => {
    const ariaExpandedValue = element.getAttribute('aria-expanded');
    const dataStateValue = element.getAttribute('data-state');
    return 'true' === ariaExpandedValue && 'open' === dataStateValue;
  });
};

/// alert > pagination click
/// pagination click 시에 sheet가 자동으로 닫히는걸 방지
const isPaginationClick = (ev: MouseEvent | TouchEvent) => {
  let node = ev.target as HTMLElement | null;
  const paginationChild = node!.querySelector('.custom-sheet-content');
  return !!paginationChild;
};

/// Radix UI Dialog/AlertDialog 등 Portal로 body에 마운트된 오버레이가 열려있는 경우
/// Sheet 외부 클릭으로 처리하지 않음
const isRadixOverlayOpen = () => {
  return !!document.querySelector('[role="dialog"][data-state="open"], [role="alertdialog"][data-state="open"]');
};

function Sheet({setResizing, isOutsideClickEnabled = true}: SheetProps) {
  const sheet = useSheetStore();
  const ref = useRef<HTMLDivElement>(null);
  const {getRegistryRefArray} = useCloseSheet();

  useOutsideClick(
    ref,
    (ev) => {
      const filter = getRegistryRefArray().filter((v) => {
        return !!v.current?.contains(ev.target as Node);
      });

      if (filter.length === 0 && !setResizing) {
        if (hasUnitElementOpen()) return;
        if (isPaginationClick(ev)) return;
        if (isRadixOverlayOpen()) return;

        sheet.setClose();
      }
    },
    isOutsideClickEnabled,
  );

  return (
    <div ref={ref} className={'flex bg-background h-full overflow-auto'}>
      {sheet.sheet}
    </div>
  );
}

export {Sheet};
