import React, { useState } from 'react';
import { Button, Input, Label, Alert, AlertTitle } from '../../ui';
import { TriangleAlert } from 'lucide-react';
import type { LoginFormTemplateProps } from './types';

/**
 * 기본 로그인 폼 템플릿
 */
export function LoginFormTemplate({
  onSubmit,
  errorMessage,
  isLoading = false,
  labels = {},
}: LoginFormTemplateProps) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState(errorMessage || '');

  // 기본 레이블
  const {
    title = 'Sign in',
    username: usernameLabel = 'Username',
    password: passwordLabel = 'Password',
    signIn = 'Sign in',
    signingIn = 'Signing in...',
  } = labels;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    try {
      await onSubmit(username, password);
    } catch (err) {
      console.error('Login failed:', err);
    }
  };

  React.useEffect(() => {
    if (errorMessage) {
      setError(errorMessage);
    }
  }, [errorMessage]);

  return (
    <div className="w-80">
      <header className="mb-8">
        <h1 className="text-2xl">{title}</h1>
      </header>

      {error && (
        <Alert className="text-red-800 rounded border-0 bg-red-100 mb-4">
          <TriangleAlert className="mr-2 shrink-0" />
          <AlertTitle className="line-clamp-none whitespace-normal wrap-break-word">
            {error}
          </AlertTitle>
        </Alert>
      )}

      <form onSubmit={handleSubmit}>
        <div className="w-full mb-6 relative flex flex-col gap-2">
          <Label htmlFor="username">{usernameLabel}</Label>
          <Input
            type="text"
            id="username"
            required
            value={username}
            disabled={isLoading}
            onChange={(e) => {
              setUsername(e.target.value);
              if (error) setError('');
            }}
          />
        </div>

        <div className="w-full mb-6 relative flex flex-col gap-2">
          <Label htmlFor="password">{passwordLabel}</Label>
          <Input
            type="password"
            id="password"
            required
            value={password}
            disabled={isLoading}
            onChange={(e) => {
              setPassword(e.target.value);
              if (error) setError('');
            }}
          />
        </div>

        <Button
          type="submit"
          variant="default"
          className="w-full cursor-pointer dark:text-white"
          disabled={isLoading}
        >
          {isLoading ? signingIn : signIn}
        </Button>
      </form>
    </div>
  );
}
