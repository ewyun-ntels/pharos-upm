import type { ResourceProvider } from '@pharos/shared/lib/data-provider/types';

export interface AuthProvider {
  login: (params: {
    email?: string;
    username?: string;
    password?: string;
    [key: string]: unknown;
  }) => Promise<{
    success: boolean;
    redirectTo?: string;
    error?: { message: string; name: string };
  }>;
  logout: () => Promise<{ success: boolean }>;
  check: () => Promise<{ authenticated: boolean; redirectTo?: string }>;
  getIdentity: () => Promise<unknown>;
  getPermissions?: () => Promise<unknown>;
  onError?: (error: unknown) => Promise<{ error?: unknown }>;
}

// ---------- DataProviderRegistry ----------

class DataProviderRegistry {
  private providers = new Map<string, ResourceProvider>();

  register(name: string, provider: ResourceProvider): void {
    this.providers.set(name, provider);
  }

  get(name?: string): ResourceProvider {
    const key = name ?? 'default';
    const provider = this.providers.get(key);
    if (!provider) {
      throw new Error(
        `DataProvider '${key}' is not registered. ` +
          `Call dataProviderRegistry.register('${key}', provider) first.`,
      );
    }
    return provider;
  }

  has(name: string): boolean {
    return this.providers.has(name);
  }
}

export const dataProviderRegistry = new DataProviderRegistry();

// ---------- AuthProvider singleton ----------

let _authProvider: AuthProvider | undefined;

export function setAuthProvider(provider: AuthProvider): void {
  _authProvider = provider;
}

export function getAuthProvider(): AuthProvider {
  if (!_authProvider) {
    throw new Error(
      'Auth provider is not set. Call setAuthProvider() before using auth hooks.',
    );
  }
  return _authProvider;
}
