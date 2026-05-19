import { authActionPluginRegistry } from '@pharos/core/login-registry';
import { registerSidebarLogo } from '@pharos/core/site-registry';
import { Logo } from './components/Logo';

registerSidebarLogo(Logo);
import type { PasswordChangeActionProps } from './plugins/PasswordChangePage';
import type { LoginActionProps } from './plugins/LoginPage';
import type { PasswordPolicy, AuthConfig as SharedAuthConfig } from '@pharos/shared/features/auth';

authActionPluginRegistry.register<LoginActionProps>({
  info: {
    id: 'login',
    label: 'Sign In',
    description: 'Default login page',
  },
  component: () => import('./plugins/LoginPage'),
});

authActionPluginRegistry.register<PasswordChangeActionProps>({
  info: {
    id: 'password_change',
    label: 'Change Password',
    description: 'Password change page for initial login or expired password',
  },
  component: () => import('./plugins/PasswordChangePage'),
  getProps: async (user, context) => {
    let passwordPolicy: PasswordPolicy = {
      min_length: 8,
      min_uppercase: 1,
      min_lowercase: 1,
      min_digits: 1,
      min_special: 1,
    };

    try {
      const config = await context?.dataProvider.getOne<SharedAuthConfig>({ resource: 'auth/config' });
      if (config?.password) {
        passwordPolicy = config.password;
      }
    } catch {
      console.warn('[default/login] Failed to fetch password policy');
    }

    const isExpired = user?.password_expired_at
      ? new Date(user.password_expired_at) < new Date()
      : false;

    return {
      passwordPolicy,
      isExpired,
      expiresAt: user?.password_expired_at ? new Date(user.password_expired_at) : null,
    };
  },
});
