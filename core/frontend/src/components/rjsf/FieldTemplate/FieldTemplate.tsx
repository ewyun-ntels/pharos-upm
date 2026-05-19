import { Title } from '@pharos/shared/components/ui-extension';
import {
  FieldTemplateProps,
  FormContextType,
  getTemplate,
  getUiOptions,
  RJSFSchema,
  StrictRJSFSchema,
} from '@rjsf/utils';

export default function FieldTemplate<
  T = any,
  S extends StrictRJSFSchema = RJSFSchema,
  F extends FormContextType = any,
>({
  id,
  children,
  displayLabel,
  rawErrors = [],
  errors,
  help,
  description,
  rawDescription,
  classNames,
  style,
  disabled,
  label,
  hidden,
  onRemoveProperty,
  onKeyRename,
  onKeyRenameBlur,
  readonly,
  required,
  schema,
  uiSchema,
  registry,
}: FieldTemplateProps<T, S, F>) {
  const uiOptions = getUiOptions(uiSchema);
  const WrapIfAdditionalTemplate = getTemplate<'WrapIfAdditionalTemplate', T, S, F>(
    'WrapIfAdditionalTemplate',
    registry,
    uiOptions,
  );
  if (hidden) {
    return <div className="hidden">{children}</div>;
  }
  return (
    <WrapIfAdditionalTemplate
      classNames={classNames}
      style={style}
      disabled={disabled}
      id={id}
      label={label}
      onRemoveProperty={onRemoveProperty}
      onKeyRename={onKeyRename}
      onKeyRenameBlur={onKeyRenameBlur}
      readonly={readonly}
      required={required}
      schema={schema}
      uiSchema={uiSchema}
      registry={registry}
    >
      <div className="flex flex-col max-w-3xl">
        {displayLabel && (
          <Title
            variant="formLabel"
            htmlFor={id}
            className={rawErrors.length > 0 ? 'text-red-500' : ''}
          >
            {label}
            {required ? '*' : null}
          </Title>
        )}
        <div className="children">
          {!(schema?.properties && Object.keys(schema.properties).length === 0) && children}
          {displayLabel && rawDescription && (
            <small className="mt-2 block text-xs">
              <div className={`${rawErrors.length > 0 ? 'text-red-500' : 'text-muted-foreground'}`}>
                {description}
              </div>
            </small>
          )}
          {errors}
          {help}
        </div>
      </div>
    </WrapIfAdditionalTemplate>
  );
}
