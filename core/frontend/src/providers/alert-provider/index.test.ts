import { alertProvider } from './index';
import { axiosInstance } from '@lib/axios';
import {
  ALERT_RESOURCES,
  AlertType,
  AlertSeverity,
  AlertStatus,
  CreateAlertRuleRequest,
  UpdateAlertRuleRequest,
  AlertStatusMaskRequest,
  AlertQueryRequest,
} from './types';
import {
  QUERY_PARAM_START_TIME,
  QUERY_PARAM_END_TIME,
} from '@lib/query-params';

// Mock axios instance
jest.mock('@lib/axios', () => ({
  axiosInstance: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn(),
  },
}));

const mockedAxios = axiosInstance as jest.Mocked<typeof axiosInstance>;

describe('alertProvider', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('getList', () => {
    it('should fetch alert rules list', async () => {
      const mockRules = [
        {
          alert_type: 'query' as AlertType,
          rule: {
            id: 'rule-1',
            name: 'Test Rule',
            datasource: 'clickhouse',
          },
          status: [],
        },
      ];

      mockedAxios.get.mockResolvedValueOnce({ data: mockRules });

      const result = await alertProvider.getList({
        resource: ALERT_RESOURCES.RULE,
      });

      expect(mockedAxios.get).toHaveBeenCalledWith('/alert/rule?', {
        headers: undefined,
      });
      expect(result.data).toEqual(mockRules);
      expect(result.total).toBe(1);
    });

    it('should fetch alert rules list with detail', async () => {
      const mockRules = [
        {
          alert_type: 'query' as AlertType,
          rule: {
            id: 'rule-1',
            name: 'Test Rule',
            datasource: 'clickhouse',
          },
          status: [
            {
              id: 'alert-1',
              alert_id: 'alert-id-1',
              alert_type: 'query' as AlertType,
              name: 'Test Rule',
              status: 'alerting' as AlertStatus,
              severity: 'Critical' as AlertSeverity,
              value: 100,
              mask: false,
              timestamp: new Date('2025-10-14T00:00:00Z'),
              updated_at: new Date('2025-10-14T00:00:00Z'),
            },
          ],
        },
      ];

      mockedAxios.get.mockResolvedValueOnce({ data: mockRules });

      const result = await alertProvider.getList({
        resource: ALERT_RESOURCES.RULE,
        meta: { detail: true },
      });

      expect(mockedAxios.get).toHaveBeenCalledWith('/alert/rule?detail=true', {
        headers: undefined,
      });
      expect(result.data).toEqual(mockRules);
      expect(result.total).toBe(1);
    });

    it('should fetch alert status list', async () => {
      const mockStatus = [
        {
          id: 'alert-1',
          alert_id: 'alert-id-1',
          alert_type: 'query' as AlertType,
          name: 'Test Rule',
          status: 'alerting' as AlertStatus,
          severity: 'Critical' as AlertSeverity,
          value: 100,
          mask: false,
          timestamp: new Date('2025-10-14T00:00:00Z'),
          updated_at: new Date('2025-10-14T00:00:00Z'),
        },
      ];

      mockedAxios.get.mockResolvedValueOnce({ data: mockStatus });

      const result = await alertProvider.getList({
        resource: ALERT_RESOURCES.STATUS,
        filters: [
          { field: 'name', operator: 'eq', value: 'Test Rule' },
          { field: 'status', operator: 'eq', value: 'alerting' },
        ],
      });

      expect(mockedAxios.get).toHaveBeenCalledWith(
        '/alert/status?name=Test+Rule&status=alerting',
        {
          headers: undefined,
        },
      );
      expect(result.data).toEqual(mockStatus);
      expect(result.total).toBe(1);
    });

    it('should fetch alert status list with pagination', async () => {
      const mockStatus = [
        {
          id: 'alert-1',
          alert_id: 'alert-id-1',
          alert_type: 'query' as AlertType,
          name: 'Test Rule',
          status: 'alerting' as AlertStatus,
          severity: 'Critical' as AlertSeverity,
          value: 100,
          mask: false,
          timestamp: new Date('2025-10-14T00:00:00Z'),
          updated_at: new Date('2025-10-14T00:00:00Z'),
        },
      ];

      mockedAxios.get.mockResolvedValueOnce({ data: mockStatus });

      const result = await alertProvider.getList({
        resource: ALERT_RESOURCES.STATUS,
        pagination: {
          currentPage: 2,
          pageSize: 10,
          mode: 'server',
        },
      });

      expect(mockedAxios.get).toHaveBeenCalledWith('/alert/status?limit=10&offset=10', {
        headers: undefined,
      });
      expect(result.data).toEqual(mockStatus);
      expect(result.total).toBe(1);
    });

    it('should fetch alert history list', async () => {
      const mockHistory = [
        {
          id: 'alert-1',
          alert_id: 'alert-id-1',
          alert_type: 'query' as AlertType,
          name: 'Test Rule',
          status: 'alerting' as AlertStatus,
          severity: 'Critical' as AlertSeverity,
          value: 100,
          mask: false,
          timestamp: new Date('2025-10-14T00:00:00Z'),
          updated_at: new Date('2025-10-14T00:00:00Z'),
        },
      ];

      mockedAxios.get.mockResolvedValueOnce({ data: mockHistory });

      const result = await alertProvider.getList({
        resource: ALERT_RESOURCES.HIST,
        filters: [
          { field: 'name', operator: 'eq', value: 'Test Rule' },
          { field: 'startTime', operator: 'eq', value: '2025-10-13T00:00:00Z' },
          { field: 'endTime', operator: 'eq', value: '2025-10-14T00:00:00Z' },
          { field: 'count', operator: 'eq', value: 100 },
        ],
      });

      expect(mockedAxios.get).toHaveBeenCalledWith(
        '/alert/hist?name=Test+Rule&start-time=2025-10-13T00%3A00%3A00Z&end-time=2025-10-14T00%3A00%3A00Z&count=100',
        {
          headers: undefined,
        },
      );
      expect(result.data).toEqual(mockHistory);
      expect(result.total).toBe(1);
    });

    it('should throw error for unsupported resource', async () => {
      await expect(
        alertProvider.getList({
          resource: 'unsupported',
        }),
      ).rejects.toThrow('Unsupported resource: unsupported');
    });
  });

  describe('getOne', () => {
    it('should fetch single alert rule', async () => {
      const mockRule = {
        alert_type: 'query' as AlertType,
        rule: {
          id: 'rule-1',
          name: 'Test Rule',
          datasource: 'clickhouse',
        },
        status: [],
      };

      mockedAxios.get.mockResolvedValueOnce({ data: mockRule });

      const result = await alertProvider.getOne({
        resource: ALERT_RESOURCES.RULE,
        id: 'rule-1',
      });

      expect(mockedAxios.get).toHaveBeenCalledWith('/alert/rule/rule-1', {
        headers: undefined,
      });
      expect(result.data).toEqual(mockRule);
    });

    it('should throw error for unsupported resource', async () => {
      await expect(
        alertProvider.getOne({
          resource: 'unsupported',
          id: '1',
        }),
      ).rejects.toThrow('Unsupported resource: unsupported');
    });
  });

  describe('create', () => {
    it('should create new alert rule', async () => {
      const newRule: CreateAlertRuleRequest = {
        alert_type: 'query',
        rule: {
          name: 'New Rule',
          datasource: 'clickhouse',
          datasource_query: {
            query: 'SELECT * FROM table',
            time_label: 'time',
            variable_label: 'value',
          },
          threshold: [
            {
              id: 'threshold-1',
              severity: 'Critical',
              condition: 'is_above',
              start_value: 100,
            },
          ],
          evaluation_interval: '*/5 * * * *',
        },
      };

      const mockResponse = { id: 'rule-1' };
      mockedAxios.post.mockResolvedValueOnce({ data: mockResponse });

      const result = await alertProvider.create({
        resource: ALERT_RESOURCES.RULE,
        variables: newRule,
      });

      expect(mockedAxios.post).toHaveBeenCalledWith('/alert/rule', newRule, {
        headers: undefined,
      });
      expect(result.data).toEqual(mockResponse);
    });

    it('should test alert query', async () => {
      const queryRequest = {
        datasource: 'clickhouse',
        query: {
          query: 'SELECT * FROM table',
          variables: {
            [QUERY_PARAM_START_TIME]: 1697222400,
            [QUERY_PARAM_END_TIME]: 1697308800,
          } as any,
        },
      } as AlertQueryRequest;

      const mockResponse = {
        columns: [
          { name: 'time', type: 'datetime' as const },
          { name: 'value', type: 'number' as const },
        ],
        data: [
          ['2025-10-14T00:00:00Z', 100],
          ['2025-10-14T01:00:00Z', 200],
        ],
        rowCount: 2,
      };

      mockedAxios.post.mockResolvedValueOnce({ data: mockResponse });

      const result = await alertProvider.create({
        resource: ALERT_RESOURCES.QUERY,
        variables: queryRequest,
      });

      expect(mockedAxios.post).toHaveBeenCalledWith('/alert/query', queryRequest, {
        headers: undefined,
      });
      expect(result.data).toEqual(mockResponse);
    });

    it('should throw error for unsupported resource', async () => {
      await expect(
        alertProvider.create({
          resource: 'unsupported',
          variables: {},
        }),
      ).rejects.toThrow('Unsupported resource: unsupported');
    });
  });

  describe('update', () => {
    it('should update alert rule', async () => {
      const updateRule: UpdateAlertRuleRequest = {
        alert_type: 'query',
        rule: {
          name: 'Updated Rule',
          datasource: 'clickhouse',
          datasource_query: {
            query: 'SELECT * FROM table',
            time_label: 'time',
            variable_label: 'value',
          },
          threshold: [
            {
              id: 'threshold-1',
              severity: 'Major',
              condition: 'is_above',
              start_value: 50,
            },
          ],
          evaluation_interval: '*/10 * * * *',
        },
      };

      mockedAxios.put.mockResolvedValueOnce({ data: {} });

      const result = await alertProvider.update({
        resource: ALERT_RESOURCES.RULE,
        id: 'rule-1',
        variables: updateRule,
      });

      expect(mockedAxios.put).toHaveBeenCalledWith('/alert/rule/rule-1', updateRule, {
        headers: undefined,
      });
      expect(result.data).toEqual({});
    });

    it('should mask/unmask alert status', async () => {
      const maskRequest: AlertStatusMaskRequest = {
        mask: true,
      };

      mockedAxios.put.mockResolvedValueOnce({ data: {} });

      const result = await alertProvider.update({
        resource: ALERT_RESOURCES.STATUS_MASK,
        id: 'Test Rule/alert-id-1',
        variables: maskRequest,
      });

      expect(mockedAxios.put).toHaveBeenCalledWith(
        '/alert/status/Test Rule/alert-id-1/mask',
        maskRequest,
        {
          headers: undefined,
        },
      );
      expect(result.data).toEqual({});
    });

    it('should throw error for unsupported resource', async () => {
      await expect(
        alertProvider.update({
          resource: 'unsupported',
          id: '1',
          variables: {},
        }),
      ).rejects.toThrow('Unsupported resource: unsupported');
    });
  });

  describe('deleteOne', () => {
    it('should delete alert rule', async () => {
      mockedAxios.delete.mockResolvedValueOnce({ data: {} });

      const result = await alertProvider.deleteOne({
        resource: ALERT_RESOURCES.RULE,
        id: 'rule-1',
      });

      expect(mockedAxios.delete).toHaveBeenCalledWith('/alert/rule/rule-1', {
        headers: undefined,
      });
      expect(result.data).toEqual({});
    });

    it('should delete alert status', async () => {
      mockedAxios.delete.mockResolvedValueOnce({ data: {} });

      const result = await alertProvider.deleteOne({
        resource: ALERT_RESOURCES.STATUS,
        id: 'Test Rule/alert-id-1',
      });

      expect(mockedAxios.delete).toHaveBeenCalledWith('/alert/status/Test Rule/alert-id-1', {
        headers: undefined,
      });
      expect(result.data).toEqual({});
    });

    it('should throw error for unsupported resource', async () => {
      await expect(
        alertProvider.deleteOne({
          resource: 'unsupported',
          id: '1',
        }),
      ).rejects.toThrow('Unsupported resource: unsupported');
    });
  });

  describe('getApiUrl', () => {
    it('should return correct API URL', () => {
      expect(alertProvider.getApiUrl?.()).toBe('/alert');
    });
  });
});
