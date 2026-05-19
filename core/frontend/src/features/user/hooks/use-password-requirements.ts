import { useOne } from '@/lib/data-provider';
import { USER_PROVIDER_NAME } from '@providers/user-provider';
import { AuthConfigSchema } from '@pharos/shared/types/auth';
import type { PasswordPolicy } from '@pharos/shared/types/auth';

const PW_LABELS: Record<string, (v: number) => string> = {
  min_length: (v) => `At least ${v} character${v > 1 ? 's' : ''}`,
  min_uppercase: (v) => `${v} uppercase letter${v > 1 ? 's' : ''}`,
  min_lowercase: (v) => `${v} lowercase letter${v > 1 ? 's' : ''}`,
  min_digits: (v) => `${v} digit${v > 1 ? 's' : ''}`,
  min_special: (v) => `${v} special character${v > 1 ? 's' : ''} (such as !, @, #, etc.)`,
};

export function usePasswordRequirements() {
  const { query: { data: pwRequires } } = useOne({
    dataProviderName: USER_PROVIDER_NAME,
    resource: `auth`,
    id: 'config',
    queryOptions: { retry: false },
  });

  const parsed = AuthConfigSchema.safeParse(pwRequires?.data);
  const pw: PasswordPolicy | undefined = parsed.success ? parsed.data.password : undefined;

  let requirementMsg;
  let config: PasswordPolicy = {
    min_length: 8,
    min_uppercase: 2,
    min_lowercase: 2,
    min_digits: 2,
    min_special: 2,
  };

  if (!pw) {
    requirementMsg =
      'At least 8 characters, include at least 2 uppercase letters, 2 lowercase letters, 2 digits, 2 special characters (such as !, @, #, etc.)';
  } else {
    config = pw;
    const requirements = Object.entries(PW_LABELS)
      .filter(([key]) => Number(pw[key as keyof PasswordPolicy]) > 0)
      .map(([key, labelFn]) => labelFn(Number(pw[key as keyof PasswordPolicy])));

    requirementMsg =
      requirements.length === 0
        ? 'No password requirements.'
        : requirements.length === 1
          ? requirements[0] + '.'
          : requirements[0] + ', include at least ' + requirements.slice(1).join(', ') + '.';
  }

  return {
    min_length: config.min_length,
    min_uppercase: config.min_uppercase,
    min_lowercase: config.min_lowercase,
    min_digits: config.min_digits,
    min_special: config.min_special,
    requirementMsg,
  };
}
