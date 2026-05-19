import { useOne } from '@/lib/data-provider';
import { USER_PROVIDER_NAME } from '@providers/user-provider';
import { AuthConfigSchema } from '@pharos/shared/types/auth';
import type { UserPolicy } from '@pharos/shared/types/auth';

export function useUsernameRequirements() {
  const { query: { data } } = useOne({
    dataProviderName: USER_PROVIDER_NAME,
    resource: `auth`,
    id: 'config',
    queryOptions: { retry: false },
  });

  const parsed = AuthConfigSchema.safeParse(data?.data);
  const username: UserPolicy | undefined = parsed.success ? parsed.data.user : undefined;

  let requirementMsg;
  let config: UserPolicy = {
    email_allowed: false,
    min_length: 4,
    max_length: 32,
  };

  if (!username) {
    requirementMsg = 'Use letters, numbers, "@", and "." only. Must be 4–32 characters long.';
  } else {
    config = username;

    const charInfo = username?.email_allowed
      ? 'Use letters, numbers, "@", and "." only.'
      : 'Use lowercase letters, numbers, and "_" only.';

    const lengthInfo = `Must be ${username?.min_length}–${username?.max_length} characters long.`;

    requirementMsg = charInfo + ' ' + lengthInfo;
  }

  return {
    email_allowed: config.email_allowed,
    min_length: config.min_length,
    max_length: config.max_length,
    requirementMsg,
  };
}
