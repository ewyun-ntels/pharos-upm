import React, { useState, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@pharos/shared/components/ui';
import { Button } from '@pharos/shared/components/ui';
import { Input } from '@pharos/shared/components/ui';
import { useUpdate } from '@/lib/data-provider';
import { useForm } from 'react-hook-form';
import { useToast } from '@hooks/use-toast';
import { z } from 'zod';
import { usePasswordRequirements } from '@features/user';
import { TriangleAlert } from '@pharos/shared/components';
import { Alert, AlertTitle } from '@pharos/shared/components/ui';
import { zodResolver } from '@hookform/resolvers/zod';
import { USER_PROVIDER_NAME, USER_RESOURCES } from '@providers/user-provider';
import { formatDateTime } from '@lib/format-date';

interface DialogProps {
  open: boolean;
  onOpenChange: () => void;
  dialogTitle: string;
  passwordExpiryDate: number;
}

export function UserPasswordDialog({
  dialogTitle,
  open,
  onOpenChange,
  passwordExpiryDate,
}: DialogProps) {
  const peKey = 'pwExp-dialog';
  const { toast } = useToast();
  const { mutate: updatePassword } = useUpdate();
  const [error, setError] = useState<string | null>(null);
  const [saveDisabled, setSaveDisabled] = useState<boolean>(true);
  const [remainingMsg, setRemainingMsg] = useState<string>('');

  const passwordRequirements = usePasswordRequirements();

  const [manualErrors, setManualErrors] = useState<{
    newPassword?: string;
    confirmPassword?: string;
  }>({});

  const passwordSchema = z
    .object({
      newPassword: z.string(),
      confirmPassword: z.string(),
    })
    .refine(
      (data) => {
        const pw = data.newPassword;
        if (!pw) return true;
        if (
          passwordRequirements.min_uppercase &&
          (pw.match(/[A-Z]/g)?.length ?? 0) < passwordRequirements.min_uppercase
        ) {
          return false;
        }
        if (
          passwordRequirements.min_lowercase &&
          (pw.match(/[a-z]/g)?.length ?? 0) < passwordRequirements.min_lowercase
        ) {
          return false;
        }
        if (
          passwordRequirements.min_digits &&
          (pw.match(/[0-9]/g)?.length ?? 0) < passwordRequirements.min_digits
        ) {
          return false;
        }
        return !(
          passwordRequirements.min_special &&
          (pw.match(/[^A-Za-z0-9]/g)?.length ?? 0) < passwordRequirements.min_special
        );
      },
      {
        message: `Invalid password. Please check the password rules and try again.`,
        path: ['newPassword'],
      },
    )
    .refine(
      (data) => {
        if (!data.newPassword && !data.confirmPassword) return true;
        if (!data.newPassword || !data.confirmPassword) return false;
        if (data.newPassword && !data.confirmPassword) return false;
        return data.newPassword === data.confirmPassword;
      },
      {
        message: 'Password does not match.',
        path: ['confirmPassword'],
      },
    );

  type PasswordForm = z.infer<typeof passwordSchema>;
  const {
    register,
    handleSubmit,
    reset,
    watch,
  } = useForm<PasswordForm>({
    resolver: zodResolver(passwordSchema),
    mode: 'onChange',
  });

  useEffect(() => {
    if (!open) {
      reset({
        newPassword: '',
        confirmPassword: '',
      });
      setError(null);
    }
  }, [open, reset]);

  const pwDialogSet = () => {
    const pathname = window.location.pathname;
    if (
      !pathname.startsWith('/ui/dashboards/')
    ) {
      localStorage.setItem(peKey, 'true');
    }
  };

  useEffect(() => {
    // 현재 시간보다 비밀번호 만료 시간이 더 크면 메시지 설정
    if (passwordExpiryDate === 0) return;
    const currentDate = new Date().getTime();
    const daysUntilExpiration = Math.ceil(
      (passwordExpiryDate * 1000 - currentDate) / (1000 * 60 * 60 * 24),
    );
    if (daysUntilExpiration <= 5) {
      const expiryDate = new Date(passwordExpiryDate * 1000);
      setRemainingMsg(
        `Your password will expire after ${formatDateTime(expiryDate)}. Please change your password.`,
      );
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (watch('newPassword') || watch('confirmPassword')) {
      setSaveDisabled(false);
    } else {
      setSaveDisabled(true);
    }

    const currentData = {
      newPassword: watch('newPassword') || '',
      confirmPassword: watch('confirmPassword') || '',
    };
    const result = passwordSchema.safeParse(currentData);
    const newManualErrors: { newPassword?: string; confirmPassword?: string } = {};
    if (!result.success) {
      result.error.issues.forEach((issue) => {
        const field = issue.path[0] as 'newPassword' | 'confirmPassword';
        newManualErrors[field] = issue.message;
      });
    }
    // Manual errors 상태 업데이트
    setManualErrors(newManualErrors);

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [watch('newPassword'), watch('confirmPassword')]);

  const onSubmit = async (data: z.infer<typeof passwordSchema>) => {
    updatePassword(
      {
        resource: USER_RESOURCES.ME_PASSWORD,
        id: '',
        dataProviderName: USER_PROVIDER_NAME,
        values: { password: data.newPassword },
      },
      {
        onSuccess: () => {
          toast({
            title: 'Password updated successfully',
            description: 'Your password has been changed.',
          });
          reset();
          pwDialogSet();
          onOpenChange();
        },
        onError: (error) => {
          const errorMessage = error.response?.data?.error || 'An unexpected error occurred.';
          setError(errorMessage);
        },
      },
    );
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{dialogTitle}</DialogTitle>
        </DialogHeader>

        <div>
          {remainingMsg !== '' && (
            <Alert className="text-amber-800 rounded border-0 bg-amber-100 mb-4">
              <TriangleAlert className="mr-2 shrink-0" />
              <AlertTitle className="line-clamp-none whitespace-normal wrap-break-word">
                {remainingMsg}
              </AlertTitle>
            </Alert>
          )}

          <div className="mb-5">
            <label
              htmlFor="title"
              className={`block text-sm mb-2 ${manualErrors.newPassword ? 'text-red-500' : ''}`}
            >
              New password{' '}
              <span className={`text-${manualErrors.newPassword ? 'red' : 'black'}-500`}>*</span>
            </label>
            <Input
              type="password"
              className={`px-2 py-1 ${manualErrors.newPassword ? 'border-red-500' : 'border-border'
                }`}
              {...register('newPassword')}
            />
            {manualErrors.newPassword && (
              <div className="inline-block justify-between items-center">
                <div className="text-sm">
                  <span className="text-red-500">{manualErrors.newPassword}</span>
                </div>
              </div>
            )}
            <div className="text-xs text-muted-foreground mt-2">
              {passwordRequirements?.requirementMsg}
            </div>
          </div>

          <div className="mb-4">
            <label
              htmlFor="title"
              className={`block text-sm mb-2 ${manualErrors.confirmPassword ? 'text-red-500' : ''}`}
            >
              Confirm new password{' '}
              <span className={`text-${manualErrors.confirmPassword ? 'red' : 'black'}-500`}>
                *
              </span>
            </label>
            <Input
              type="password"
              className={`px-2 py-1 ${manualErrors.confirmPassword ? 'border-red-500' : 'border-border'}`}
              {...register('confirmPassword')}
            />
            {manualErrors.confirmPassword && (
              <div className="inline-block justify-between items-center">
                <div className="text-sm">
                  <span className="text-red-500">{manualErrors.confirmPassword}</span>
                </div>
              </div>
            )}
          </div>
          {/* 오늘 다시 열지 않기 체크박스 */}
          {/* <div className="flex items-center gap-2">
            <Checkbox id="terms" />
            <Label htmlFor="terms">Don't show again today</Label>
          </div> */}
        </div>
        {error && (
          <Alert className="text-destructive border  border-destructive mb-4">
            <TriangleAlert className="mr-2 shrink-0" />
            <AlertTitle className="line-clamp-none whitespace-normal wrap-break-word">
              <p className="lowercase first-letter:uppercase after:content-['.']">{error}</p>
            </AlertTitle>
          </Alert>
        )}
        <DialogFooter>
          <Button
            size="sm"
            variant="outline"
            className="h-8 px-3 py-1 cursor-pointer"
            onClick={onOpenChange}
          >
            Cancel
          </Button>
          <Button
            size="sm"
            variant="default"
            className="h-8 shadow-none px-3 py-1 cursor-pointer text-white"
            disabled={saveDisabled}
            onClick={handleSubmit(onSubmit)}
          >
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
