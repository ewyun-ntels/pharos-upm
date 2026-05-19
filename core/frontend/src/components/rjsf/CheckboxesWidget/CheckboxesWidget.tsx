import {
  ariaDescribedByIds,
  enumOptionsDeselectValue,
  enumOptionsIsSelected,
  enumOptionsSelectValue,
  FormContextType,
  RJSFSchema,
  StrictRJSFSchema,
  WidgetProps,
} from '@rjsf/utils';
import {Checkbox} from '@pharos/shared/components/ui';
import {v4 as uuidv4} from 'uuid';

export default function CheckboxesWidget<
  T = any,
  S extends StrictRJSFSchema = RJSFSchema,
  F extends FormContextType = any,
>({
  id,
  disabled,
  options,
  value,
  autofocus,
  readonly,
  required,
  onChange,
}: WidgetProps<T, S, F>) {
  const {enumOptions, enumDisabled} = options;
  const checkboxesValues = Array.isArray(value) ? value : [value];


  return (
    <div className="space-y-2">
      {Array.isArray(enumOptions) &&
        enumOptions.map((option, index: number) => {
          const checked = enumOptionsIsSelected<S>(option.value, checkboxesValues);
          const itemDisabled =
            Array.isArray(enumDisabled) && enumDisabled.indexOf(option.value) !== -1;

          return (
            <div className="flex items-center space-x-2" key={uuidv4()}>
              <Checkbox
                id={id}
                name={id}
                required={required}
                disabled={disabled || itemDisabled || readonly}
                onCheckedChange={(state) => {
                  if (state) {
                    onChange(enumOptionsSelectValue<S>(index, checkboxesValues, enumOptions));
                  } else {
                    onChange(enumOptionsDeselectValue<S>(index, checkboxesValues, enumOptions));
                  }
                }}
                defaultChecked={checked}
                autoFocus={autofocus && index === 0}
                aria-describedby={ariaDescribedByIds(id)}
              />
              <label htmlFor={id} className="form-checkbox text-primary">
                {option.label}
              </label>
            </div>
          );
        })}
    </div>
  );
}
