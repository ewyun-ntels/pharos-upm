import {
  FormContextType,
  getUiOptions,
  RJSFSchema,
  StrictRJSFSchema,
  TitleFieldProps,
} from "@rjsf/utils"
import { Title } from "@pharos/shared/components/ui-extension"

export default function TitleField<
  T = any,
  S extends StrictRJSFSchema = RJSFSchema,
  F extends FormContextType = any,
>({ id, title, uiSchema }: TitleFieldProps<T, S, F>) {
  const uiOptions = getUiOptions<T, S, F>(uiSchema)

  let titleClass = "text-md font-medium" // 기본 타이틀 스타일

  if (id === "root__title") {
    titleClass = "text-lg font-bold"
  } 
  else if (/^root_.*_\d+__title$/.test(id)) {
    titleClass = "text-sm mt-0 pt-3 font-bold text-gray-500 border-border border-t"
  }

  return (
    <div id={id}>
      <Title variant ="h3" className={`leading-tight ${titleClass}`}>
        {uiOptions.title || title}
      </Title>
    </div>
  )
}
