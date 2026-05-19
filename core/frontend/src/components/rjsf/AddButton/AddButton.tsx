import {FormContextType, IconButtonProps, RJSFSchema, StrictRJSFSchema} from '@rjsf/utils';
import {Button} from '@pharos/shared/components/ui';
import {PlusCircle} from '@pharos/shared/components';

export default function AddButton<
  T = any,
  S extends StrictRJSFSchema = RJSFSchema,
  F extends FormContextType = any,
>({uiSchema, registry, ...props}: IconButtonProps<T, S, F>) {
  return (
    <div className="add-button p-0 m-0">
      <Button {...props} className="w-fit gap-2" variant="outline">
        <PlusCircle size={16}></PlusCircle> Add
      </Button>
    </div>
  );
}
