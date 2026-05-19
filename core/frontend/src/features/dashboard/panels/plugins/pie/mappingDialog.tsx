import {Button} from '@pharos/shared/components/ui';
import {useEffect} from 'react';
import {useFieldArray, useForm} from 'react-hook-form';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@pharos/shared/components/ui';
import {Input} from '@pharos/shared/components/ui';
import {TrashIcon} from '@pharos/shared/components';
import {DialogDescription} from '@pharos/shared/components';
import {ColorMapping} from './types';

interface DialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  colorMappings?: ColorMapping[];
  onSave: (newMappings: ColorMapping[]) => void;
}

export const MappingDialog = ({open, onOpenChange, colorMappings = [], onSave}: DialogProps) => {
  const {control, register, getValues, setValue, watch, trigger, formState} = useForm();

  // useFieldArray를 사용하여 행 관리
  const {fields, append, remove} = useFieldArray({
    control,
    name: `mappings`, // Field Array의 이름
  });

  // 다이얼로그가 열릴 때 기존 데이터 로드
  useEffect(() => {
    if (open) {
      remove(); // 기존 행 제거
      if (colorMappings.length === 0) {
        append({legendName: '', color: '#ffffff'}); // 기본 행 추가
      } else {
        append(colorMappings); // 기존 데이터로 행 추가
      }
    } else {
      // 다이얼로그가 닫힐 때 필드 초기화
      remove(); // 모든 행 제거
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, colorMappings]);

  const handleAddMapping = () => {
    append({legendName: '', color: '#ffffff'}); // 새 행 추가
  };

  /* 
  const transformMappingData = (data: MappingItem[]): TransformedMappingData => {
    return data.reduce((acc: TransformedMappingData, {condition, color, displayText}) => {
      acc[condition] = {color, displayText};
      return acc;
    }, {});
  }; */

  const handleSave = async () => {
    const isValid = await trigger('mappings'); // 'mappings' 필드의 유효성 검사 수행
    if (!isValid) {
      console.log('Validation failed:', formState.errors);
      return; // 유효성 검사 실패 시 저장 로직 중단
    }

    const mappings = getValues('mappings'); // 현재 mappings 값 가져오기
    onSave(mappings); // props로 전달받은 onSave 콜백 호출
    onOpenChange(false); // 다이얼로그 닫기
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="max-w-none"
        style={{maxWidth: '70rem'}}
        onInteractOutside={(event) => event.preventDefault()}
      >
        <DialogHeader className="mr-2">
          <DialogTitle className="break-all">Value mappings</DialogTitle>
          <DialogDescription></DialogDescription>
        </DialogHeader>
        <div className="p-4">
          <div className="mb-4">
            <table className="w-full">
              <thead>
                <tr>
                  <th className="px-4 py-2 text-left text-sm font-normal">Condition</th>
                  {/* <th className="px-4 py-2 text-left text-sm font-normal">Display text</th> */}
                  <th className="px-4 py-2 text-center text-sm font-normal">Color</th>
                  <th className="px-4 py-2 text-left text-sm font-normal"></th>
                </tr>
              </thead>
              <tbody>
                {fields.map((field, index) => (
                  <tr key={field.id}>
                    <td className="border-gray-300 px-4 py-2">
                      <Input
                        className={`${
                          Array.isArray(formState.errors?.mappings) &&
                          formState.errors?.mappings[index]?.legendName
                            ? 'border-red-500' // 유효성 검사 실패 시 빨간 테두리
                            : 'border-gray-300' // 기본 테두리
                        } px-4 py-2`}
                        {...register(`mappings.${index}.legendName`, {
                          required: 'Condition is required',
                        })}
                        placeholder="Exact value to match"
                      />
                    </td>
                    {/* <td className="border-gray-300 px-4 py-2">
                      <Input
                        {...register(`mappings.${index}.displayText`)}
                        placeholder="Optional display text"
                      />
                    </td> */}
                    <td className="border-gray-300 px-4 py-2 text-center">
                      <div className="flex justify-center items-center">
                        <Input
                          type="color"
                          defaultValue={watch(`mappings.${index}.color`)} // 현재 색상 값
                          onBlur={(e) => {
                            setValue(`mappings.${index}.color`, e.target.value);
                          }}
                          className="w-10 h-10 p-0 border-none cursor-pointer"
                        />
                      </div>
                    </td>
                    <td className="border-gray-300 px-4 py-2 text-center">
                      <button
                        className=""
                        onClick={() => remove(index)} // 행 삭제
                      >
                        <TrashIcon className="w-6 h-6" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
        <DialogFooter className="flex justify-between items-center gap-2">
          {/* Add a new mapping 버튼을 왼쪽 끝에 위치 */}
          <Button variant="outline" size="sm" className="px-3 py-1" onClick={handleAddMapping}>
            + Add a new mapping
          </Button>
          {/* Cancel 및 Update 버튼을 오른쪽으로 밀기 */}
          <div className="flex gap-2 ml-auto">
            <Button
              variant="outline"
              size="sm"
              className="px-3 py-1"
              onClick={() => {
                remove();
                // setValue('mappings', []); // 배열 초기화
                onOpenChange(false); // 다이얼로그 닫기
              }}
            >
              Cancel
            </Button>
            <Button size="sm" className="px-3 py-1" onClick={handleSave}>
              Update
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

{
  /*  */
}
