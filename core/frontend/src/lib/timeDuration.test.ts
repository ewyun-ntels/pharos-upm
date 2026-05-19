import timeDuration from './timeDuration';

describe('timeDuration', () => {
  test('difference of less than 3 minutes', () => {
    const result = timeDuration('2024-06-21T09:53:06Z', '2024-06-21T09:55:00Z');
    expect(result).toBe('1m54s');
  });

  test('difference of less than an hour', () => {
    const result = timeDuration('2024-06-21T09:00:00Z', '2024-06-21T09:45:00Z');
    expect(result).toBe('45m');
  });

  test('difference of less than 3 hours', () => {
    const result = timeDuration('2024-06-21T09:00:00Z', '2024-06-21T11:30:00Z');
    expect(result).toBe('2h30m');
  });

  test('difference of less than 1 day', () => {
    const result = timeDuration('2024-06-21T09:00:00Z', '2024-06-21T23:30:00Z');
    expect(result).toBe('14h');
  });

  test('difference of less than 2 day', () => {
    const result = timeDuration('2024-06-21T09:00:00Z', '2024-06-23T23:30:00Z');
    expect(result).toBe('2d14h');
  });

  test('difference of less than 7 days', () => {
    const result = timeDuration('2024-06-14T09:00:00Z', '2024-06-21T09:00:00Z');
    expect(result).toBe('7d');
  });

  test('difference of less than 7 days 3 hour', () => {
    const result = timeDuration('2024-06-14T09:00:00Z', '2024-06-21T12:00:00Z');
    expect(result).toBe('7d');
  });

  test('difference of more than 7 days', () => {
    const result = timeDuration('2024-06-01T09:00:00Z', '2024-06-21T09:00:00Z');
    expect(result).toBe('20d');
  });
});
