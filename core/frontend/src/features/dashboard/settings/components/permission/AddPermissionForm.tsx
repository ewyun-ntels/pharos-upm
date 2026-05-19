import React, { useState } from 'react';
import { SelectBox } from '@pharos/shared/components/ui-extension/select/box';
import { Button } from '@pharos/shared/components/ui';
import { USER_PROVIDER_NAME, USER_RESOURCES } from '@providers/user-provider';
import { useList } from '@/lib/data-provider';
import { X } from '@pharos/shared/components';
import { ActionSchema } from '@pharos/shared/types/dashboard';

interface PermissionData {
  subject: string;
  object?: string;
  action: string;
}

interface UserData {
  name: string;
  [key: string]: any;
}

const PERMISSION_OPTIONS = [
  { value: ActionSchema.enum.viewer, label: 'Viewer' },
  { value: ActionSchema.enum.editor, label: 'Editor' },
  { value: ActionSchema.enum.owner, label: 'Owner' },
];

interface AddPermissionFormProps {
  permissionList: PermissionData[];
  isPending: boolean;
  onAdd: (data: PermissionData) => void;
  onCancel: () => void;
}

export function AddPermissionForm({ permissionList, isPending, onAdd, onCancel }: AddPermissionFormProps) {
  const [selectedType, setSelectedType] = useState<string | undefined>(undefined);
  const [formData, setFormData] = useState<PermissionData>({ subject: '', action: '' });

  const {
    query: { data: userList },
  } = useList({
    dataProviderName: USER_PROVIDER_NAME,
    resource: USER_RESOURCES.USER,
    queryOptions: { retry: false },
  });

  const handleAdd = () => {
    onAdd(formData);
    setFormData((prev) => ({ ...prev, subject: '' }));
  };

  return (
    <div className="mb-3">
      <div className="grid grid-cols-[200px_400px_250px_auto] gap-4 items-center">
        <div>
          <SelectBox
            size="full"
            options={[{ value: 'user', label: 'User' }]}
            placeholder="Type"
            value={selectedType}
            onChange={(value) => setSelectedType(value)}
            autoSelectFirstOption={false}
          />
        </div>
        <div>
          {userList?.data && (
            <SelectBox
              key={formData.subject || 'empty-user'}
              size="full"
              placeholder="User"
              options={(userList.data as UserData[])
                .filter((item) => !!item.name && !permissionList.some((p) => p.subject === item.name))
                .map((item) => ({ value: item.name, label: item.name }))}
              onChange={(value) => {
                if (value !== formData.subject) {
                  setFormData((prev) => ({ ...prev, subject: value }));
                }
              }}
              value={formData.subject || undefined}
              autoSelectFirstOption={false}
            />
          )}
        </div>
        <div>
          <SelectBox
            size="full"
            placeholder="Permission"
            options={PERMISSION_OPTIONS}
            value={formData.action || undefined}
            onChange={(value: string) => setFormData((prev) => ({ ...prev, action: value }))}
            autoSelectFirstOption={false}
          />
        </div>
        <div className="flex gap-2">
          <Button
            onClick={handleAdd}
            disabled={!selectedType || !formData.subject || !formData.action || isPending}
          >
            Add
          </Button>
          <Button variant="ghost" onClick={onCancel} className='w-8'>
            <X className="w-4 h-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
