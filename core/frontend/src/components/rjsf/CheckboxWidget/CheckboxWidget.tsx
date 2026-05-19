import {
  ariaDescribedByIds,
  descriptionId,
  FormContextType,
  getTemplate,
  labelValue,
  RJSFSchema,
  schemaRequiresTrueValue,
  StrictRJSFSchema,
  WidgetProps,
} from '@rjsf/utils';
import {Checkbox} from '@pharos/shared/components/ui';

export default function CheckboxWidget<
  T = any,
  S extends StrictRJSFSchema = RJSFSchema,
  F extends FormContextType = any,
>(props: WidgetProps<T, S, F>) {
  const {
    id,
    value,
    disabled,
    readonly,
    label,
    hideLabel,
    schema,
    autofocus,
    options,
    onChange,
    registry,
    uiSchema,
  } = props;
  // Because an unchecked checkbox will cause html5 validation to fail, only add
  // the "required" attribute if the field value must be "true", due to the
  // "const" or "enum" keywords
  const required = schemaRequiresTrueValue<S>(schema);
  const DescriptionFieldTemplate = getTemplate<'DescriptionFieldTemplate', T, S, F>(
    'DescriptionFieldTemplate',
    registry,
    options,
  );

  const description = options.description || schema.description;
  return (
    <div
      className={`relative ${disabled || readonly ? 'cursor-not-allowed opacity-50' : ''}`}
      aria-describedby={ariaDescribedByIds(id)}
    >
      {!hideLabel && !!description && (
        <DescriptionFieldTemplate
          id={descriptionId(id)}
          description={description}
          schema={schema}
          uiSchema={uiSchema}
          registry={registry}
        />
      )}
      <div className="flex items-center space-x-2">
        <Checkbox
          id={id}
          name={id}
          required={required}
          disabled={disabled || readonly}
          autoFocus={autofocus}
          onCheckedChange={onChange}
          defaultChecked={typeof value !== 'undefined'}
        />
        <label htmlFor={id} className="form-checkbox text-primary">
          {labelValue(label, hideLabel || !label)}
        </label>
      </div>
    </div>
  );
}
