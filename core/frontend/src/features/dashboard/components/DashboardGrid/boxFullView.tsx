import {Dialog, DialogContent, DialogTitle} from '@pharos/shared/components/ui';
import {VisuallyHidden} from '@pharos/shared/components';

export function BoxFullView({
  open,
  onClose,
  children,
}: {
  open: boolean;
  onClose: () => void;
  children: React.ReactNode;
}) {
  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="!w-[90%] !max-w-full !h-[90vh] !max-h-[90vh] overflow-auto p-0">
        {/* 접근성을 위한 숨겨진 타이틀 없으면 에러발생 */}
        <VisuallyHidden>
          <DialogTitle>전체 화면 보기</DialogTitle>
        </VisuallyHidden>

        <div className="flex flex-col max-w-full !max-h-[800px]">{children}</div>
      </DialogContent>
    </Dialog>
  );
}
