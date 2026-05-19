import { useState } from 'react';
import { useUpdate, useDelete, useOne } from '@/lib/data-provider';
import { useToast } from '@hooks/use-toast';
import { useSheetStore } from '@components/sheet/sheetStore';
import { USER_PROVIDER_NAME, USER_RESOURCES } from '@providers/user-provider';
import type { User } from '@pharos/shared/types/user';

export interface JsonSchemaProperty {
  title?: string;
  type?: string;
  description?: string;
  [key: string]: unknown;
}

export interface JsonSchema {
  type?: string;
  properties?: Record<string, JsonSchemaProperty>;
  [key: string]: unknown;
}

export function useUserSheet(user: User, refetch: () => void) {
  const { setClose } = useSheetStore();
  const { toast } = useToast();
  const { mutate: setBlocked } = useUpdate();
  const { mutateAsync: deleteUser } = useDelete();
  const [showDeleteAlert, setShowDeleteAlert] = useState(false);

  const {
    query: { data: freshUserData, refetch: refetchUser },
  } = useOne<User>({
    resource: USER_RESOURCES.USER,
    id: user.name,
    dataProviderName: USER_PROVIDER_NAME,
    queryOptions: {
      retry: (failureCount, error) => {
        const status = (error as { statusCode?: number })?.statusCode;
        if (status !== undefined && status >= 400 && status < 500) return false;
        return failureCount < 1;
      },
    },
  });

  const liveUser: User = (freshUserData?.data ?? user) as User;
  const isBlocked = liveUser.blocked ?? false;

  const {
    query: { data: metaConfigData },
  } = useOne({
    resource: USER_RESOURCES.USER_METADATA_CONFIG,
    id: 'active',
    dataProviderName: USER_PROVIDER_NAME,
    queryOptions: { retry: 1, staleTime: 5 * 60 * 1000 },
  });

  const rawMetaConfig = metaConfigData?.data?.data ?? metaConfigData?.data;
  const metaSchema = (rawMetaConfig?.schema ?? null) as JsonSchema | null;

  const userInfoEntries = Object.entries(liveUser.attributes?.info ?? {}).filter(
    ([, v]) => v != null,
  ) as [string, string | number | boolean][];

  const handleBlockToggle = () => {
    setBlocked(
      {
        resource: isBlocked ? 'user-unblock' : 'user-block',
        id: liveUser.name,
        dataProviderName: USER_PROVIDER_NAME,
        values: { block: !isBlocked },
      },
      {
        onSuccess: () => {
          refetchUser();
          refetch();
          toast({ description: `User ${isBlocked ? 'unblocked' : 'blocked'} successfully.` });
        },
        onError: () => {
          toast({ description: 'Failed to update user block status.' });
        },
      },
    );
  };

  const handleDeleteConfirm = async () => {
    try {
      await deleteUser({
        resource: USER_RESOURCES.USER,
        id: liveUser.name,
        dataProviderName: USER_PROVIDER_NAME,
      });
      toast({ description: 'User has been deleted.' });
      setClose();
      refetch();
    } catch {
      toast({ description: 'Failed to delete user.' });
    }
  };

  return {
    liveUser,
    setClose,
    isBlocked,
    showDeleteAlert,
    setShowDeleteAlert,
    metaSchema,
    userInfoEntries,
    handleBlockToggle,
    handleDeleteConfirm,
  };
}
