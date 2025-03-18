import { render, screen } from '../utils/test-utils';
import { MemoryRouter } from 'react-router-dom';
import ConfigurationList from './ConfigurationList';
import { ConfigurationSummary } from '../api/gogenApi';

const mockConfigurations: ConfigurationSummary[] = [
  {
    gogen: 'test-config',
    description: 'Test Configuration',
  },
  {
    gogen: 'another-config',
    description: 'Another Configuration',
  },
];

const renderWithRouter = (props: {
  configurations: ConfigurationSummary[];
  loading: boolean;
  error: string | null;
}) => {
  return render(
    <MemoryRouter>
      <ConfigurationList {...props} />
    </MemoryRouter>
  );
};

describe('ConfigurationList', () => {
  it('shows loading state', () => {
    renderWithRouter({ configurations: [], loading: true, error: null });
    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('shows error message', () => {
    const errorMessage = 'Failed to load configurations';
    renderWithRouter({ configurations: [], loading: false, error: errorMessage });
    expect(screen.getByText(errorMessage)).toBeInTheDocument();
  });

  it('renders list of configurations', () => {
    renderWithRouter({ configurations: mockConfigurations, loading: false, error: null });
    
    mockConfigurations.forEach((config) => {
      expect(screen.getByText(config.gogen)).toBeInTheDocument();
      expect(screen.getByText(config.description)).toBeInTheDocument();
    });
  });

  it('handles empty configurations list', () => {
    renderWithRouter({ configurations: [], loading: false, error: null });
    expect(screen.getByText('No configurations found matching your search.')).toBeInTheDocument();
  });

  it('has correct styling classes', () => {
    renderWithRouter({ configurations: mockConfigurations, loading: false, error: null });
    
    // Check table styling
    const table = screen.getByRole('table');
    expect(table).toHaveClass('min-w-full');
    
    // Check header styling
    const headers = screen.getAllByRole('columnheader');
    headers.forEach(header => {
      expect(header).toHaveClass('px-6', 'py-3', 'text-left', 'text-xs', 'font-medium', 'text-gray-500', 'uppercase', 'tracking-wider');
    });
    
    // Check table container styling
    const tableContainer = screen.getByRole('table').closest('div');
    expect(tableContainer).toHaveClass('bg-white', 'rounded-lg', 'shadow', 'overflow-hidden');
  });
}); 