import { render, screen } from '../utils/test-utils';
import Footer from './Footer';

describe('Footer', () => {
  beforeEach(() => {
    // Mock the current year to make tests deterministic
    jest.useFakeTimers();
    jest.setSystemTime(new Date('2024-01-01'));
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('renders copyright text with current year', () => {
    render(<Footer />);
    const currentYear = new Date().getFullYear();
    expect(
      screen.getByText((content) => {
        return content.includes(currentYear.toString()) && 
               content.includes('Clint Sharp') && 
               content.includes('All rights reserved');
      })
    ).toBeInTheDocument();
  });

  it('renders GitHub link with correct attributes', () => {
    render(<Footer />);
    const githubLink = screen.getByLabelText('GitHub');
    
    expect(githubLink).toHaveAttribute('href', 'https://github.com/coccyx/gogen');
    expect(githubLink).toHaveAttribute('target', '_blank');
    expect(githubLink).toHaveAttribute('rel', 'noopener noreferrer');
  });

  it('renders GitHub icon', () => {
    render(<Footer />);
    // Find SVG by its role and verify it's within the GitHub link
    const githubLink = screen.getByLabelText('GitHub');
    const svg = githubLink.querySelector('svg');
    expect(svg).toBeInTheDocument();
    expect(svg).toHaveAttribute('aria-hidden', 'true');
  });

  it('has correct responsive layout classes', () => {
    render(<Footer />);
    const footer = screen.getByRole('contentinfo');
    const container = footer.firstElementChild;
    const flexContainer = container?.firstElementChild;

    expect(footer).toHaveClass('bg-cribl-primary', 'text-white', 'py-6');
    expect(container).toHaveClass('container-custom', 'mx-auto', 'px-4');
    expect(flexContainer).toHaveClass('flex', 'flex-col', 'md:flex-row', 'justify-between', 'items-center');
  });
}); 