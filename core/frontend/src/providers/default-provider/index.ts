import type { DataProvider } from '@/lib/data-provider';

export const DEFAULT_PROVIDER_NAME = 'default' as const;

export const DEFAULT_RESOURCES = {
  METRICS: 'metrics',
} as const;

/**
 * Secure default provider that rejects all operations
 *
 * This provider is used as a fallback to satisfy TypeScript requirements
 * but prevents any actual API calls from being made without explicit
 * dataProviderName specification.
 */
export const defaultProvider: DataProvider = {
  getList: () => Promise.reject(new Error('Default provider should not be used. Specify explicit dataProviderName.')),
  getOne: () => Promise.reject(new Error('Default provider should not be used. Specify explicit dataProviderName.')),
  create: () => Promise.reject(new Error('Default provider should not be used. Specify explicit dataProviderName.')),
  update: () => Promise.reject(new Error('Default provider should not be used. Specify explicit dataProviderName.')),
  deleteOne: () => Promise.reject(new Error('Default provider should not be used. Specify explicit dataProviderName.')),
  getApiUrl: () => '',
};