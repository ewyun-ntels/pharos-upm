import * as React from 'react';
import { cn } from '../../lib';
import { Input, Switch, Checkbox, Textarea, Select, SelectContent, SelectTrigger, SelectValue } from '../ui';


// ─── FormLabel (<label> + required * 표시) ───────────────────────────────────
export interface FormLabelProps extends React.LabelHTMLAttributes<HTMLLabelElement> {
  required?: boolean;
}

const FormLabel = React.forwardRef<HTMLLabelElement, FormLabelProps>(
  ({ className, children, required, ...props }, ref) => (
    <label
      ref={ref}
      className={cn('text-sm font-medium leading-none flex items-center gap-1.5', className)}
      {...props}
    >
      {children}
      {required && (
        <span className="text-destructive text-sm leading-none" aria-hidden>*</span>
      )}
    </label>
  ),
);
FormLabel.displayName = 'FormLabel';


export interface FieldInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  description?: string;
  required?: boolean;
  error?: string | boolean;
  containerClassName?: string;
  children?: React.ReactNode;
}

const FieldInput = React.forwardRef<HTMLInputElement, FieldInputProps>(
  ({ className, containerClassName, label, description, required, error, children, ...props }, ref) => (
    <div 
      className={cn('flex flex-col gap-1.5', containerClassName)}
      data-invalid={error ? true : undefined}
    >
      {label && <FormLabel required={required} htmlFor={props.id}>{label}</FormLabel>}
      {description && (
        <p className="text-[13px] text-muted-foreground">{description}</p>
      )}
      <Input
        ref={ref}
        required={required}
        className={className}
        aria-invalid={error ? true : undefined}
        {...props} 
      />
      {typeof error === 'string' && (
        <p className="text-[13px] font-medium text-destructive">{error}</p>
      )}
      {children}
    </div>
  ),
);
FieldInput.displayName = 'FieldInput';

// ─── FieldTextarea ────────────────────────────────────────────────────────────
export interface FieldTextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string;
  description?: string;
  required?: boolean;
  error?: string | boolean;
  containerClassName?: string;
  children?: React.ReactNode;
}

const FieldTextarea = React.forwardRef<HTMLTextAreaElement, FieldTextareaProps>(
  ({ className, containerClassName, label, description, required, error, children, ...props }, ref) => (
    <div 
      className={cn('flex flex-col gap-1.5', containerClassName)}
      data-invalid={error ? true : undefined}
    >
      {label && <FormLabel required={required} htmlFor={props.id}>{label}</FormLabel>}
      {description && (
        <p className="text-[13px] text-muted-foreground">{description}</p>
      )}
      <Textarea
        ref={ref}
        required={required}
        className={className}
        aria-invalid={error ? true : undefined}
        {...props} 
      />
      {typeof error === 'string' && (
        <p className="text-[13px] font-medium text-destructive">{error}</p>
      )}
      {children}
    </div>
  ),
);
FieldTextarea.displayName = 'FieldTextarea';

// ─── FieldSelect ──────────────────────────────────────────────────────────────
export interface FieldSelectProps extends React.ComponentPropsWithoutRef<typeof Select> {
  label?: string;
  description?: string;
  required?: boolean;
  containerClassName?: string;
  className?: string;
  id?: string;
  placeholder?: string;
}

const FieldSelect = React.forwardRef<HTMLDivElement, FieldSelectProps>(
  (
    { className, containerClassName, label, description, required, children, id, placeholder, ...props },
    ref
  ) => {
    const descriptionId = id ? `${id}-description` : undefined;

    return (
      <div ref={ref} className={cn('flex flex-col gap-2', containerClassName)}>
        {label && <FormLabel required={required} htmlFor={id}>{label}</FormLabel>}
        {description && (
          <p className="text-[13px] text-muted-foreground" id={descriptionId}>{description}</p>
        )}
        <Select required={required} {...props}>
          <SelectTrigger id={id} className={className} aria-describedby={descriptionId}>
            <SelectValue placeholder={placeholder} />
          </SelectTrigger>
          <SelectContent>
            {children}
          </SelectContent>
        </Select>
      </div>
    );
  }
);
FieldSelect.displayName = 'FieldSelect';

export interface FieldSwitchProps extends React.ComponentPropsWithoutRef<typeof Switch> {
  label?: string;
  description?: string;
  required?: boolean;
  containerClassName?: string;
  children?: React.ReactNode;
}

const FieldSwitch = React.forwardRef<React.ComponentRef<typeof Switch>, FieldSwitchProps>(
  (
    { className, containerClassName, label, description, required, children, id, ...props },
    ref
  ) => {
    const descriptionId = id ? `${id}-description` : undefined;

    return (
      <div className={cn('flex flex-col gap-1.5', containerClassName)}>
        {(label || description) && (
          <div className="flex flex-col gap-1.5 shrink-0">
            {label && <FormLabel required={required} htmlFor={id}>{label}</FormLabel>}
            {description && (
              <p className="text-[13px] text-muted-foreground" id={descriptionId}>{description}</p>
            )}
          </div>
        )}
        <Switch
          ref={ref}
          id={id}
          required={required}
          className={className}
          aria-describedby={descriptionId}
          {...props}
        />
        {children}
      </div>
    );
  }
);
FieldSwitch.displayName = 'FieldSwitch';

// ─── FieldCheckbox ────────────────────────────────────────────────────────────
export interface FieldCheckboxProps extends React.ComponentPropsWithoutRef<typeof Checkbox> {
  label?: string;
  description?: string;
  required?: boolean;
  containerClassName?: string;
  children?: React.ReactNode;
}

const FieldCheckbox = React.forwardRef<React.ComponentRef<typeof Checkbox>, FieldCheckboxProps>(
  ({ className, containerClassName, label, description, required, children, id, ...props }, ref) => {
    const descriptionId = id ? `${id}-description` : undefined;

    return (
      <div className={cn('flex items-start gap-2', containerClassName)}>
        <Checkbox
          ref={ref}
          id={id}
          required={required}
          className={className}
          aria-describedby={descriptionId}
          {...props}
        />
        {(label || description) && (
          <div className="flex flex-col gap-1 mt-0.5">
            {label && <FormLabel required={required} htmlFor={id}>{label}</FormLabel>}
            {description && (
              <p className="text-[13px] text-muted-foreground" id={descriptionId}>{description}</p>
            )}
          </div>
        )}
        {children}
      </div>
    );
  }
);
FieldCheckbox.displayName = 'FieldCheckbox';

export {
  FormLabel,
  FieldInput,
  FieldTextarea,
  FieldSelect,
  FieldSwitch,
  FieldCheckbox,
};
