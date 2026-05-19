import { useRef, useEffect, useState, useMemo } from 'react';
import { useForm as useReactHookForm } from 'react-hook-form';
import { getUserCreateSchema } from '@features/user/schemas';
import { zodResolver } from '@hookform/resolvers/zod';
import type { PasswordPolicy, UserPolicy } from '@pharos/shared/types/auth';

// create 전용 — POST /user
type FormValues = {
  username: string;
  password: string;
  confirmPassword: string;
};

export function useUserCreateForm({
  usernameRequirements,
  usernameValidationTxt,
  passwordRequirements,
  passwordValidationTxt,
  setErrMsg,
}: {
  usernameRequirements: UserPolicy & { requirementMsg?: string };
  usernameValidationTxt: string;
  passwordRequirements?: PasswordPolicy & { requirementMsg?: string };
  passwordValidationTxt?: string;
  setErrMsg: (msg: string) => void;
}) {
  const formRef = useRef<any>(null);
  const [saveDisabled, setSaveDisabled] = useState(true);
  const [permissions, setPermissions] = useState<Record<string, boolean>>({});
  const [manualErrors, setManualErrors] = useState<{
    username?: string;
    password?: string;
    confirmPassword?: string;
  }>({});

  const schema = useMemo(
    () =>
      getUserCreateSchema(
        usernameRequirements,
        passwordRequirements ?? {
          min_length: 1,
          min_uppercase: 0,
          min_lowercase: 0,
          min_digits: 0,
          min_special: 0,
        },
        usernameValidationTxt,
        passwordValidationTxt ?? '',
      ),
    [usernameRequirements, passwordRequirements, usernameValidationTxt, passwordValidationTxt],
  );

  const {
    register,
    formState: { errors },
    reset,
    watch,
    setValue,
    trigger,
    getValues,
    clearErrors,
  } = useReactHookForm<FormValues>({
    resolver: zodResolver(schema),
    mode: 'onChange',
    defaultValues: { username: '', password: '', confirmPassword: '' },
  });

  const watchedUsername = watch('username');
  const watchedPassword = watch('password');
  const watchedConfirmPassword = watch('confirmPassword');

  // 모든 필드 입력 여부 + validation 실시간 반영
  useEffect(() => {
    const username = watchedUsername;
    const password = watchedPassword;
    const confirmPassword = watchedConfirmPassword;
    const isEmpty = !username?.trim() || !password?.trim() || !confirmPassword?.trim();
    setSaveDisabled(isEmpty);

    const result = schema.safeParse({
      username: username ?? '',
      password: password ?? '',
      confirmPassword: confirmPassword ?? '',
    });
    const newErrors: { username?: string; password?: string; confirmPassword?: string } = {};
    if (!result.success) {
      result.error.issues.forEach((issue) => {
        const field = issue.path[0] as keyof typeof newErrors;
        if (!newErrors[field]) newErrors[field] = issue.message;
      });
    }
    setManualErrors(newErrors);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [watchedUsername, watchedPassword, watchedConfirmPassword]);

  const resetForm = () => {
    reset({ username: '', password: '', confirmPassword: '' });
    setPermissions({});
    setManualErrors({});
    setErrMsg('');
    setSaveDisabled(true);
  };

  return {
    formRef,
    register,
    errors,
    reset,
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
  };
}
