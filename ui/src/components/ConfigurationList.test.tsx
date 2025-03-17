import { render, screen, fireEvent } from '../utils/test-utils';
import ConfigurationList from './ConfigurationList';

describe('ConfigurationList', () => {
  const mockConfigs = [
    { gogen: 'test-config-1', description: 'Test description 1' },
    { gogen: 'test-config-2', description: 'Test description 2' }
  ];

  it('displays loading spinner when loading', () => {
    render(<ConfigurationList configurations={[]} loading={true} error={null} />);
    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('displays error message when there is an error', () => {
    const errorMessage = 'Failed to load configurations';
    render(<ConfigurationList configurations={[]} loading={false} error={errorMessage} />);
    expect(screen.getByText(errorMessage)).toBeInTheDocument();
  });

  it('displays message when no configurations are available', () => {
    render(<ConfigurationList configurations={[]} loading={false} error={null} />);
    expect(screen.getByText('No configurations found matching your search.')).toBeInTheDocument();
  });

  it('renders configurations list', () => {
    render(<ConfigurationList configurations={mockConfigs} loading={false} error={null} />);
    
    // Check if both configurations are rendered
    expect(screen.getByText('test-config-1')).toBeInTheDocument();
    expect(screen.getByText('test-config-2')).toBeInTheDocument();
    expect(screen.getByText('Test description 1')).toBeInTheDocument();
    expect(screen.getByText('Test description 2')).toBeInTheDocument();
  });

  it('renders default text when description is missing', () => {
    const configWithoutDescription = [
      { gogen: 'test-config-1', description: '' }
    ];

    render(<ConfigurationList configurations={configWithoutDescription} loading={false} error={null} />);
    expect(screen.getByText('-')).toBeInTheDocument();
  });

  it('handles search input', () => {
    render(<ConfigurationList configurations={mockConfigs} loading={false} error={null} />);
    
    const searchInput = screen.getByPlaceholderText('Search configurations...');
    fireEvent.change(searchInput, { target: { value: 'test-config-1' } });
    
    expect(screen.getByText('test-config-1')).toBeInTheDocument();
    expect(screen.queryByText('test-config-2')).not.toBeInTheDocument();
  });

  it('navigates to configuration detail', () => {
    render(<ConfigurationList configurations={mockConfigs} loading={false} error={null} />);
    
    const firstConfig = mockConfigs[0];
    const configLink = screen.getByText(firstConfig.gogen).closest('a');
    expect(configLink).toHaveAttribute('href', `/configurations/${firstConfig.gogen}`);
  });
}); 