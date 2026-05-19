import React, { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { Button, Input, Label, Alert, AlertTitle } from '../../ui';
import { TriangleAlert } from 'lucide-react';
import type { PasswordChangeFormTemplateProps } from './types';
import { validatePasswordPolicy } from '../../../utils/validatePasswordPolicy';

const passwordSchema = z
  .object({
    currentPassword: z.string().optional(),
    newPassword: z.string().min(1, 'Password is required'),
    confirmPassword: z.string().min(1, 'Confirm password is required'),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  });

type PasswordFormData = z.infer<typeof passwordSchema>;

/**
 * 기본 비밀번호 변경 폼 템플릿
 */
export function PasswordChangeFormTemplate({
  onSubmit,
  onBack,
  errorMessage,
  guideMessage,
  policyMessage,
  passwordPolicy,
  isLoading = false,
  isExpired = false,
  labels = {},
}: PasswordChangeFormTemplateProps) {
  const [error, setError] = useState(errorMessage || '');

  // 기본 레이블
  const {
    title = 'Change password',
    currentPassword: currentPasswordLabel = 'Current password',
    newPassword: newPasswordLabel = 'New password',
    confirmPassword: confirmPasswordLabel = 'Confirm password',
    changePassword = 'Change password',
    changing = 'Changing...',
    backToLogin = 'Back to Login',
  } = labels;

  const {
    register,
    handleSubmit,
    reset,
    watch,
    formState: { errors },
  } = useForm<PasswordFormData>({
    resolver: zodResolver(passwordSchema),
    mode: 'onChange',
  });

  const newPassword = watch('newPassword');
  const confirmPassword = watch('confirmPassword');
  const isFormEmpty = !newPassword && !confirmPassword;

  useEffect(() => {
    if (errorMessage) {
      setError(errorMessage);
    }
  }, [errorMessage]);

  const handleFormSubmit = async (data: PasswordFormData) => {
    setError('');
    const policyError = validatePasswordPolicy(passwordPolicy, data.newPassword);
    if (policyError) {
      setError(policyError);
      return;
    }
    try {
      await onSubmit(data.newPassword, data.currentPassword);
      reset();
    } catch (err) {
      console.error('Password change failed:', err);
    }
  };

  return (
    <div className="w-80">
      <header className="mb-8">
        <h1 className="text-2xl">{title}</h1>
      </header>

      {guideMessage && (
        <Alert className="text-yellow-800 rounded border-0 bg-yellow-100 mb-4">
          <TriangleAlert className="mr-2 shrink-0" />
          <AlertTitle className="line-clamp-none whitespace-normal wrap-break-word">
            {guideMessage}
          </AlertTitle>
        </Alert>
      )}

      {!isExpired && (
        <div className="w-full mb-6 relative flex flex-col gap-2">
          <Label htmlFor="currentPassword">{currentPasswordLabel}</Label>
          <Input
            type="password"
            id="currentPassword"
            disabled={isLoading}
            {...register('currentPassword')}
          />
          {errors.currentPassword && (
            <div className="text-sm text-red-500">
              {errors.currentPassword.message}
            </div>
          )}
        </div>
      )}

      <div className="w-full mb-6 relative flex flex-col gap-2">
        <Label htmlFor="newPassword">{newPasswordLabel}</Label>
        <Input
          type="password"
          id="newPassword"
          disabled={isLoading}
          {...register('newPassword', {
            onChange: () => {
              if (error) setError('');
            },
          })}
        />
        {policyMessage && (
          <div className="text-xs text-muted-foreground mt-2">
            {policyMessage}
          </div>
        )}
        {error && <div className="text-sm text-red-500">{error}</div>}
        {errors.newPassword && (
          <div className="text-sm text-red-500">{errors.newPassword.message}</div>
        )}
      </div>

      <div className="w-full mb-6 relative flex flex-col gap-2">
        <Label htmlFor="confirmPassword">{confirmPasswordLabel}</Label>
        <Input
          type="password"
          id="confirmPassword"
          disabled={isLoading}
          {...register('confirmPassword')}
        />
        {errors.confirmPassword && (
          <div className="text-sm text-red-500">
            {errors.confirmPassword.message}
          </div>
        )}
      </div>

      <div className="flex flex-col gap-4">
        <Button
          type="button"
          variant="default"
          className="w-full cursor-pointer dark:text-white"
          disabled={isFormEmpty || isLoading}
          onClick={handleSubmit(handleFormSubmit)}
        >
          {isLoading ? changing : changePassword}
        </Button>

        <Button
          type="button"
          variant="ghost"
          className="w-full cursor-pointer dark:text-white"
          onClick={onBack}
          disabled={isLoading}
        >
          {backToLogin}
        </Button>
      </div>
    </div>
  );
}
