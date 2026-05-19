import { useToast } from '@hooks/use-toast';
import { useForm, useUpdate, useInvalidate } from '@/lib/data-provider';
import { useCallback } from 'react';
import { USER_PROVIDER_NAME, USER_RESOURCES } from '@providers/user-provider';
import type {
  CreateUserRequest,
  SetAttributesRequest,
} from '@pharos/shared/types/user';

interface UseUserActionsProps {
  userListRefetch?: () => void;
  onOpenChange: () => void;
  setErrMsg: React.Dispatch<React.SetStateAction<string>>;
}

export function useUserActions({
  userListRefetch,
  onOpenChange,
  setErrMsg,
}: UseUserActionsProps) {
  const { toast } = useToast();
  const invalidate = useInvalidate();
  const { onFinish: onFinishNew } = useForm({
    action: 'create',
    resource: USER_RESOURCES.USER,
    dataProviderName: USER_PROVIDER_NAME,
    redirect: false,
  });
  const { mutate: updateMutate } = useUpdate();

  const handleAddUser = useCallback(
    (formData: any) => {
      const roles: Record<string, boolean> = {};
      if (formData.permissions) {
        Object.entries(formData.permissions).forEach(([key, value]) => {
          if (value === true && key !== 'user') {
            roles[key] = true;
          }
        });
      }

      const info: Record<string, any> =
        formData.userinfo && Object.keys(formData.userinfo).length > 0 ? formData.userinfo : {};

      const userData: CreateUserRequest = {
        username: formData.username,
        password: formData.password,
        attributes: { roles, info },
      };

      onFinishNew(userData)
        .then(async (result) => {
          setErrMsg('');
          if (result) {
            toast({ description: 'User information has been registered.' });
            await invalidate({
              resource: USER_RESOURCES.USER,
              dataProviderName: USER_PROVIDER_NAME,
              invalidates: ['list'],
            });
            userListRefetch?.();
          } else {
            toast({ description: 'Failed to register user.' });
          }
          onOpenChange();
        })
        .catch((error) => {
          const err = error as any;
          const errorMsg = err?.response?.data?.error ?? err?.message ?? 'An error occurred';
          setErrMsg(errorMsg);
        });
    },
    [onFinishNew, toast, invalidate, userListRefetch, onOpenChange, setErrMsg],
  );

  const handleUpdateUser = useCallback(
    (formData: any) => {
      const roles: Record<string, boolean> = {};
      if (formData.permissions) {
        Object.entries(formData.permissions).forEach(([key, value]) => {
          if (value === true && key !== 'user') {
            roles[key] = true;
          }
        });
      }

      const info: Record<string, any> =
        formData.userinfo && Object.keys(formData.userinfo).length > 0 ? formData.userinfo : {};

      const attributesData: SetAttributesRequest = {
        attributes: { roles, info },
      };

      updateMutate(
        {
          resource: 'user-attributes',
          id: formData.username,
          dataProviderName: USER_PROVIDER_NAME,
          values: attributesData,
        },
        {
          onSuccess: () => {
            toast({ description: 'User information has been updated.' });
            userListRefetch?.();
            onOpenChange();
          },
          onError: (error: unknown) => {
            const err = error as any;
            const errorMsg =
              err?.response?.data?.error ?? err?.message ?? 'Failed to update user information';
            setErrMsg(errorMsg);
          },
        },
      );
    },
    [updateMutate, toast, userListRefetch, onOpenChange, setErrMsg],
  );

  return { handleAddUser, handleUpdateUser };
}
