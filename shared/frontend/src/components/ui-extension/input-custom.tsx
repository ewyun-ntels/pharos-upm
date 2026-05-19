import * as React from 'react';
import {cn} from '../../lib';

export interface InputProps
  extends React.InputHTMLAttributes<HTMLInputElement | HTMLTextAreaElement> {
  multiLine?: boolean;
}

const InputCustom = React.forwardRef<HTMLInputElement | HTMLTextAreaElement, InputProps>(
  ({multiLine, className, type, ...props}, ref) => {
    const textareaRef = React.useRef<HTMLTextAreaElement | null>(null);

    /* textarea 데이터에 맞도록 height 조정 */
    const resizeTextarea = () => {
      if (textareaRef.current) {
        const textarea = textareaRef.current;
        textarea.style.height = 'auto';
        textarea.style.height = `${textarea.scrollHeight + 5}px`;
      }
    };

    React.useEffect(() => {
      if (multiLine) {
        resizeTextarea();
      }
    }, [multiLine, props.value]);

    if (!multiLine) {
      return (
        <input
          type={type}
          className={cn(
            'flex w-full rounded-md border border-input bg-transparent px-3 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
            'h-9 py-1',
            className,
          )}
          ref={ref as React.Ref<HTMLInputElement>}
          {...(props as React.InputHTMLAttributes<HTMLInputElement>)}
        />
      );
    } else {
      return (
        <textarea
          className={cn(
            'resize-none flex w-full rounded-md border border-input bg-transparent px-3 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
            'min-h-[50px] max-h-[500px] py-2 scroll-auto',
            className,
          )}
          ref={(el) => {
            textareaRef.current = el;
            if (typeof ref === 'function') ref(el);
            else if (ref) (ref as React.MutableRefObject<HTMLTextAreaElement | null>).current = el;
          }}
          {...(props as React.TextareaHTMLAttributes<HTMLTextAreaElement>)}
          onInput={resizeTextarea}
        />
      );
    }
  },
);

InputCustom.displayName = 'InputCustom';

export default InputCustom;
