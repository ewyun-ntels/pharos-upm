import React from 'react';
import {render, screen, fireEvent} from '@testing-library/react';
import {GeneralSettingDialog} from './generalSettingDialog';
import type {DashboardData} from '@pharos/shared/types/dashboard';

// Mock UI components
jest.mock('@pharos/shared/components/ui/dialog', () => ({
  Dialog: ({children, open}: any) => (open ? <div data-testid="dialog">{children}</div> : null),
  DialogContent: ({children}: any) => <div data-testid="dialog-content">{children}</div>,
  DialogHeader: ({children}: any) => <div data-testid="dialog-header">{children}</div>,
  DialogTitle: ({children}: any) => <div data-testid="dialog-title">{children}</div>,
  DialogFooter: ({children}: any) => <div data-testid="dialog-footer">{children}</div>,
}));

jest.mock('@pharos/shared/components/ui/button', () => ({
  Button: ({children, onClick, disabled, ...props}: any) => (
    <button onClick={onClick} disabled={disabled} {...props}>
      {children}
    </button>
  ),
}));

jest.mock('@pharos/shared/components/ui/input', () => ({
  Input: ({value, onChange, ...props}: any) => (
    <input value={value || ''} onChange={onChange} {...props} />
  ),
}));

jest.mock('@pharos/shared/components/ui-extension/toggle-icon-button', () => ({
  ToggleIconButton: ({toggled, onClick}: any) => (
    <button onClick={onClick} data-testid="toggle-favorite" data-toggled={toggled}>
      {toggled ? 'Favorited' : 'Not Favorited'}
    </button>
  ),
}));

jest.mock('@pharos/shared/components', () => ({
  Star: () => <span>★</span>,
}));

const makeData = (overrides: Partial<DashboardData['config']> = {}): DashboardData => ({
  id: 'test-id',
  permission: 'owner',
  config: {
    title: 'Test Title',
    displayName: 'Test Display Name',
    description: '',
    favorite: false,
    filters: [],
    type: 'default',
    panels: [],
    ...overrides,
  },
});

