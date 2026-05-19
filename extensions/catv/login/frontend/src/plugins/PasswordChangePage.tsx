import { useState, useEffect } from 'react';
import type { AuthActionBaseProps, PasswordPolicy } from '@pharos/shared/features/auth';
import {
  LoginTemplate,
  PasswordChangeFormTemplate,
} from '@pharos/shared/components/template/login-modern-card';
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogAction,
} from '@pharos/shared/components/ui';
import { Logo } from '../components/Logo';
import {
  passwordChangeLabels,
  errorMessages,
  getPolicyMessage,
} from './labels';

export interface PasswordChangeActionProps extends AuthActionBaseProps {
  passwordPolicy: PasswordPolicy;
  isExpired: boolean;
  expiresAt?: Date | null;
}

export default function PasswordChangePage({
  user,
  passwordPolicy,
  isExpired,
  context,
  onCancel,
}: PasswordChangeActionProps) {
  const [errorMessage, setErrorMessage] = useState('');
  const [guideMessage, setGuideMessage] = useState('');
  const [showAlert, setShowAlert] = useState(false);

  useEffect(() => {
    if (!user?.password_expired_at) return;

    const expirationDate = new Date(user.password_expired_at).getTime();
    const currentDate = new Date().getTime();

    if (expirationDate <= currentDate) {
      setGuideMessage(errorMessages.passwordExpired);
    } else {
      setGuideMessage(errorMessages.firstLogin);
    }
  }, [user]);

  const policyMessage = passwordPolicy ? getPolicyMessage(passwordPolicy) : '';

  const handleSubmit = async (newPassword: string, currentPassword?: string) => {
    setErrorMessage('');

    const result = await context.changePassword(newPassword, currentPassword);

    if (result.success) {
      setShowAlert(true);
    } else {
      setErrorMessage(result.error || 'An unexpected error occurred.');
    }
  };

  const handleBack = () => {
    if (onCancel) {
      onCancel();
    }
  };

  const handleAlertConfirm = () => {
    setShowAlert(false);
    handleBack();
  };

  return (
    <LoginTemplate
      header={{
        logo: <Logo width={16} height={16} />,
        brandName: 'Pharos',
      }}
    >
      <PasswordChangeFormTemplate
        onSubmit={handleSubmit}
        onBack={handleBack}
        errorMessage={errorMessage}
        guideMessage={guideMessage}
        policyMessage={policyMessage}
        passwordPolicy={passwordPolicy}
        isExpired={isExpired}
        labels={passwordChangeLabels}
      />

      <AlertDialog open={showAlert} onOpenChange={setShowAlert}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{errorMessages.passwordChangeSuccess}</AlertDialogTitle>
            <AlertDialogDescription>{errorMessages.passwordChangeDescription}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogAction onClick={handleAlertConfirm}>OK</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </LoginTemplate>
  );
}
