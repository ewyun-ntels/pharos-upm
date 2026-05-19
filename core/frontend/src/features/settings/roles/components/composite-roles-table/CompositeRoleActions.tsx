import React from 'react';
import {Button} from '@pharos/shared/components/ui';
import {Plus} from 'lucide-react';

type CompositeRoleActionsProps = {
  isLoading: boolean;
  isSaving: boolean;
  onAdd: () => void;
};

export function CompositeRoleActions({isLoading, isSaving, onAdd}: CompositeRoleActionsProps) {
  return (
    <Button
      size="default"
      variant="default"
      className="shadow-none px-3 cursor-pointer"
      onClick={onAdd}
      disabled={isLoading || isSaving}
    >
      <Plus className="h-3.5 w-3.5" />
      Add composite role
    </Button>
  );
}
