import { render, screen } from '../utils/test-utils';
import Hero from './Hero';

describe('Hero', () => {
  it('renders hero section with correct content', () => {
    render(<Hero />);
    
    // Check for main heading
    expect(screen.getByRole('heading', { 
      level: 2, 
      name: /gogen helps generate telemetry data, quickly and easily\./i 
    })).toBeInTheDocument();

    // Check for description text
    expect(screen.getByText(/view and manage your gogen configurations\./i)).toBeInTheDocument();
  });

  it('has correct accessibility attributes', () => {
    render(<Hero />);
    
    // Check for correct ARIA role and label
    const heroSection = screen.getByRole('region', { name: /hero/i });
    expect(heroSection).toBeInTheDocument();
    expect(heroSection).toHaveAttribute('aria-label', 'hero');
  });

  it('has correct styling classes', () => {
    render(<Hero />);
    
    // Check section styling
    const heroSection = screen.getByRole('region', { name: /hero/i });
    expect(heroSection).toHaveClass('bg-blue-800', 'text-white', 'py-12');

    // Check container styling
    const container = heroSection.firstElementChild;
    expect(container).toHaveClass('container', 'mx-auto', 'px-4');

    // Check content container styling
    const contentContainer = container?.firstElementChild;
    expect(contentContainer).toHaveClass('max-w-3xl');

    // Check heading styling
    const heading = screen.getByRole('heading', { level: 2 });
    expect(heading).toHaveClass('text-4xl', 'font-bold', 'mb-4');

    // Check paragraph styling
    const paragraph = screen.getByText(/view and manage/i);
    expect(paragraph).toHaveClass('text-xl');
  });
}); 