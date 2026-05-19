import { renderHook } from '@testing-library/react';
import { useUpdate } from '@/lib/data-provider';
import { useUpdateAlertStatus, useMaskAlertStatus } from './useAlertStatusMutation';
import { ALERT_RESOURCES } from '../types';

// Mock @/lib/data-provider
jest.mock('@/lib/data-provider', () => ({
  useUpdate: jest.fn(),
  HttpError: class HttpError extends Error {
    statusCode: number;
    constructor(message: string, statusCode: number) {
      super(message);
      this.statusCode = statusCode;
    }
  }
}));

describe('useAlertStatusMutation', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('useUpdateAlertStatus', () => {
    it('should update alert status with partial data', () => {
      const mockMutate = jest.fn();
      (useUpdate as jest.Mock).mockReturnValue({
        mutate: mockMutate,
        isLoading: false,
        error: null
      });

      const { result } = renderHook(() => useUpdateAlertStatus());

      const updateData = {
        status: 'normal' as const,
        severity: 'Normal' as const
      };

      result.current.mutate('alert-123', updateData);

      expect(mockMutate).toHaveBeenCalledWith({
        resource: ALERT_RESOURCES.STATUS,
        id: 'alert-123',
        values: updateData
      });
    });

    it('should throw error for invalid status data', () => {
      const mockMutate = jest.fn();
      (useUpdate as jest.Mock).mockReturnValue({
        mutate: mockMutate,
        isLoading: false,
        error: null
      });

      const { result } = renderHook(() => useUpdateAlertStatus());

      const invalidData = {
        status: 'invalid_status', // Invalid enum value
      };

      expect(() => {
        result.current.mutate('alert-123', invalidData as any);
      }).toThrow();

      expect(mockMutate).not.toHaveBeenCalled();
    });
  });

  describe('useMaskAlertStatus', () => {
    it('should mask alert status', () => {
      const mockMutate = jest.fn();
      (useUpdate as jest.Mock).mockReturnValue({
        mutate: mockMutate,
        isLoading: false,
        error: null
      });

      const { result } = renderHook(() => useMaskAlertStatus());

      const maskData = {
        mask: true
      };

      result.current.mutate('alert-123', maskData);

      expect(mockMutate).toHaveBeenCalledWith({
        resource: ALERT_RESOURCES.STATUS_MASK,
        id: 'alert-123',
        values: maskData
      });
    });

    it('should unmask alert status', () => {
      const mockMutate = jest.fn();
      (useUpdate as jest.Mock).mockReturnValue({
        mutate: mockMutate,
        isLoading: false,
        error: null
      });

      const { result } = renderHook(() => useMaskAlertStatus());

      const unmaskData = {
        mask: false
      };

      result.current.mutate('alert-123', unmaskData);

      expect(mockMutate).toHaveBeenCalledWith({
        resource: ALERT_RESOURCES.STATUS_MASK,
        id: 'alert-123',
        values: unmaskData
      });
    });

    it('should throw error for invalid mask data', () => {
      const mockMutate = jest.fn();
      (useUpdate as jest.Mock).mockReturnValue({
        mutate: mockMutate,
        isLoading: false,
        error: null
      });

      const { result } = renderHook(() => useMaskAlertStatus());

      const invalidData = {
        mask: 'not_a_boolean' // Should be boolean
      };

      expect(() => {
        result.current.mutate('alert-123', invalidData as any);
      }).toThrow();

      expect(mockMutate).not.toHaveBeenCalled();
    });
  });
});
