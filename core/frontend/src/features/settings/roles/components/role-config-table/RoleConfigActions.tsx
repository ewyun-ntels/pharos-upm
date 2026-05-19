import React from 'react';
import {Button} from '@pharos/shared/components/ui';

type RoleConfigActionsProps = {
  isDirty: boolean;
  isSaving: boolean;
  isLoading: boolean;
  onSave: () => void;
};

export function RoleConfigActions({
  isDirty,
  isSaving,
  isLoading,
  onSave,
}: RoleConfigActionsProps) {
  return (
    <>
      <Button
        size="default"
        className="shadow-none px-3 cursor-pointer"
        onClick={onSave}
        disabled={!isDirty || isSaving || isLoading}
      >
        {isSaving ? 'Saving...' : 'Save'}
      </Button>
    </>
  );
}
