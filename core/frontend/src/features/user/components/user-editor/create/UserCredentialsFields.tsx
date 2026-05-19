import React from 'react';
import { Input } from '@pharos/shared/components/ui';
import { Title } from '@pharos/shared/components/ui-extension';

// create 전용 컴포넌트 — POST /user 에 대응 (username + password + confirmPassword)
interface Props {
  register: any;
  setValue: any;
  usernameRequirements: {
    min_length: number;
    max_length: number;
    email_allowed: boolean;
    requirementMsg: string;
  };
  passwordRequirements: {
    min_length: number;
    min_uppercase: number;
    min_lowercase: number;
    min_digits: number;
    min_special: number;
    requirementMsg: string;
  };
  errors: any;
  manualErrors: any;
  clearErrors: (field?: string | string[]) => void;
  getValues?: (field: string) => string;
  trigger?: (field: string) => Promise<boolean>;
}

export const UserCredentialsFields: React.FC<Props> = ({
  register,
  setValue,
  manualErrors,
  usernameRequirements,
  passwordRequirements,
  clearErrors,
  getValues,
  trigger,
}) => (
  <div className="flex flex-col gap-4 max-w-3xl mb-3">
    <div>
      <Title variant="formLabel" htmlFor="username">
        Username <span className="text-red-500">*</span>
      </Title>
      <Input
        type="text"
        id="username"
        className={`px-2 py-1 ${getValues?.('username') !== '' && manualErrors.username ? 'border-red-500' : 'border-border'}`}
        {...register('username', {
          onChange: (e: any) => {
            const value = e.target.value.replace(/\s/g, '');
            if (!usernameRequirements.email_allowed && /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
              return;
            }
            setValue('username', value);
            if (value === '') clearErrors('username');
          },
        })}
        maxLength={usernameRequirements.max_length}
      />
      {getValues?.('username') !== '' && manualErrors.username && (
        <div className="text-sm mt-1">
          <span className="text-red-500">{manualErrors.username as string}</span>
        </div>
      )}
      <div className="text-xs text-muted-foreground mt-2">{usernameRequirements.requirementMsg}</div>
    </div>

    <div>
      <Title variant="formLabel" htmlFor="password">
        Password <span className="text-red-500">*</span>
      </Title>
      <Input
        type="password"
        id="password"
        className={`px-2 py-1 ${getValues?.('password') !== '' && manualErrors.password ? 'border-red-500' : 'border-border'}`}
        {...register('password', {
          onChange: async (e: any) => {
            const value = e.target.value.replace(/\s/g, '');
            setValue('password', value);
            if (value === '') clearErrors('password');
            if (trigger) await trigger('confirmPassword');
          },
        })}
      />
      {getValues?.('password') !== '' && manualErrors.password && (
        <div className="text-sm mt-1">
          <span className="text-red-500">{manualErrors.password as string}</span>
        </div>
      )}
      <div className="text-xs text-muted-foreground mt-2">{passwordRequirements.requirementMsg}</div>
    </div>

    <div>
      <Title variant="formLabel" htmlFor="confirmPassword">
        Confirm password <span className="text-red-500">*</span>
      </Title>
      <Input
        type="password"
        id="confirmPassword"
        className={`px-2 py-1 ${manualErrors.confirmPassword ? 'border-red-500' : 'border-border'}`}
        {...register('confirmPassword', {
          onChange: async (e: any) => {
            const value = e.target.value.replace(/\s/g, '');
            setValue('confirmPassword', value);
            if (value === '') clearErrors('confirmPassword');
            if (trigger) await trigger('confirmPassword');
          },
        })}
      />
      {manualErrors.confirmPassword && (
        <div className="text-sm mt-1">
          <span className="text-red-500">{manualErrors.confirmPassword as string}</span>
        </div>
      )}
    </div>
  </div>
);
