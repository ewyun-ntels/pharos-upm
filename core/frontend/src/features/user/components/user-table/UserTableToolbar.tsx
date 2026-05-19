import React from 'react';
import {Button} from '@pharos/shared/components/ui';
import {IconButton} from '@pharos/shared/components/ui-extension';
import {Plus, RefreshCcw} from '@pharos/shared/components';
import {useHasPermission} from '@pharos/shared/features/auth';
import {PermissionKeysSchema} from '@pharos/shared/types/role';

type UsersSettingsProps = {
  selectedUsers: string[];
  setShowAlert: (show: boolean) => void;
  userListRefetch: () => void;
  handelEditDialogOpen: (type: 'create' | 'edit', rowData?: any) => void;
};

export function UsersSettings({
  selectedUsers,
  setShowAlert,
  userListRefetch,
  handelEditDialogOpen,
}: UsersSettingsProps) {
  // 권한 체크
  const hasUserCreate = useHasPermission(PermissionKeysSchema.enum['role:user_create']);
  const hasUserDelete = useHasPermission(PermissionKeysSchema.enum['role:user_delete']);

  return (
    <>
      {hasUserDelete && (
        <Button
          size="default"
          variant="destructive"
          className="h-8 shadow-none px-3 py-1 cursor-pointer disabled:opacity-0"
          onClick={() => setShowAlert(true)}
          disabled={selectedUsers.length === 0}
        >
          Delete {selectedUsers?.length > 0 ? `${selectedUsers.length}` : ''}
        </Button>
      )}
      <IconButton
        type="button"
        variant="outline"
        icon={<RefreshCcw />}
        onClick={() => {
          userListRefetch();
        }}
      ></IconButton>
      {hasUserCreate && (
        <Button
          size="default"
          className="h-8 shadow-none px-3 py-1 cursor-pointer"
          onClick={() => handelEditDialogOpen('create')}
        >
          <Plus />
          New user
        </Button>
      )}
    </>
  );
}
