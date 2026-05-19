import React from 'react';
import { render } from '@testing-library/react';
import { PageBreadcrumb } from './breadcrumb';

// Mock lucide-react icons used in breadcrumb
jest.mock('lucide-react', () => ({
  Folder: () => <span data-testid="folder-icon" />,
}));

// Mock react-router-dom Link to a simple anchor
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  Link: ({ to, children }: any) => <a href={to}>{children}</a>,
}));

// Mock shadcn breadcrumb ui primitives to simple passthrough elements
jest.mock('@pharos/shared/components/ui/breadcrumb', () => ({
  __esModule: true,
  Breadcrumb: ({ children }: any) => <nav data-testid="breadcrumb">{children}</nav>,
  BreadcrumbList: ({ children }: any) => <ol>{children}</ol>,
  BreadcrumbItem: ({ children }: any) => <li>{children}</li>,
  BreadcrumbSeparator: () => <span aria-hidden>/</span>,
  BreadcrumbLink: ({ children }: any) => <>{children}</>,
  BreadcrumbPage: ({ children }: any) => <span aria-current="page">{children}</span>,
  useSidebar: () => ({ open: true }),
}));

// Mock useSidebar used directly in PageBreadcrumb
jest.mock('@pharos/shared/components/ui', () => ({
  ...jest.requireActual('@pharos/shared/components/ui'),
  useSidebar: () => ({ open: true }),
}));

// Control useBreadcrumb return value per test
let mockBreadcrumbs: Array<{ label: string; href?: string; isFolder?: boolean }>;

jest.mock('@/lib/data-provider/nav', () => ({
  useBreadcrumb: (_currentLabel?: string) => ({ breadcrumbs: mockBreadcrumbs }),
}));

describe('PageBreadcrumb', () => {
  beforeEach(() => {
    mockBreadcrumbs = [];
  });

  it('renders nothing when breadcrumbs is empty', () => {
    const { queryByTestId } = render(<PageBreadcrumb />);
    expect(queryByTestId('breadcrumb')).not.toBeInTheDocument();
  });

  it('renders breadcrumb items with links and current page', () => {
    mockBreadcrumbs = [
      { label: 'Settings', href: '/settings' },
      { label: 'Roles' },
    ];

    const { getByRole, getByText } = render(<PageBreadcrumb />);

    expect(getByRole('link', { name: 'Settings' })).toHaveAttribute('href', '/settings');
    expect(getByText('Roles')).toBeInTheDocument();
  });

  it('renders a folder icon and muted text for an isFolder crumb without href', () => {
    mockBreadcrumbs = [
      { label: 'Dashboards', href: '/dashboards' },
      { label: 'My Folder', isFolder: true },
      { label: 'My Dashboard' },
    ];

    const { getByText, queryByRole } = render(<PageBreadcrumb />);

    // The folder crumb should not be a link
    expect(queryByRole('link', { name: 'My Folder' })).not.toBeInTheDocument();
    // The folder crumb label should be present
    expect(getByText('My Folder')).toBeInTheDocument();
    // The folder icon should be rendered
    expect(document.querySelector('[data-testid="folder-icon"]')).toBeInTheDocument();
    // The last crumb is the current page
    expect(getByText('My Dashboard')).toBeInTheDocument();
  });

  it('overrides the last breadcrumb label when currentLabel is provided', () => {
    mockBreadcrumbs = [
      { label: 'Settings', href: '/settings' },
      { label: 'Add Composite Role' },
    ];

    const { getByRole, getByText, queryByText } = render(<PageBreadcrumb currentLabel="Add Composite Role" />);

    expect(getByText('Add Composite Role')).toBeInTheDocument();
    expect(queryByText('Roles')).not.toBeInTheDocument();
    expect(getByRole('link', { name: 'Settings' })).toHaveAttribute('href', '/settings');
  });
});

