import { useMemo } from 'react';
import { useOne } from '@/lib/data-provider';
import { USER_PROVIDER_NAME, USER_RESOURCES } from '@providers/user-provider';
import { UserMetadataConfigSchema } from '@pharos/shared/types/user';
import type { UserMetadataConfig } from '@pharos/shared/types/user';

export function useUserMetadataConfig(): { userMetadataConfig: UserMetadataConfig | null; isLoading: boolean } {
  const {
    query: { data: userMetadataConfigData, isLoading },
  } = useOne({
    resource: USER_RESOURCES.USER_METADATA_CONFIG,
    id: 'active',
    dataProviderName: USER_PROVIDER_NAME,
    queryOptions: {
      retry: 1,
      staleTime: 5 * 60 * 1000,
    },
  });

  const userMetadataConfig = useMemo((): UserMetadataConfig | null => {
    if (!userMetadataConfigData?.data) return null;
    const rawData = userMetadataConfigData.data.data || userMetadataConfigData.data;
    if (!rawData.schema) return null;
    const result = UserMetadataConfigSchema.safeParse(rawData);
    if (!result.success) {
      console.error('User metadata config validation failed:', result.error);
      return null;
    }
    return result.data;
  }, [userMetadataConfigData]);

  return { userMetadataConfig, isLoading };
}
