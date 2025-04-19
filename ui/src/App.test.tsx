import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import App from './App';
import '@testing-library/jest-dom';

// Mock the child components
jest.mock('./components/Layout', () => {
  return {
    __esModule: true,
    default: ({ children }: { children: React.ReactNode }) => <div data-testid="mock-layout">{children}</div>
  };
});

jest.mock('./pages/HomePage', () => {
  return {
    __esModule: true,
    default: () => <div data-testid="mock-home-page">Home Page</div>
  };
});

jest.mock('./pages/ConfigurationDetailPage', () => {
  return {
    __esModule: true,
    default: () => <div data-testid="mock-config-detail-page">Configuration Detail Page</div>
  };
});

jest.mock('./pages/NotFoundPage', () => {
  return {
    __esModule: true,
    default: () => <div data-testid="mock-not-found-page">404 Page Not Found</div>
  };
});

// Mock the BrowserRouter component from react-router-dom
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  BrowserRouter: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

describe('App Component', () => {
  const renderWithRouter = (initialEntries = ['/', '/configurations/owner/config-name']) => {
    return render(
      <MemoryRouter initialEntries={initialEntries}>
        <App />
      </MemoryRouter>
    );
  };

  it('renders the layout component', () => {
    renderWithRouter();
    expect(screen.getByTestId('mock-layout')).toBeInTheDocument();
  });

  it('renders home page on root path', () => {
    renderWithRouter(['/']);
    expect(screen.getByTestId('mock-home-page')).toBeInTheDocument();
  });

  it('renders configuration detail page on configuration path', () => {
    renderWithRouter(['/configurations/owner/config-name']);
    expect(screen.getByTestId('mock-config-detail-page')).toBeInTheDocument();
  });

  it('renders not found page for unknown routes', () => {
    renderWithRouter(['/unknown-route']);
    expect(screen.getByTestId('mock-not-found-page')).toBeInTheDocument();
  });
}); 