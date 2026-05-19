import React, { useState } from 'react';
import { Button } from '@pharos/shared/components/ui';
import { LoadingIndicator } from '@pharos/shared/components/ui-extension';
import { Alert, AlertTitle } from '@pharos/shared/components/ui';
import { TriangleAlert } from '@pharos/shared/components';
import { useHasPermission } from '@pharos/shared/features/auth';
import { PermissionKeysSchema } from '@pharos/shared/types/role';
import { useAllRoles } from '@features/user';
import { useUsernameRequirements } from '@features/user';
import { usePasswordRequirements } from '@features/user';
import { useUserMetadataConfig } from '@features/user/hooks';
import { useUserFormSchema } from '@features/user/hooks';
import { useUserCreateForm } from '@features/user/hooks';
import { useUserActions } from '@features/user/hooks';
import { UserCredentialsFields } from './UserCredentialsFields';
import { UserFormAndPermissions } from '../common/UserFormAndPermissions';

interface UserCreateEditorProps {
  onSuccess: () => void;
  onCancel: () => void;
}

export function UserCreateEditor({ onSuccess, onCancel }: UserCreateEditorProps) {
  const [errMsg, setErrMsg] = useState('');
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

  const usernameRequirements = useUsernameRequirements();
  const passwordRequirements = usePasswordRequirements();

  const {
    formRef,
    register,
    errors,
    setValue,
    trigger,
    getValues,
    clearErrors,
    saveDisabled,
    setSaveDisabled,
    permissions,
    setPermissions,
    manualErrors,
    resetForm,
  } = useUserCreateForm({
    usernameRequirements,
    usernameValidationTxt: 'Invalid username. Check the username rules.',
    passwordRequirements,
    passwordValidationTxt: 'Invalid password. Check the password rules.',
    setErrMsg,
  });

  const { handleAddUser } = useUserActions({ onOpenChange: onSuccess, setErrMsg });

  const handleCancel = () => {
    resetForm();
    onCancel();
  };

  const handleSave = (rjsfFormData: any) => {
    handleAddUser({
      ...getValues(),
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
        <>
          <UserCredentialsFields
            register={register}
            setValue={setValue}
            errors={errors}
            manualErrors={manualErrors}
            usernameRequirements={usernameRequirements}
            passwordRequirements={passwordRequirements}
            clearErrors={clearErrors as (field?: string | string[] | undefined) => void}
            getValues={getValues}
            trigger={(field: string) => trigger(field as any)}
          />
          <UserFormAndPermissions
            formRef={formRef}
            schema={schema}
            uiSchema={uiSchema}
            handleSave={handleSave}
            setSaveDisabled={setSaveDisabled}
            userinfo={{}}
            permissions={permissions}
            setPermissions={setPermissions}
            openType="create"
            isSuperAdmin={isSuperAdmin}
            roleGroups={roleGroups!}
          />
        </>
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
        <Button onClick={handleCancel} size="default" variant="outline" type="button">
          Cancel
        </Button>
        <Button
          size="default"
          variant="default"
          disabled={saveDisabled}
          onClick={async () => {
            setErrMsg('');
            const rhfValid = await trigger();
            if (rhfValid && formRef.current) {
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
