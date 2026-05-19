import React, {useCallback, useEffect, useRef, useState} from 'react';
import {Popover, PopoverContent, PopoverTrigger} from '@pharos/shared/components/ui';
import {type Tag as TagType} from './tag-input';
import {TagList, TagListProps} from './tag-list';
import {Button} from '@pharos/shared/components/ui';
import {usePopoverInputHandlers} from './hooks/usePopoverInputHandlers';

type TagPopoverProps = {
  children: React.ReactNode;
  tags: TagType[];
  customTagRenderer?: (tag: TagType, isActiveTag: boolean) => React.ReactNode;
  activeTagIndex?: number | null;
  setActiveTagIndex?: (index: number | null) => void;
} & TagListProps;

export const TagPopover: React.FC<TagPopoverProps> = ({
  children,
  tags,
  customTagRenderer,
  activeTagIndex,
  setActiveTagIndex,
  ...tagProps
}) => {
  const triggerContainerRef = useRef<HTMLDivElement | null>(null);
  const triggerRef = useRef<HTMLButtonElement | null>(null);
  const inputRef = useRef<HTMLInputElement | null>(null);

  const [popoverWidth, setPopoverWidth] = useState<number>(0);
  const [isPopoverOpen, setIsPopoverOpen] = useState(false);
  const [inputFocused, setInputFocused] = useState(false);
  const [sideOffset, setSideOffset] = useState<number>(0);

  useEffect(() => {
    const handleResize = () => {
      if (triggerContainerRef.current && triggerRef.current) {
        setPopoverWidth(triggerContainerRef.current.offsetWidth);
        setSideOffset(triggerContainerRef.current.offsetWidth - triggerRef?.current?.offsetWidth);
      }
    };

    handleResize(); // Call on mount and layout changes

    window.addEventListener('resize', handleResize); // Adjust on window resize
    return () => window.removeEventListener('resize', handleResize);
  }, [triggerContainerRef, triggerRef]);

  const handleOpenChange = useCallback(
    (open: boolean) => {
      if (open && triggerContainerRef.current) {
        setPopoverWidth(triggerContainerRef.current.offsetWidth);
      }

      if (open) {
        inputRef.current?.focus();
        setIsPopoverOpen(open);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [inputFocused],
  );

  const {handleInputFocus, handleInputBlur} = usePopoverInputHandlers({
    isPopoverOpen,
    setInputFocused,
    setIsPopoverOpen,
    children,
  });

  return (
    <Popover open={isPopoverOpen} onOpenChange={handleOpenChange}>
      <div
        className="relative flex items-center rounded-md border border-input bg-transparent pr-3"
        ref={triggerContainerRef}
      >
        {/* {children} */}
        {React.cloneElement(children as React.ReactElement<any>, {
          onFocus: handleInputFocus,
          onBlur: handleInputBlur,
          ref: inputRef,
        })}
        <PopoverTrigger asChild>
          <Button
            ref={triggerRef}
            variant="ghost"
            size="icon"
            role="combobox"
            className="hover:bg-transparent"
            onClick={() => setIsPopoverOpen(!isPopoverOpen)}
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              className="lucide lucide-chevron-down h-4 w-4 shrink-0 opacity-50"
            >
              <path d="m6 9 6 6 6-6"></path>
            </svg>
          </Button>
        </PopoverTrigger>
      </div>
      <PopoverContent
        className="w-full space-y-3"
        style={{
          marginLeft: `-${sideOffset}px`,
          width: `${popoverWidth}px`,
        }}
      >
        <div className="space-y-1">
          <h4 className="text-sm font-medium leading-none">Entered Tags</h4>
          <p className="text-sm text-muted-foregrounsd text-left">
            These are the tags you&apos;ve entered.
          </p>
        </div>
        <TagList
          tags={tags}
          customTagRenderer={customTagRenderer}
          activeTagIndex={activeTagIndex}
          setActiveTagIndex={setActiveTagIndex}
          {...tagProps}
        />
      </PopoverContent>
    </Popover>
  );
};
