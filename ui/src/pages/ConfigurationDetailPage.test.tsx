import { render, screen, waitFor } from '../utils/test-utils';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import ConfigurationDetailPage from './ConfigurationDetailPage';
import gogenApi from '../api/gogenApi';
import type { Configuration } from '../api/gogenApi';

// Mock the Monaco Editor
jest.mock('@monaco-editor/react', () => {
  return function MockEditor({ value }: { value: string }) {
    return <div className="monaco-editor" data-testid="mock-editor">{value}</div>;
  };
});

// Mock the gogenApi
jest.mock('../api/gogenApi', () => {
  return {
    __esModule: true,
    default: {
      getConfiguration: jest.fn(),
      listConfigurations: jest.fn(),
      searchConfigurations: jest.fn()
    }
  };
});

const mockConfiguration: Configuration = {
  gogen: 'testconfig',
  description: 'Test Configuration',
  config: 'test: yaml',
};

const renderWithRouter = (owner: string, configName: string) => {
  return render(
    <MemoryRouter initialEntries={[`/configurations/${owner}/${configName}`]}>
      <Routes>
        <Route path="/configurations/:owner/:configName" element={<ConfigurationDetailPage />} />
      </Routes>
    </MemoryRouter>
  );
};

describe('ConfigurationDetailPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // Reset mock implementation before each test
    (gogenApi.getConfiguration as jest.Mock).mockReset();
  });

  it('shows loading state initially', async () => {
    // Create a promise that never resolves to keep the loading state
    (gogenApi.getConfiguration as jest.Mock).mockImplementation(() => new Promise(() => {}));
    
    renderWithRouter('testowner', 'testconfig');
    expect(await screen.findByRole('status')).toBeInTheDocument();
  });

  it('displays configuration details when loaded', async () => {
    (gogenApi.getConfiguration as jest.Mock).mockResolvedValue(mockConfiguration);
    renderWithRouter('testowner', 'testconfig');

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'testconfig' })).toBeInTheDocument();
      expect(screen.getByText('Test Configuration')).toBeInTheDocument();
      expect(screen.getByTestId('mock-editor')).toHaveTextContent('test: yaml');
    });
  });

  it('displays error when API call fails', async () => {
    (gogenApi.getConfiguration as jest.Mock).mockRejectedValue(new Error('API Error'));
    renderWithRouter('testowner', 'testconfig');

    await waitFor(() => {
      expect(screen.getByText('Failed to load configuration. Please try again later.')).toBeInTheDocument();
    });
  });

  it('has correct styling classes', async () => {
    (gogenApi.getConfiguration as jest.Mock).mockResolvedValue(mockConfiguration);
    renderWithRouter('testowner', 'testconfig');

    await waitFor(() => {
      // Check main container
      const mainContainer = screen.getByRole('main');
      expect(mainContainer).toHaveClass('container', 'mx-auto', 'px-4', 'py-8');

      // Check title
      const title = screen.getByRole('heading', { level: 1 });
      expect(title).toHaveClass('text-3xl', 'font-bold', 'text-gray-800');

      // Check back link
      const backLink = screen.getByRole('link', { name: 'Back to List' });
      expect(backLink).toHaveClass('btn-primary');
    });
  });
}); 