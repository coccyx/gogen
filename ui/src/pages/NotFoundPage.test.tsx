import { render, screen } from '../utils/test-utils';
import { MemoryRouter } from 'react-router-dom';
import NotFoundPage from './NotFoundPage';

const renderWithRouter = () => {
  return render(
    <MemoryRouter>
      <NotFoundPage />
    </MemoryRouter>
  );
};

describe('NotFoundPage', () => {
  it('renders 404 message', () => {
    renderWithRouter();
    expect(screen.getByText('404')).toBeInTheDocument();
    expect(screen.getByText('Page Not Found')).toBeInTheDocument();
  });

  it('renders link to home page', () => {
    renderWithRouter();
    const homeLink = screen.getByRole('link', { name: 'Go to Homepage' });
    expect(homeLink).toBeInTheDocument();
    expect(homeLink).toHaveAttribute('href', '/');
  });

  it('has correct styling classes', () => {
    renderWithRouter();
    
    // Check container styling
    const container = screen.getByTestId('not-found-container');
    expect(container).toHaveClass('flex', 'flex-col', 'items-center', 'justify-center', 'py-16');
    
    // Check heading styling
    const heading = screen.getByRole('heading', { level: 1 });
    expect(heading).toHaveClass('text-6xl', 'font-bold', 'text-cribl-primary', 'mb-4');
    
    // Check link styling
    const link = screen.getByRole('link', { name: 'Go to Homepage' });
    expect(link).toHaveClass('btn-primary');
  });
}); 