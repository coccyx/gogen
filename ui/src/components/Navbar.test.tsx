import { render, screen, fireEvent } from '../utils/test-utils';
import Navbar from './Navbar';

describe('Navbar', () => {
  it('renders navigation links', () => {
    render(<Navbar />);
    expect(screen.getByText('Home')).toBeInTheDocument();
    expect(screen.getByText('Configurations')).toBeInTheDocument();
  });

  it('mobile menu button toggles navigation', () => {
    render(<Navbar />);
    
    const menuButton = screen.getByLabelText('Toggle navigation menu');
    const mobileNavigation = screen.getByTestId('mobile-nav');
    
    // Menu should be hidden initially on mobile
    expect(mobileNavigation.className).toContain('hidden');
    
    // Click the menu button
    fireEvent.click(menuButton);
    
    // Menu should be visible after clicking
    expect(mobileNavigation.className).not.toContain('hidden');
    
    // Click again to hide
    fireEvent.click(menuButton);
    
    // Menu should be hidden again
    expect(mobileNavigation.className).toContain('hidden');
  });

  it('navigation links are clickable and use correct paths', () => {
    render(<Navbar />);
    
    const homeLink = screen.getByText('Home');
    const configurationsLink = screen.getByText('Configurations');
    
    expect(homeLink).toHaveAttribute('href', '/');
    expect(configurationsLink).toHaveAttribute('href', '/configurations');
  });
}); 