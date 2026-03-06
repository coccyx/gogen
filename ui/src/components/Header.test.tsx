import { render, screen } from '../utils/test-utils';
import { MemoryRouter } from 'react-router-dom';
import Header from './Header';

jest.mock('../context/AuthContext', () => ({
  useAuth: () => ({
    user: null,
    isAuthenticated: false,
    logout: jest.fn(),
    isLoading: false,
  }),
}));

describe('Header', () => {
  const renderWithRouter = () => {
    render(
      <MemoryRouter>
        <Header />
      </MemoryRouter>
    );
  };

  it('renders logo text', () => {
    renderWithRouter();
    
    const logo = screen.getByText('gogen');
    expect(logo).toBeInTheDocument();
    expect(logo).toHaveAttribute('href', '/');
    expect(logo).toHaveClass('font-mono', 'text-xl', 'font-semibold', 'no-underline', 'text-term-text');
  });

  it('renders navigation links', () => {
    renderWithRouter();
    
    const homeLink = screen.getByText('Home');
    expect(homeLink).toBeInTheDocument();
    expect(homeLink).toHaveAttribute('href', '/');
    expect(homeLink).toHaveClass('text-sm', 'hover:text-term-green', 'transition-colors');
    expect(screen.getByText('Login')).toHaveAttribute('href', '/login');
  });

  it('has correct layout structure and styling', () => {
    renderWithRouter();
    
    // Check header styling
    const header = screen.getByRole('banner');
    expect(header).toHaveClass('bg-term-bg-elevated', 'text-term-text', 'px-4', 'py-2', 'border-b', 'border-term-border');

    // Check container styling
    const container = header.firstElementChild;
    expect(container).toHaveClass('container', 'mx-auto', 'px-4', 'flex', 'justify-between', 'items-center');

    // Check navigation styling
    const nav = screen.getByRole('navigation');
    expect(nav).toBeInTheDocument();
    expect(nav.querySelector('ul')).toHaveClass('flex', 'items-center', 'space-x-4');
  });
});
