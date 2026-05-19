import React, { useState, useRef, useMemo } from 'react';
import { Button } from '@pharos/shared/components/ui';
import { LoadingIndicator } from '@pharos/shared/components/ui-extension';
import { Alert, AlertTitle } from '@pharos/shared/components/ui';
import { TriangleAlert } from '@pharos/shared/components';
import { useHasPermission } from '@pharos/shared/features/auth';
import { PermissionKeysSchema } from '@pharos/shared/types/role';
import type { User } from '@pharos/shared/types/user';
import { useAllRoles } from '@features/user';
import { useUserMetadataConfig } from '@features/user/hooks';
import { useUserFormSchema } from '@features/user/hooks';
import { useUserActions } from '@features/user/hooks';
import { UserFormAndPermissions } from '../common/UserFormAndPermissions';

interface UserEditEditorProps {
  data: User;
  onSuccess: () => void;
  onCancel: () => void;
}

export function UserEditEditor({ data, onSuccess, onCancel }: UserEditEditorProps) {
  const [errMsg, setErrMsg] = useState('');
  const [saveDisabled, setSaveDisabled] = useState(false);
  const formRef = useRef<any>(null);

  const isSuperAdmin = useHasPermission(PermissionKeysSchema.enum['role:super_admin']);
  const { data: roleGroups, isLoading: isLoadingRoles, error: roleError } = useAllRoles();
  const { userMetadataConfig, isLoading: isLoadingUserMetadata } = useUserMetadataConfig();
  const { schema, uiSchema } = useUserFormSchema({
    roleGroups,
    isSuperAdmin,
    userMetadataConfig,
    isLoadingRoles,
    isLoadingUserMetadata,
  });

  // permissions — user 데이터에서 초기화
  const [permissions, setPermissions] = useState<Record<string, boolean>>(() => {
    if (data.roles && Array.isArray(data.roles) && data.roles.length > 0) {
      return data.roles.reduce((acc: Record<string, boolean>, role: any) => {
        return { ...acc, [role.role]: true };
      }, {});
    }
    const roles = data.attributes?.roles ?? {};
    return Object.keys(roles).reduce((acc: Record<string, boolean>, key) => {
      if (roles[key]) return { ...acc, [key]: true };
      return acc;
    }, {});
  });

  const initialUserInfo = useMemo(() => data.attributes?.info ?? {}, [data]);

  const { handleUpdateUser } = useUserActions({ onOpenChange: onSuccess, setErrMsg });

  const handleSave = (rjsfFormData: any) => {
    handleUpdateUser({
      username: data.name,
      userinfo: rjsfFormData.userinfo ?? {},
      permissions,
    });
  };

  return (
    <div>
      {isLoadingRoles || isLoadingUserMetadata ? (
        <LoadingIndicator className="h-40" />
      ) : roleError ? (
        <Alert className="text-destructive border border-destructive mb-4">
          <TriangleAlert className="mr-2 shrink-0" />
          <AlertTitle>Failed to load roles: {roleError.message}</AlertTitle>
        </Alert>
      ) : (
        <UserFormAndPermissions
          formRef={formRef}
          schema={schema}
          uiSchema={uiSchema}
          handleSave={handleSave}
          setSaveDisabled={setSaveDisabled}
          userinfo={initialUserInfo}
          permissions={permissions}
          setPermissions={setPermissions}
          openType="edit"
          isSuperAdmin={isSuperAdmin}
          roleGroups={roleGroups!}
        />
      )}

      {errMsg && (
        <Alert className="text-destructive border border-destructive mb-4">
          <TriangleAlert className="mr-2 shrink-0" />
          <AlertTitle className="line-clamp-none whitespace-normal break-words">
            <p className="lowercase first-letter:uppercase after:content-['.']">{errMsg}</p>
          </AlertTitle>
        </Alert>
      )}

      <div className="flex justify-start gap-2 mt-6">
        <Button onClick={onCancel} size="default" variant="outline" type="button">
          Cancel
        </Button>
        <Button
          size="default"
          variant="default"
          disabled={saveDisabled}
          onClick={() => {
            setErrMsg('');
            if (formRef.current) {
              formRef.current.submit();
            }
          }}
        >
          Save
        </Button>
      </div>
    </div>
  );
}
