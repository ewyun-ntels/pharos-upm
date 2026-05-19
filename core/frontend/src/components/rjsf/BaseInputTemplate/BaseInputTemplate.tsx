import {
  ariaDescribedByIds,
  BaseInputTemplateProps,
  examplesId,
  FormContextType,
  getInputProps,
  RJSFSchema,
  StrictRJSFSchema,
} from '@rjsf/utils';
import {ChangeEvent, FocusEvent} from 'react';
import {Input} from '@pharos/shared/components/ui';
import {cn} from '@lib/utils';

export default function BaseInputTemplate<
  T = any,
  S extends StrictRJSFSchema = RJSFSchema,
  F extends FormContextType = any,
>({
  id,
  placeholder,
  required,
  readonly,
  disabled,
  type,
  value,
  onChange,
  onChangeOverride,
  onBlur,
  onFocus,
  autofocus,
  options,
  schema,
  rawErrors = [],
  children,
  extraProps,
}: BaseInputTemplateProps<T, S, F>) {
  const inputProps = {
    ...extraProps,
    ...getInputProps<T, S, F>(schema, type, options),
  };
  const _onChange = ({target: {value}}: ChangeEvent<HTMLInputElement>) =>
    onChange(value === '' ? options.emptyValue : value);
  const _onBlur = ({target: {value}}: FocusEvent<HTMLInputElement>) => onBlur(id, value);
  const _onFocus = ({target: {value}}: FocusEvent<HTMLInputElement>) => onFocus(id, value);

  const targetIDNumber = [
    'root_integer',
    'root_number',
    'root_integerRange',
    'root_integerRangeSteps',
  ];
  const targetTypesDate = ['datetime-local', 'time', 'date'];
  const targetTypesETC = ['color', 'number'];
  const inputClassName = cn(
    targetTypesDate.includes(type) && 'w-fit',
    (targetTypesETC.includes(type) || targetIDNumber.includes(id)) && '!w-[160px]',
    rawErrors?.length > 0 ? 'border-red-500' : 'border-border',
  );

  return (
    <>
      <Input
        id={id}
        name={id}
        type={type}
        placeholder={placeholder}
        autoFocus={autofocus}
        required={required}
        disabled={disabled}
        readOnly={readonly}
        list={schema.examples ? examplesId(id) : undefined}
        {...inputProps}
        value={value || value === 0 ? value : ''}
        className={inputClassName}
        onChange={onChangeOverride || _onChange}
        onBlur={_onBlur}
        onFocus={_onFocus}
        aria-describedby={ariaDescribedByIds(id, !!schema.examples)}
      />
      {children}
      {Array.isArray(schema.examples) ? (
        <datalist id={examplesId(id)}>
          {(schema.examples as string[])
            .concat(
              schema.default && !schema.examples.includes(schema.default)
                ? ([schema.default] as string[])
                : [],
            )
            .map((example: any) => {
              return <option key={example} value={example} />;
            })}
        </datalist>
      ) : null}
    </>
  );
}
