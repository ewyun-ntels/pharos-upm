import { renderHook } from '@testing-library/react';
import { useCreate, useUpdate } from '@/lib/data-provider';
import { useCreateAlertRule, useUpdateAlertRule } from './useAlertRuleMutation';
import { ALERT_RESOURCES } from '../types';

// Mock @/lib/data-provider
jest.mock('@/lib/data-provider', () => ({
  useCreate: jest.fn(),
  useUpdate: jest.fn(),
  HttpError: class HttpError extends Error {
    statusCode: number;
    constructor(message: string, statusCode: number) {
      super(message);
      this.statusCode = statusCode;
    }
  }
}));

describe('useAlertRuleMutation', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('useCreateAlertRule', () => {
    it('should validate and create a query alert rule', () => {
      const mockMutate = jest.fn();
      (useCreate as jest.Mock).mockReturnValue({
        mutate: mockMutate,
        isLoading: false,
        error: null
      });

      const { result } = renderHook(() => useCreateAlertRule());

      const validData = {
        name: 'Test Alert',
        alert_type: 'query' as const,
        datasource: 'clickhouse',
        evaluation_interval: '1m',
        datasource_query: {
          query: 'SELECT * FROM metrics',
          time_label: 'timestamp',
          variable_label: 'host'
        },
        threshold: [
          {
            id: '1',
            condition: 'is_above' as const,
            start_value: 80,
            severity: 'Critical' as const
          }
        ]
      };

      result.current.mutate(validData);

      expect(mockMutate).toHaveBeenCalledWith({
        resource: ALERT_RESOURCES.RULE,
        values: expect.objectContaining({
          name: 'Test Alert',
          alert_type: 'query',
          datasource: 'clickhouse'
        })
      });
    });

    it('should throw error for invalid alert rule data', () => {
      const mockMutate = jest.fn();
      (useCreate as jest.Mock).mockReturnValue({
        mutate: mockMutate,
        isLoading: false,
        error: null
      });

      const { result } = renderHook(() => useCreateAlertRule());

      const invalidData = {
        name: 'Test Alert',
        alert_type: 'invalid_type', // Invalid enum value
      };

      // Zod will throw ZodError which should be caught
      expect(() => {
        result.current.mutate(invalidData as any);
      }).toThrow();

      expect(mockMutate).not.toHaveBeenCalled();
    });

    it('should validate event-history alert rule', () => {
      const mockMutate = jest.fn();
      (useCreate as jest.Mock).mockReturnValue({
        mutate: mockMutate,
        isLoading: false,
        error: null
      });

      const { result } = renderHook(() => useCreateAlertRule());

      const validData = {
        name: 'Event History Alert',
        alert_type: 'event-history' as const,
        retention_period: 30,
        notifications: ['email@example.com']
      };

      result.current.mutate(validData);

      expect(mockMutate).toHaveBeenCalledWith({
        resource: ALERT_RESOURCES.RULE,
        values: expect.objectContaining({
          name: 'Event History Alert',
          alert_type: 'event-history',
          retention_period: 30
        })
      });
    });
  });

  describe('useUpdateAlertRule', () => {
    it('should update alert rule with partial data', () => {
      const mockMutate = jest.fn();
      (useUpdate as jest.Mock).mockReturnValue({
        mutate: mockMutate,
        isLoading: false,
        error: null
      });

      const { result } = renderHook(() => useUpdateAlertRule());

      const updateData = {
        name: 'Updated Alert Name',
        evaluation_interval: '5m'
      };

      result.current.mutate('rule-123', updateData);

      expect(mockMutate).toHaveBeenCalledWith({
        resource: ALERT_RESOURCES.RULE,
        id: 'rule-123',
        values: updateData
      });
    });

    it('should throw error for invalid update data', () => {
      const mockMutate = jest.fn();
      (useUpdate as jest.Mock).mockReturnValue({
        mutate: mockMutate,
        isLoading: false,
        error: null
      });

      const { result } = renderHook(() => useUpdateAlertRule());

      expect(() => {
        result.current.mutate('rule-123', null as any);
      }).toThrow('Update data must be an object');

      expect(mockMutate).not.toHaveBeenCalled();
    });
  });
});
