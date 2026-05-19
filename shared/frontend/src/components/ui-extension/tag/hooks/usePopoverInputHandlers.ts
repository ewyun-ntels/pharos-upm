import {useCallback} from 'react';

interface UsePopoverInputHandlersProps {
  isPopoverOpen: boolean;
  setInputFocused: (focused: boolean) => void;
  setIsPopoverOpen: (open: boolean) => void;
  children: React.ReactNode;
}

interface UsePopoverInputHandlersReturn {
  handleInputFocus: (event: React.FocusEvent<HTMLInputElement>) => void;
  handleInputBlur: (event: React.FocusEvent<HTMLInputElement>) => void;
}

/**
 * Custom hook for managing popover input focus/blur handlers
 * 
 * Handles input focus/blur events while preserving user-defined event handlers
 * and managing popover open state based on input focus.
 */
export const usePopoverInputHandlers = ({
  isPopoverOpen,
  setInputFocused,
  setIsPopoverOpen,
  children,
}: UsePopoverInputHandlersProps): UsePopoverInputHandlersReturn => {
  const handleInputFocus = useCallback(
    (event: React.FocusEvent<HTMLInputElement>) => {
      // Only set inputFocused to true if the popover is already open.
      // This will prevent the popover from opening due to an input focus if it was initially closed.
      if (isPopoverOpen) {
        setInputFocused(true);
      }

      const userOnFocus = (children as React.ReactElement<any>).props.onFocus;
      if (userOnFocus) userOnFocus(event);
    },
    [isPopoverOpen, setInputFocused, children],
  );

  const handleInputBlur = useCallback(
    (event: React.FocusEvent<HTMLInputElement>) => {
      setInputFocused(false);

      // Allow the popover to close if no other interactions keep it open
      if (!isPopoverOpen) {
        setIsPopoverOpen(false);
      }

      const userOnBlur = (children as React.ReactElement<any>).props.onBlur;
      if (userOnBlur) userOnBlur(event);
    },
    [isPopoverOpen, setInputFocused, setIsPopoverOpen, children],
  );

  return {
    handleInputFocus,
    handleInputBlur,
  };
};