describe('GeneralSettingDialog', () => {
  const defaultProps = {
    open: true,
    onOpenChange: jest.fn(),
    dialogTitle: 'Test Dialog',
    data: makeData(),
    onSave: jest.fn(),
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render dialog when open is true', () => {
      render(<GeneralSettingDialog {...defaultProps} />);

      expect(screen.getByTestId('dialog')).toBeInTheDocument();
      expect(screen.getByTestId('dialog-title')).toHaveTextContent('Test Dialog');
    });

    it('should not render dialog when open is false', () => {
      render(<GeneralSettingDialog {...defaultProps} open={false} />);

      expect(screen.queryByTestId('dialog')).not.toBeInTheDocument();
    });

    it('should render title and displayName with initial values', () => {
      render(<GeneralSettingDialog {...defaultProps} />);

      expect(screen.getByRole('textbox', {name: /title/i})).toHaveValue('Test Title');
      expect(screen.getByRole('textbox', {name: /display name/i})).toHaveValue('Test Display Name');
      expect(screen.getByTestId('toggle-favorite')).toHaveAttribute('data-toggled', 'false');
    });

    it('should render favorite as true when config.favorite is true', () => {
      render(<GeneralSettingDialog {...defaultProps} data={makeData({favorite: true})} />);

      expect(screen.getByTestId('toggle-favorite')).toHaveAttribute('data-toggled', 'true');
    });
  });

  describe('Form Validation', () => {
    it('should disable Apply button when title is empty', () => {
      render(<GeneralSettingDialog {...defaultProps} />);

      fireEvent.change(screen.getByRole('textbox', {name: /title/i}), {target: {value: ''}});

      expect(screen.getByRole('button', {name: /apply/i})).toBeDisabled();
      expect(screen.getByText('Title을 입력해주세요.')).toBeInTheDocument();
    });

    it('should enable Apply button when title is filled', () => {
      render(<GeneralSettingDialog {...defaultProps} />);

      expect(screen.getByRole('button', {name: /apply/i})).not.toBeDisabled();
    });
  });

  describe('Form Interactions', () => {
    it('should update title input value', () => {
      render(<GeneralSettingDialog {...defaultProps} />);

      fireEvent.change(screen.getByRole('textbox', {name: /title/i}), {target: {value: 'New Title'}});

      expect(screen.getByRole('textbox', {name: /title/i})).toHaveValue('New Title');
    });

    it('should update displayName input value', () => {
      render(<GeneralSettingDialog {...defaultProps} />);

      fireEvent.change(screen.getByRole('textbox', {name: /display name/i}), {target: {value: 'New Display Name'}});

      expect(screen.getByRole('textbox', {name: /display name/i})).toHaveValue('New Display Name');
    });

    it('should toggle favorite status', () => {
      render(<GeneralSettingDialog {...defaultProps} />);

      const favoriteToggle = screen.getByTestId('toggle-favorite');
      expect(favoriteToggle).toHaveAttribute('data-toggled', 'false');

      fireEvent.click(favoriteToggle);

      expect(favoriteToggle).toHaveAttribute('data-toggled', 'true');
    });
  });

  describe('Data Prop Updates', () => {
    it('should update form values when data prop changes', () => {
      const {rerender} = render(<GeneralSettingDialog {...defaultProps} />);

      expect(screen.getByRole('textbox', {name: /title/i})).toHaveValue('Test Title');

      rerender(<GeneralSettingDialog {...defaultProps} data={makeData({title: 'Updated Title', favorite: true})} />);

      expect(screen.getByRole('textbox', {name: /title/i})).toHaveValue('Updated Title');
      expect(screen.getByTestId('toggle-favorite')).toHaveAttribute('data-toggled', 'true');
    });

    it('should handle null data gracefully', () => {
      render(<GeneralSettingDialog {...defaultProps} data={null} />);

      const inputs = screen.getAllByRole('textbox');
      expect(inputs[0]).toHaveValue(''); // title
      expect(inputs[1]).toHaveValue(''); // displayName
    });
  });

  describe('Button Actions', () => {
    it('should call onOpenChange when Cancel button is clicked', () => {
      render(<GeneralSettingDialog {...defaultProps} />);

      fireEvent.click(screen.getByRole('button', {name: /cancel/i}));

      expect(defaultProps.onOpenChange).toHaveBeenCalledTimes(1);
    });

    it('should call onSave with id and updated config when Apply is clicked', () => {
      render(<GeneralSettingDialog {...defaultProps} />);

      fireEvent.change(screen.getByRole('textbox', {name: /title/i}), {target: {value: 'Modified Title'}});
      fireEvent.change(screen.getByRole('textbox', {name: /display name/i}), {target: {value: 'Modified Display Name'}});
      fireEvent.click(screen.getByTestId('toggle-favorite'));
      fireEvent.click(screen.getByRole('button', {name: /apply/i}));

      expect(defaultProps.onSave).toHaveBeenCalledWith(
        'test-id',
        expect.objectContaining({
          title: 'Modified Title',
          displayName: 'Modified Display Name',
          favorite: true,
        }),
      );
      expect(defaultProps.onOpenChange).toHaveBeenCalledTimes(1);
    });

    it('should not call onSave when title is empty', () => {
      render(<GeneralSettingDialog {...defaultProps} />);

      fireEvent.change(screen.getByRole('textbox', {name: /title/i}), {target: {value: ''}});
      fireEvent.click(screen.getByRole('button', {name: /apply/i}));

      expect(defaultProps.onSave).not.toHaveBeenCalled();
    });

    it('should not call onSave when data is null', () => {
      render(<GeneralSettingDialog {...defaultProps} data={null} />);

      // Apply button is disabled (title is empty), so onSave won't be called
      expect(screen.getByRole('button', {name: /apply/i})).toBeDisabled();
    });
  });
});
