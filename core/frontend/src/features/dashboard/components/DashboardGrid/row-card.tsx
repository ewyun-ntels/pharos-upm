import {useState} from 'react';
import {ChevronDown, ChevronRight, Settings, Trash2} from '@pharos/shared/components';
import {IconButton} from '@pharos/shared/components/ui-extension';
import {Panel} from '@pharos/shared/types/dashboard';
import {Dialog, DialogContent, DialogHeader, DialogTitle} from '@pharos/shared/components/ui';
import {Button} from '@pharos/shared/components/ui';

interface RowCardProps {
  rowData: Panel;
  isOpen?: boolean;
  toggleRow: (id: string) => void;
  removeBox: (id: string) => void;
  updateRowTitle: (id: string, title: string) => void;
  panelCount?: number; // Row 안의 패널 개수
}

export const RowCard: React.FC<RowCardProps> = ({
  rowData,
  isOpen = false,
  toggleRow,
  removeBox,
  updateRowTitle,
  panelCount = 0,
}: RowCardProps) => {
  const [isDialog, setIsDialog] = useState(false);
  const [inputValue, setInputValue] = useState(rowData.title);

  const onOpenChange = (open: boolean) => {
    setIsDialog(open);
  };

  return (
    <>
      <div className="flex items-center h-full text-muted-foreground text-sm px-2">
        <button onClick={() => toggleRow(rowData.id)} className="flex items-center p-2 gap-2">
          {isOpen ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
          <div className="flex items-center gap-1">
            <span className="text-sm">{rowData.title}</span>
            {panelCount > 0 && (
              <span className="text-xs text-muted-foreground">
                ({panelCount} {panelCount === 1 ? 'item' : 'items'})
              </span>
            )}
          </div>
        </button>
        <IconButton
          onClick={() => setIsDialog(true)}
          variant="ghost"
          icon={<Settings size={16} />}
        />
        <IconButton
          onClick={() => removeBox(rowData.id)}
          variant="ghost"
          icon={<Trash2 size={16} />}
        />
      </div>
      <Dialog open={isDialog} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-4xl">
          <DialogHeader>
            <DialogTitle>Row Options</DialogTitle>
          </DialogHeader>
          <div className="mt-4">
            <div className="flex items-center gap-2">
              <label htmlFor="row-name" className="text-sm font-medium text-muted-foreground">
                Row Name:
              </label>
              <input
                id="row-name"
                type="text"
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value)}
                className="border rounded px-2 py-1"
              />
            </div>
          </div>
          <div className="flex gap-2 justify-end">
            <Button variant="outline" size="sm" onClick={() => setIsDialog(false)}>
              취소
            </Button>
            <Button
              size="sm"
              onClick={() => {
                updateRowTitle(rowData.id, inputValue);
                setIsDialog(false);
              }}
            >
              저장
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
};
