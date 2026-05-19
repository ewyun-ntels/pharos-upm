import {
  enumOptionsIndexForValue,
  FormContextType,
  RJSFSchema,
  StrictRJSFSchema,
  WidgetProps,
} from '@rjsf/utils';
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@pharos/shared/components/ui';
import {FancyMultiSelect} from '../fancy-multi-select';
import {useState} from 'react';

export default function SelectWidget<
  T = any,
  S extends StrictRJSFSchema = RJSFSchema,
  F extends FormContextType = any,
>({
  options,
  required,
  disabled,
  value,
  multiple,
  onChange,
  placeholder,
}: WidgetProps<T, S, F>) {
  const {enumOptions, enumDisabled} = options;

  const [selectedIndex, setSelectedIndex] = useState(
    enumOptionsIndexForValue<S>(value, enumOptions, false) as unknown as string,
  );

  return !multiple ? (
    <Select
      required={required}
      disabled={disabled}
      value={selectedIndex}
      onValueChange={(v) => {
        setSelectedIndex(v);
        onChange((enumOptions as any)[v].value);
      }}
    >
      <SelectTrigger className="w-45">
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          {(enumOptions as any).map(({value: _value, label}: any, i: number) => {
            const disabled: any =
              Array.isArray(enumDisabled) && (enumDisabled as any).indexOf(_value) != -1;
            return (
              <SelectItem key={i} id={label} value={i.toString()} disabled={disabled}>
                {label}
              </SelectItem>
            );
          })}
        </SelectGroup>
      </SelectContent>
    </Select>
  ) : (
    <FancyMultiSelect
      multiple
      items={enumOptions}
      selected={value}
      onValueChange={onChange}
    ></FancyMultiSelect>
  );
}
