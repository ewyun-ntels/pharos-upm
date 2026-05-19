import React, { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button, Input, Textarea } from '@pharos/shared/components/ui';
import { ToggleIconButton, Title } from '@pharos/shared/components/ui-extension';
import { Star } from '@pharos/shared/components';
import { useDashboardData } from '@features/dashboard/hooks/use-dashboard-store';
import { useHandleGeneralSave } from '@features/dashboard/hooks/use-handle-general-save';
import { useToast } from '@hooks/use-toast';

export function GeneralSettingTab() {
  const navigate = useNavigate();
  const { toast } = useToast();
  const { dashboard, id: dashboardId } = useDashboardData();
  const { handleGeneralSave } = useHandleGeneralSave();

  const [title, setTitle] = useState<string>(dashboard?.title || '');
  const [displayName, setDisplayName] = useState<string>(dashboard?.displayName || '');
  const [description, setDescription] = useState<string>(dashboard?.description || '');
  const [favorite, setFavorite] = useState<boolean>(dashboard?.favorite || false);

  // dashboard가 처음 로드될 때만 form 초기화
  const initialized = useRef(false);
  useEffect(() => {
    if (dashboard && !initialized.current) {
      initialized.current = true;
      setTitle(dashboard.title || '');
      setDisplayName(dashboard.displayName || '');
      setDescription(dashboard.description || '');
      setFavorite(dashboard.favorite || false);
    }
  }, [dashboard]);

  // hasChanges를 state 대신 직접 파생
  const hasChanges =
    title !== (dashboard?.title || '') ||
    displayName !== (dashboard?.displayName || '') ||
    description !== (dashboard?.description || '') ||
    favorite !== (dashboard?.favorite || false);

  const handleSave = async () => {
    if (title === '') {
      toast({ description: 'Title is required.', variant: 'destructive' });
      return;
    }
    if (!dashboardId) return;

    try {
      await handleGeneralSave(dashboardId, { title, displayName, description, favorite });
      // 저장 후 로컬 state를 저장된 값으로 동기화
      initialized.current = false;
    } catch {
      // handleGeneralSave 내부에서 toast 처리
    }
  };


  return (
    <div className="px-5 py-3">
      <div className="max-w-3xl space-y-6">
        <div>
          <Title variant="formLabel">
            Title <span className="text-red-500">*</span>
          </Title>
          <Input
            id="title"
            type="text"
            className="px-2 py-1"
            value={title}
            aria-label="Title"
            onChange={(e) => setTitle(e.target.value)}
          />
          {title === '' && (
            <div className="text-sm text-red-500 mt-1">Title을 입력해주세요.</div>
          )}
        </div>

        <div>
          <Title variant="formLabel">
            Display name
          </Title>
          <Input
            id="displayName"
            type="text"
            className="px-2 py-1"
            value={displayName}
            aria-label="Display Name"
            onChange={(e) => setDisplayName(e.target.value)}
          />
        </div>

        <div>
          <Title variant="formLabel">
            Description
          </Title>
          <Textarea
            id="description"
            className="px-2 py-1 shadow-none"
            value={description}
            rows={3}
            aria-label="Description"
            onChange={(e) => setDescription(e.target.value)}
          />
        </div>

        <div>
          <Title variant="formLabel">
            Favorite
          </Title>
          <ToggleIconButton
            icon={<Star />}
            tooltip="Favorite"
            toggledClassName="bg-amber-100 text-amber-600 hover:bg-amber-200"
            unToggledClassName="text-foreground"
            toggled={favorite}
            onClick={() => setFavorite(!favorite)}
          />
        </div>

        <div className="flex gap-2 pt-4">
          <Button
            onClick={() => navigate(-1)}
            size="default"
            variant="outline"
          >
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            size="default"
            variant="default"
            disabled={!hasChanges || title === ''}
          >
            Save
          </Button>
        </div>
      </div>
    </div>
  );
}
