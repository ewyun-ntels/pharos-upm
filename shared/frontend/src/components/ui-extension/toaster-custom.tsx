'use client';

import {
  Toast,
  ToastClose,
  ToastDescription,
  ToastProvider,
  ToastTitle,
  ToastViewport,
} from '@pharos/shared/components/ui';
import {useToast} from '@pharos/shared/components/ui';
export function ToasterCustom() {
  const {toasts} = useToast();

  return (
    <ToastProvider>
      {toasts.length === 0 && <div>test</div>}
      {toasts.map(({id, title, description, action, ...props}) => (
        <Toast key={id} {...props}>
          <div className="grid gap-1">
            {title && <ToastTitle>{title}</ToastTitle>}
            {description && <ToastDescription>{description}</ToastDescription>}
          </div>
          {action}
          <ToastClose />
        </Toast>
      ))}
      <ToastViewport />
    </ToastProvider>
  );
}
