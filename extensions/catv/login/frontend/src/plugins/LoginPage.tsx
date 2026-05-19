import { useState, useEffect } from 'react';
import type { AuthActionBaseProps, AuthConfig } from '@pharos/shared/features/auth';
import {
  LoginTemplate,
  LoginFormTemplate,
} from '@pharos/shared/components/template/login-modern-card';
import { Logo } from '../components/Logo';
import { loginLabels, errorMessages } from './labels';

export interface LoginActionProps extends AuthActionBaseProps {
  errorMessage?: string;
}

export default function LoginPage({ context, onComplete }: LoginActionProps) {
  const [errorMessage, setErrorMessage] = useState('');
  const [emailAllowed, setEmailAllowed] = useState<boolean | null>(null);

  useEffect(() => {
    context.dataProvider.getOne<AuthConfig>({ resource: 'auth', id: 'config' })
      .then((config) => setEmailAllowed(config.user.email_allowed))
      .catch(() => setEmailAllowed(true));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleLogin = async (username: string, password: string) => {
    setErrorMessage('');

    const result = await context.login(username, password);

    if (result.success) {
      onComplete();
    } else {
      const errorKey = result.error as keyof typeof errorMessages;
      setErrorMessage(errorMessages[errorKey] || errorMessages.userNotFound);
    }
  };

  const usernameLabel = emailAllowed === false ? 'Username' : 'Email or Username';

  return (
    <LoginTemplate>
      <LoginFormTemplate
        logo={<Logo height={80} />}
        onSubmit={handleLogin}
        errorMessage={errorMessage}
        labels={{ ...loginLabels, username: usernameLabel }}
      />
    </LoginTemplate>
  );
}
