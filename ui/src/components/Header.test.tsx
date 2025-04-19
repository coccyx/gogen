import { render, screen } from '../utils/test-utils';
import { MemoryRouter } from 'react-router-dom';
import Header from './Header';

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
    
    const logo = screen.getByText('Gogen UI');
    expect(logo).toBeInTheDocument();
    expect(logo).toHaveAttribute('href', '/');
    expect(logo).toHaveClass('text-2xl', 'font-bold', 'no-underline', 'text-white');
  });

  it('renders navigation links', () => {
    renderWithRouter();
    
    const homeLink = screen.getByText('Home');
    expect(homeLink).toBeInTheDocument();
    expect(homeLink).toHaveAttribute('href', '/');
    expect(homeLink).toHaveClass('hover:text-cyan-400', 'transition-colors');
  });

  it('has correct layout structure and styling', () => {
    renderWithRouter();
    
    // Check header styling
    const header = screen.getByRole('banner');
    expect(header).toHaveClass('bg-blue-900', 'text-white', 'p-4', 'shadow-md');

    // Check container styling
    const container = header.firstElementChild;
    expect(container).toHaveClass('container', 'mx-auto', 'px-4', 'flex', 'justify-between', 'items-center');

    // Check navigation styling
    const nav = screen.getByRole('navigation');
    expect(nav).toBeInTheDocument();
    expect(nav.querySelector('ul')).toHaveClass('flex', 'space-x-4');
  });
}); 