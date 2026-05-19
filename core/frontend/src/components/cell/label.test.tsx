import React from 'react';
import {render, screen} from '@testing-library/react';
import CellLabel from './label';

// Mock BadgeCustom to a simple span for easier assertions
jest.mock('@pharos/shared/components/ui-extension/badge-custom', () => ({
  __esModule: true,
  BadgeCustom: ({children}: any) => <span data-testid="badge">{children}</span>,
}));

describe('CellLabel', () => {
  const renderWithValue = (value: any, columnId = 'col') => {
    const getValue = () => value;
    const column = {id: columnId} as any;
    return render(<CellLabel {...({getValue, column} as any)} />);
  };

  it('renders badges for an array of up to 10 labels without ellipsis', () => {
    const values = ['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'];
    renderWithValue(values);

    const badges = screen.getAllByTestId('badge');
    expect(badges).toHaveLength(10);
    expect(screen.queryByText('...')).not.toBeInTheDocument();
    expect(badges[0]).toHaveTextContent('a');
    expect(badges[9]).toHaveTextContent('j');
  });

  it('renders first 10 badges and ellipsis when array is longer than 10', () => {
    const values = Array.from({length: 12}, (_, i) => `v${i + 1}`);
    renderWithValue(values);

    const badges = screen.getAllByTestId('badge');
    expect(badges).toHaveLength(10);
    expect(badges[0]).toHaveTextContent('v1');
    expect(badges[9]).toHaveTextContent('v10');
    expect(screen.getByText('...')).toBeInTheDocument();
  });

  it('renders key=value pairs when value is an object', () => {
    const value = {a: 1, b: 'two', c: true};
    renderWithValue(value);

    const badges = screen.getAllByTestId('badge');
    const texts = badges.map((b) => b.textContent);
    expect(texts).toEqual(expect.arrayContaining(['a=1', 'b=two', 'c=true']));
  });
});
