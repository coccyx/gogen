import { render, screen, waitFor } from '../utils/test-utils';
import { MemoryRouter } from 'react-router-dom';
import HomePage from './HomePage';
import gogenApi from '../api/gogenApi';

// Mock the gogenApi module
jest.mock('../api/gogenApi', () => ({
  __esModule: true,
  default: {
    listConfigurations: jest.fn()
  }
}));

// Mock the Hero component
jest.mock('../components/Hero', () => {
  return {
    __esModule: true,
    default: () => <div data-testid="mock-hero">Hero Component</div>
  }
});

// Mock the ConfigurationList component
jest.mock('../components/ConfigurationList', () => {
  return {
    __esModule: true,
    default: ({ configurations, loading, error }: { configurations: any[], loading: boolean, error: string | null }) => (
      <div data-testid="mock-configuration-list">
        {loading && <div key="loading">Loading...</div>}
        {error && <div key="error">Failed to load configurations</div>}
        {configurations.map((config: any) => (
          <div key={config.gogen}>{config.gogen}</div>
        ))}
      </div>
    )
  }
});

describe('HomePage', () => {
  const mockConfigurations = [
    { gogen: 'test1', description: 'Test 1 Description' },
    { gogen: 'test2', description: 'Test 2 Description' }
  ];

  const renderWithRouter = () => {
    render(
      <MemoryRouter>
        <HomePage />
      </MemoryRouter>
    );
  };

  beforeEach(() => {
    // Clear all mocks before each test
    jest.clearAllMocks();
  });

  it('renders Hero and ConfigurationList components', () => {
    renderWithRouter();
    
    expect(screen.getByTestId('mock-hero')).toBeInTheDocument();
    expect(screen.getByTestId('mock-configuration-list')).toBeInTheDocument();
    expect(screen.getByText('Configurations')).toBeInTheDocument();
  });

  it('shows loading state initially', async () => {
    // Mock API to delay response
    (gogenApi.listConfigurations as jest.Mock).mockImplementation(
      () => new Promise(resolve => setTimeout(resolve, 100))
    );

    renderWithRouter();
    
    expect(screen.getByText('Loading...')).toBeInTheDocument();
  });

  it('displays configurations after successful fetch', async () => {
    // Mock successful API response
    (gogenApi.listConfigurations as jest.Mock).mockResolvedValue(mockConfigurations);

    renderWithRouter();

    // Wait for configurations to be displayed
    await waitFor(() => {
      expect(screen.getByText('test1')).toBeInTheDocument();
      expect(screen.getByText('test2')).toBeInTheDocument();
    });

    // Verify API was called
    expect(gogenApi.listConfigurations).toHaveBeenCalledTimes(1);
  });

  it('displays error message when fetch fails', async () => {
    // Mock API error
    (gogenApi.listConfigurations as jest.Mock).mockRejectedValue(new Error('API Error'));

    renderWithRouter();

    // Wait for error message to be displayed
    await waitFor(() => {
      expect(screen.getByText('Failed to load configurations')).toBeInTheDocument();
    });

    // Verify API was called
    expect(gogenApi.listConfigurations).toHaveBeenCalledTimes(1);
  });

  it('handles empty configuration list', async () => {
    // Mock empty API response
    (gogenApi.listConfigurations as jest.Mock).mockResolvedValue([]);

    renderWithRouter();

    // Wait for loading to finish
    await waitFor(() => {
      expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
    });

    // Verify empty state is handled
    expect(screen.queryByText(/test\d/)).not.toBeInTheDocument();
  });
}); 