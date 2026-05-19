import React from 'react';
import { render, screen } from '@testing-library/react';
import CellAge from './age';

describe('CellAge', () => {
  const fixedNow = new Date('2024-06-21T09:55:00Z');

  beforeAll(() => {
    jest.useFakeTimers({ now: fixedNow });
  });

  afterAll(() => {
    jest.useRealTimers();
  });

  it('renders human-friendly duration from value to now', () => {
    const getValue = () => '2024-06-21T09:53:06Z';

    render(<CellAge {...({ getValue } as any)} />);

    expect(screen.getByText('1m54s')).toBeInTheDocument();
  });

  it('supports Date instance as value', () => {
    const fiveMinutesAgo = new Date(fixedNow.getTime() - 5 * 60 * 1000);
    const getValue = () => fiveMinutesAgo as any;

    render(<CellAge {...({ getValue } as any)} />);

    expect(screen.getByText('5m')).toBeInTheDocument();
  });
});
