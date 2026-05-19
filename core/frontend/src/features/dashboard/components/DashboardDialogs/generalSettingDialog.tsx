import React, { useEffect, useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogDescription,
} from '@pharos/shared/components/ui';
import { ToggleIconButton } from '@pharos/shared/components/ui-extension';
import { Star } from '@pharos/shared/components';
import { Button } from '@pharos/shared/components/ui';
import { Input } from '@pharos/shared/components/ui';
import { Textarea } from '@pharos/shared/components/ui';

interface DialogProps {
  open: boolean;
  onOpenChange: () => void;
  dialogTitle: string;
  data: any;
  onSave: (id: string, newData: any) => void | Promise<void>;
}

export function GeneralSettingDialog({ dialogTitle, open, onOpenChange, data, onSave }: DialogProps) {
  const config = data?.config ?? data ?? {};
  const id = data?.id;
  const [saveDisabled, setSaveDisabled] = useState<boolean>(true);
  const [title, setTitle] = useState<string>(config?.title || '');
  const [displayName, setDisplayName] = useState<string>(config?.displayName || '');
  const [description, setDescription] = useState<string>(config?.description || '');
  const [favorite, setFavorite] = useState<boolean>(config?.favorite || false);

  const handleSettingDialogClose = () => {
    onOpenChange();
  };

  const handleSave = () => {
    if (onSave) {
      onSave(id, {
        ...config,
        title,
        displayName,
        description,
        favorite,
      });
      onOpenChange();
    }
  };

  useEffect(() => {
    const cfg = data?.config ?? data ?? {};
    if (data) {
      setTitle(cfg.title || '');
      setDisplayName(cfg.displayName || '');
      setDescription(cfg.description || '');
      setFavorite(cfg.favorite || false);
    }
  }, [data]);

  useEffect(() => {
    setSaveDisabled(title.length === 0);
  }, [title]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-none" style={{ maxWidth: '40rem' }}>
        {/* 타이틀 하단에 내용이 없어도 p태그 영역이 잡혀 조건부로 변경 */}
        <DialogHeader>
          <DialogTitle>{dialogTitle}</DialogTitle>
          {description && <DialogDescription>{description}</DialogDescription>}
        </DialogHeader>
        <div>
          <label htmlFor={'title'} className="block font-small mb-1">
            Title <span className="text-black-500">*</span>
          </label>
          <Input
            type="text"
            className="border border-border rounded px-2 py-1"
            value={title}
            aria-label="Title"
            onChange={(e) => {
              setTitle(e.target.value);
            }}
          />
          {title === '' && <div className="text-sm text-red-500 mt-1">Title을 입력해주세요.</div>}
        </div>
        <div>
          <label htmlFor={'displayName'} className="block font-small mb-1">
            Display Name
          </label>
          <Input
            type="text"
            className="border border-border rounded px-2 py-1"
            value={displayName}
            aria-label="Display Name"
            onChange={(e) => {
              setDisplayName(e.target.value);
            }}
          />
        </div>
        <div>
          <label htmlFor={'description'} className="block font-small mb-1">
            Description
          </label>
          <Textarea
            className="border border-border rounded px-2 py-1"
            defaultValue={description || ''}
            rows={3}
            aria-label="Description"
            onChange={(e) => {
              setDescription(e.target.value);
            }}
          />
        </div>
        <div>
          <label htmlFor={'favorite'} className="block font-small mb-1">
            Favorite
          </label>
          <ToggleIconButton
            icon={<Star />}
            tooltip="Favorite"
            toggledClassName="bg-amber-100 text-amber-600 hover:bg-amber-20"
            unToggledClassName="text-foreground"
            toggled={favorite || false}
            onClick={() => {
              setFavorite(!favorite);
            }}
          />
        </div>
        <DialogFooter>
          <Button
            onClick={() => {
              handleSettingDialogClose();
            }}
            size="sm"
            variant="outline"
            disabled={false}
          >
            Cancel
          </Button>
          <Button
            onClick={() => {
              if (title === '') {
                return;
              } else {
                handleSave();
              }
            }}
            size="sm"
            variant="default"
            disabled={saveDisabled}
          >
            Apply
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
