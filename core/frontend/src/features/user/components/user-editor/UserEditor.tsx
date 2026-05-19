import React from 'react';
import type { User } from '@pharos/shared/types/user';
import { UserCreateEditor } from './create/UserCreateEditor';
import { UserEditEditor } from './edit/UserEditEditor';

interface UserEditorProps {
  openType: 'create' | 'edit';
  data?: User;
  onSuccess: () => void;
  onCancel: () => void;
}

export function UserEditor({ openType, data, onSuccess, onCancel }: UserEditorProps) {
  if (openType === 'create') {
    return <UserCreateEditor onSuccess={onSuccess} onCancel={onCancel} />;
  }
  return <UserEditEditor data={data!} onSuccess={onSuccess} onCancel={onCancel} />;
}
