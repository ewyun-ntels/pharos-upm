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

interface BadgeData {
  [condition: string]: {
    displayText?: string;
    color: string;
    textColor?: string;
  };
}

interface DialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  badgeData?: BadgeData;
  onSave: (newBadgeData: BadgeData) => void;
}

export const MappingDialog = ({open, onOpenChange, badgeData, onSave}: DialogProps) => {
  const {control, register, getValues, setValue, trigger, formState} = useForm();

  // useFieldArray를 사용하여 행 관리
  const {fields, append, remove} = useFieldArray({
    control,
    name: `mappings`, // Field Array의 이름
  });

  // 다이얼로그가 열릴 때 배열이 비어 있으면 기본 행 추가
  useEffect(() => {
    if (open) {
      remove(); // 다이얼로그가 열릴 때 기존 행 제거
      if (!badgeData) {
        append({condition: '', displayText: '', color: '#ffffff', textColor: '#ffffff'}); // 기본 행 추가
      } else {
        const reverseData = reverseTransformMappingData(badgeData);
        append(reverseData); // 기존 데이터로 행 추가
      }
    } else {
      // 다이얼로그가 닫힐 때 필드 초기화
      remove(); // 모든 행 제거
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  const handleAddMapping = () => {
    append({condition: '', displayText: '', color: '#ffffff'}); // 새 행 추가
  };

  interface MappingItem {
    condition: string;
    displayText?: string;
    color: string;
    textColor?: string;
  }

  interface TransformedMappingData {
    [condition: string]: {
      displayText?: string;
      color: string;
      textColor?: string;
    };
  }

  const transformMappingData = (data: MappingItem[]): TransformedMappingData => {
    return data.reduce(
      (acc: TransformedMappingData, {condition, color, displayText, textColor}) => {
        acc[condition] = {color, displayText, textColor};
        return acc;
      },
      {},
    );
  };

  const reverseTransformMappingData = (data: TransformedMappingData): MappingItem[] => {
    return Object.entries(data).map(([condition, value]) => ({
      condition,
      displayText: value?.displayText || '',
      color: value.color,
      textColor: value.textColor || '#ffffff',
    }));
  };

  const handleSave = async () => {
    const isValid = await trigger('mappings'); // 'mappings' 필드의 유효성 검사 수행
    if (!isValid) {
      return; // 유효성 검사 실패 시 저장 로직 중단
    }

    const mappings = getValues('mappings'); // 현재 mappings 값 가져오기
    const transformedData = transformMappingData(mappings); // 필요한 경우 데이터 변환

    onSave(transformedData); // props를 통해 부모에게 전달
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
          <DialogDescription>
            Map exact cell values to a colored badge. <strong>Condition</strong> is the exact value in the cell (e.g. <code>active</code>, <code>error</code>, <code>1</code>). <strong>Display text</strong> overrides the label shown inside the badge. <strong>Color</strong> sets the badge background; <strong>Text Color</strong> overrides the text color.
          </DialogDescription>
        </DialogHeader>
        <div className="p-4">
          <div className="mb-4">
            <table className="w-full">
              <thead>
                <tr>
                  <th className="px-4 py-2 text-left text-sm font-normal">
                    Condition
                    <span className="ml-1 text-[11px] text-muted-foreground font-normal">(exact cell value)</span>
                  </th>
                  <th className="px-4 py-2 text-left text-sm font-normal">
                    Display text
                    <span className="ml-1 text-[11px] text-muted-foreground font-normal">(optional)</span>
                  </th>
                  <th className="px-4 py-2 text-center text-sm font-normal">Badge color</th>
                  <th className="px-4 py-2 text-center text-sm font-normal">Text color</th>
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
                          formState.errors?.mappings[index]?.condition
                            ? 'border-red-500' // 유효성 검사 실패 시 빨간 테두리
                            : 'border-gray-300' // 기본 테두리
                        } px-4 py-2`}
                        {...register(`mappings.${index}.condition`, {
                          required: 'Condition is required',
                        })}
                        placeholder="Exact value to match"
                      />
                    </td>
                    <td className="border-gray-300 px-4 py-2">
                      <Input
                        {...register(`mappings.${index}.displayText`)}
                        placeholder="Optional display text"
                      />
                    </td>
                    <td className="border-gray-300 px-4 py-2 text-center">
                      <div className="flex justify-center items-center">
                        <Input
                          type="color"
                          className="w-10 h-10 p-0 border-none cursor-pointer"
                          {...register(`mappings.${index}.color`, {
                            onBlur(e) {
                              setValue(`mappings.${index}.color`, e.target.value);
                              setValue(`mappings.${index}.textColor`, undefined);
                            },
                          })}
                        />
                      </div>
                    </td>
                    <td className="border-gray-300 px-4 py-2 text-center">
                      <div className="flex justify-center items-center">
                        <Input
                          type="color"
                          className="w-10 h-10 p-0 border-none cursor-pointer"
                          {...register(`mappings.${index}.textColor`, {
                            onBlur(e) {
                              setValue(`mappings.${index}.textColor`, e.target.value);
                              setValue(`mappings.${index}.color`, undefined);
                            },
                          })}
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
