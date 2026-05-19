import { renderHook, act } from '@testing-library/react';
import { useAuthStore } from '../auth-store';
import type { User } from '../../../types/user';

describe('useAuthStore', () => {
  beforeEach(() => {
    // Reset store before each test
    const { result } = renderHook(() => useAuthStore());
    act(() => {
      result.current.reset();
    });
  });

  describe('initialize', () => {
    it('should initialize with user data and trigger onLogin callback', async () => {
      const mockUser: User = {
        name: 'testuser',
        created_at: new Date('2025-01-01'),
        attributes: {
          roles: {
            'role:super_admin': true,
            'role:user_read': true,
          }
        },
      };

      const mockFetchUser = jest.fn().mockResolvedValue(mockUser);
      const mockOnLogin = jest.fn();

      const { result } = renderHook(() => useAuthStore());

      await act(async () => {
        await result.current.initialize(mockFetchUser, {
          onLogin: mockOnLogin,
        });
      });

      expect(mockFetchUser).toHaveBeenCalledTimes(1);
      expect(mockOnLogin).toHaveBeenCalledWith(mockUser);
      expect(result.current.user).toEqual(mockUser);
      expect(result.current.isAuthenticated).toBe(true);
    });

    it('should handle fetch error gracefully', async () => {
      const mockError = new Error('Fetch failed');
      const mockFetchUser = jest.fn().mockRejectedValue(mockError);

      const { result } = renderHook(() => useAuthStore());

      await act(async () => {
        await result.current.initialize(mockFetchUser);
      });

      expect(result.current.user).toBeNull();
      expect(result.current.isAuthenticated).toBe(false);
    });
  });

  describe('logout', () => {
    it('should clear user data and trigger onLogout callback', async () => {
      const mockUser: User = {
        name: 'testuser',
        created_at: new Date('2025-01-01'),
        attributes: {},
      };

      const mockFetchUser = jest.fn().mockResolvedValue(mockUser);
      const mockOnLogout = jest.fn();

      const { result } = renderHook(() => useAuthStore());

      // Initialize first
      await act(async () => {
        await result.current.initialize(mockFetchUser, {
          onLogout: mockOnLogout,
        });
      });

      expect(result.current.isAuthenticated).toBe(true);

      // Then logout
      await act(async () => {
        await result.current.logout();
      });

      expect(mockOnLogout).toHaveBeenCalledTimes(1);
      expect(result.current.user).toBeNull();
      expect(result.current.isAuthenticated).toBe(false);
    });
  });

  describe('setUser', () => {
    it('should update user data and trigger onUserUpdate callback', async () => {
      const mockUser: User = {
        name: 'testuser',
        created_at: new Date('2025-01-01'),
        attributes: { 
          roles: { 'role:user_read': false } 
        },
      };

      const updatedUser: User = {
        ...mockUser,
        attributes: { 
          roles: { 'role:user_read': true } 
        },
      };

      const mockFetchUser = jest.fn().mockResolvedValue(mockUser);
      const mockOnUserUpdate = jest.fn();

      const { result } = renderHook(() => useAuthStore());

      // Initialize first
      await act(async () => {
        await result.current.initialize(mockFetchUser, {
          onUserUpdate: mockOnUserUpdate,
        });
      });

      // Update user
      act(() => {
        result.current.setUser(updatedUser);
      });

      expect(mockOnUserUpdate).toHaveBeenCalledWith(updatedUser);
      expect(result.current.user?.attributes?.roles?.['role:user_read']).toBe(true);
    });
  });

  describe('hasPermission', () => {
    it('should return true if user has permission', async () => {
      const mockUser: User = {
        name: 'testuser',
        created_at: new Date('2025-01-01'),
        attributes: {
          roles: {
            'role:user_read': true,
            'role:dashboard_create': false,
          },
          info: {},
        },
      };

      const mockFetchUser = jest.fn().mockResolvedValue(mockUser);
      const { result } = renderHook(() => useAuthStore());

      await act(async () => {
        await result.current.initialize(mockFetchUser);
      });

      expect(result.current.hasPermission('role:user_read')).toBe(true);
      expect(result.current.hasPermission('role:dashboard_create')).toBe(false);
    });

    it('should return true for any permission if user is super_admin', async () => {
      const mockUser: User = {
        name: 'superuser',
        created_at: new Date('2025-01-01'),
        attributes: {
          roles: {
            'role:super_admin': true,
            'role:user_read': false,
          },
          info: {},
        },
      };

      const mockFetchUser = jest.fn().mockResolvedValue(mockUser);
      const { result } = renderHook(() => useAuthStore());

      await act(async () => {
        await result.current.initialize(mockFetchUser);
      });

      expect(result.current.hasPermission('role:super_admin')).toBe(true);
      expect(result.current.hasPermission('role:user_read')).toBe(true); // super_admin bypasses
      expect(result.current.hasPermission('role:dashboard_create')).toBe(true); // super_admin bypasses
    });

    it('should return false if user is not authenticated', () => {
      const { result } = renderHook(() => useAuthStore());
      expect(result.current.hasPermission('role:super_admin')).toBe(false);
    });
  });

  describe('hasAnyPermission', () => {
    it('should return true if user has any of the permissions', async () => {
      const mockUser: User = {
        name: 'testuser',
        created_at: new Date('2025-01-01'),
        attributes: {
          roles: {
            'role:user_read': true,
            'role:dashboard_create': false,
          },
          info: {},
        },
      };

      const mockFetchUser = jest.fn().mockResolvedValue(mockUser);
      const { result } = renderHook(() => useAuthStore());

      await act(async () => {
        await result.current.initialize(mockFetchUser);
      });

      expect(result.current.hasAnyPermission(['role:user_read', 'role:dashboard_create'])).toBe(true);
      expect(result.current.hasAnyPermission(['role:dashboard_create', 'role:alert_read'])).toBe(false);
    });

    it('should return true for super_admin even if other permissions are false', async () => {
      const mockUser: User = {
        name: 'superuser',
        created_at: new Date('2025-01-01'),
        attributes: {
          roles: {
            'role:super_admin': true,
            'role:user_read': false,
          },
          info: {},
        },
      };

      const mockFetchUser = jest.fn().mockResolvedValue(mockUser);
      const { result } = renderHook(() => useAuthStore());

      await act(async () => {
        await result.current.initialize(mockFetchUser);
      });

      expect(result.current.hasAnyPermission(['role:user_read', 'role:dashboard_create'])).toBe(true);
    });
  });

  describe('hasAllPermissions', () => {
    it('should return true only if user has all permissions', async () => {
      const mockUser: User = {
        name: 'testuser',
        created_at: new Date('2025-01-01'),
        attributes: {
          roles: {
            'role:super_admin': true,
            'role:user_read': true,
          },
          info: {},
        },
      };

      const mockFetchUser = jest.fn().mockResolvedValue(mockUser);
      const { result } = renderHook(() => useAuthStore());

      await act(async () => {
        await result.current.initialize(mockFetchUser);
      });

      expect(result.current.hasAllPermissions(['role:super_admin', 'role:user_read'])).toBe(true);
      expect(result.current.hasAllPermissions(['role:super_admin', 'role:dashboard_create'])).toBe(true); // super_admin bypasses
    });

    it('should return true for super_admin with any permission combination', async () => {
      const mockUser: User = {
        name: 'superuser',
        created_at: new Date('2025-01-01'),
        attributes: {
          roles: {
            'role:super_admin': true,
            'role:user_read': false,
          },
          info: {},
        },
      };

      const mockFetchUser = jest.fn().mockResolvedValue(mockUser);
      const { result } = renderHook(() => useAuthStore());

      await act(async () => {
        await result.current.initialize(mockFetchUser);
      });

      expect(result.current.hasAllPermissions(['role:user_read', 'role:dashboard_create'])).toBe(true);
    });
  });

  describe('Selectors', () => {
    it('useHasPermission selector should work', async () => {
      const mockUser: User = {
        name: 'testuser',
        created_at: new Date('2025-01-01'),
        attributes: {
          roles: {
            'role:super_admin': true,
          }
        },
      };

      const mockFetchUser = jest.fn().mockResolvedValue(mockUser);
      const storeHook = renderHook(() => useAuthStore());

      await act(async () => {
        await storeHook.result.current.initialize(mockFetchUser);
      });

      const { result: selectorResult } = renderHook(() =>
        useAuthStore((state) => state.hasPermission('role:super_admin'))
      );

      expect(selectorResult.current).toBe(true);
    });
  });
});
