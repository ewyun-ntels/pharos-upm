import {
  ArrayFieldTemplateProps,
  FormContextType,
  getTemplate,
  getUiOptions,
  RJSFSchema,
  StrictRJSFSchema,
} from '@rjsf/utils';
import {Separator} from '@pharos/shared/components/ui';

export default function ArrayFieldTemplate<
  T = any,
  S extends StrictRJSFSchema = RJSFSchema,
  F extends FormContextType = any,
>(props: ArrayFieldTemplateProps<T, S, F>) {
  const {
    canAdd,
    disabled,
    fieldPathId,
    uiSchema,
    items,
    onAddClick,
    readonly,
    registry,
    required,
    schema,
    title,
  } = props;
  const uiOptions = getUiOptions<T, S, F>(uiSchema);
  const ArrayFieldDescriptionTemplate = getTemplate<'ArrayFieldDescriptionTemplate', T, S, F>(
    'ArrayFieldDescriptionTemplate',
    registry,
    uiOptions,
  );
  const ArrayFieldTitleTemplate = getTemplate<'ArrayFieldTitleTemplate', T, S, F>(
    'ArrayFieldTitleTemplate',
    registry,
    uiOptions,
  );
  // Button templates are not overridden in the uiSchema
  const {
    ButtonTemplates: {AddButton},
  } = registry.templates;

  return (
    <div>
      <div className="m-0 flex p-0">
        <div className="m-0 w-full p-0">
          <ArrayFieldTitleTemplate
            fieldPathId={fieldPathId}
            title={uiOptions.title || title}
            schema={schema}
            uiSchema={uiSchema}
            required={required}
            registry={registry}
          />
          <ArrayFieldDescriptionTemplate
            fieldPathId={fieldPathId}
            description={uiOptions.description || schema.description}
            schema={schema}
            uiSchema={uiSchema}
            registry={registry}
          />
          <div className="m-0 w-full p-0">
            {items.map((item) => item)}
            {canAdd ? (
              <>
                <Separator orientation="horizontal"></Separator>
                <div className="add-button mt-2 flex">
                  <AddButton
                    className="array-item-add"
                    onClick={onAddClick}
                    disabled={disabled || readonly}
                    uiSchema={uiSchema}
                    registry={registry}
                  />
                </div>
              </>
            ) : null}
          </div>
        </div>
      </div>
    </div>
  );
}
