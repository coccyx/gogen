import { render, screen } from '../utils/test-utils';
import Layout from './Layout';

// Mock the Header component
jest.mock('./Header', () => {
  return function MockHeader() {
    return <header data-testid="mock-header">Header Content</header>;
  };
});

// Mock the Footer component
jest.mock('./Footer', () => {
  return function MockFooter() {
    return <footer data-testid="mock-footer">Footer Content</footer>;
  };
});

describe('Layout', () => {
  it('renders all main sections', () => {
    render(
      <Layout>
        <div>Test Content</div>
      </Layout>
    );
    
    expect(screen.getByTestId('mock-header')).toBeInTheDocument();
    expect(screen.getByText('Test Content')).toBeInTheDocument();
    expect(screen.getByTestId('mock-footer')).toBeInTheDocument();
  });

  it('has correct layout structure and styling', () => {
    render(
      <Layout>
        <div>Test Content</div>
      </Layout>
    );
    
    const container = screen.getByTestId('layout-container');
    expect(container).toHaveClass('min-h-screen', 'bg-gray-100', 'flex', 'flex-col');
  });
}); 