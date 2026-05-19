import { z } from 'zod';
import type { PasswordPolicy, UserPolicy } from '@pharos/shared/types/auth';

// create 전용 스키마 — POST /user 에 대응
export function getUserCreateSchema(
  usernameRequirements: UserPolicy & { requirementMsg?: string },
  passwordRequirements: PasswordPolicy & { requirementMsg?: string },
  usernameValidationTxt: string,
  passwordValidationTxt: string,
) {
  const usernameSchema = usernameRequirements.email_allowed
    ? z
        .string()
        .min(usernameRequirements.min_length, usernameValidationTxt)
        .max(usernameRequirements.max_length, usernameValidationTxt)
        .regex(/^([a-z0-9_]+|[^\s@]+@[^\s@]+\.[^\s@]+)$/, usernameValidationTxt)
    : z
        .string()
        .min(usernameRequirements.min_length, usernameValidationTxt)
        .max(usernameRequirements.max_length, usernameValidationTxt)
        .regex(/^[a-z0-9_]+$/, usernameValidationTxt)
        .refine((val) => !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(val), {
          message: 'Email format is not allowed.',
          path: ['username'],
        });

  return z
    .object({
      username: usernameSchema,
      password: z.string().min(passwordRequirements.min_length, passwordValidationTxt),
      confirmPassword: z.string(),
    })
    .refine(
      (data) => {
        const pw = data.password;
        if (!pw) return true;
        return (
          (!passwordRequirements.min_uppercase || (pw.match(/[A-Z]/g)?.length ?? 0) >= passwordRequirements.min_uppercase) &&
          (!passwordRequirements.min_lowercase || (pw.match(/[a-z]/g)?.length ?? 0) >= passwordRequirements.min_lowercase) &&
          (!passwordRequirements.min_digits    || (pw.match(/[0-9]/g)?.length ?? 0) >= passwordRequirements.min_digits) &&
          (!passwordRequirements.min_special   || (pw.match(/[^A-Za-z0-9]/g)?.length ?? 0) >= passwordRequirements.min_special)
        );
      },
      { message: passwordValidationTxt, path: ['password'] },
    )
    .refine((data) => data.password === data.confirmPassword, {
      message: 'Passwords do not match.',
      path: ['confirmPassword'],
    });
}
