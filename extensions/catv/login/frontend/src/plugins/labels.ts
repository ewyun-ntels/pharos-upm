import type { PasswordPolicy } from '@pharos/shared/features/auth';
import type {
  LoginFormLabels,
  PasswordChangeFormLabels,
} from '@pharos/shared/components/template/login-modern-card';

export const loginLabels: LoginFormLabels = {
  title: 'Welcome Back',
  subtitle: 'Enter your credentials to access your account',
  username: 'Email or Username',
  password: 'Password',
  signIn: 'Sign in',
  signingIn: 'Signing in...',
  forgotPassword: 'Forgot your password?',
  noAccount: "Don't have an account?",
  signUp: 'Sign up',
};

export const passwordChangeLabels: PasswordChangeFormLabels = {
  title: 'Change Password',
  currentPassword: 'Current Password',
  newPassword: 'New Password',
  confirmPassword: 'Confirm Password',
  submit: 'Change Password',
  submitting: 'Changing...',
  passwordMismatch: 'Passwords do not match',
};

export const errorMessages = {
  userNotFound: 'Invalid username or password.',
  userBlocked: 'This account has been blocked.',
  userTemporarilyBlocked: 'This account is temporarily blocked. Please try again later.',
  passwordExpired: 'Your password has expired. Please change your password.',
  firstLogin: 'Please change your password on first login.',
  passwordChangeSuccess: 'Password changed successfully.',
  passwordChangeDescription: 'Please sign in with your new password.',
};

export function getPolicyMessage(policy: PasswordPolicy): string {
  return (
    `Minimum ${policy.min_length} character${policy.min_length > 1 ? 's' : ''}, ` +
    (policy.require_uppercase ? 'uppercase, ' : '') +
    (policy.require_lowercase ? 'lowercase, ' : '') +
    (policy.require_number ? 'number, ' : '') +
    (policy.require_special_char ? 'special character' : '')
  )
    .replace(/,\s*$/, '')
    .replace(/, ([^,]+)$/, ' and $1');
}
