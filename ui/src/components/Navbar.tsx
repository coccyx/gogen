import { Link, NavLink } from 'react-router-dom';
import { useState } from 'react';

const Navbar = () => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  const toggleMenu = () => {
    setIsMenuOpen(!isMenuOpen);
  };

  return (
    <nav className="bg-cribl-primary text-white shadow-md">
      <div className="container-custom mx-auto px-4">
        <div className="flex justify-between items-center py-4">
          <Link to="/" className="flex items-center space-x-3">
            <span className="text-2xl font-bold">Gogen UI</span>
          </Link>

          {/* Mobile menu button */}
          <div className="md:hidden">
            <button
              onClick={toggleMenu}
              className="text-white hover:text-cribl-pink focus:outline-none"
              aria-label="Toggle menu"
            >
              <svg
                className="h-6 w-6"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                {isMenuOpen ? (
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                ) : (
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M4 6h16M4 12h16M4 18h16"
                  />
                )}
              </svg>
            </button>
          </div>

          {/* Desktop menu */}
          <div className="hidden md:flex md:items-center md:space-x-6">
            <NavLink
              to="/"
              className={({ isActive }) =>
                isActive
                  ? "text-cribl-pink font-medium"
                  : "text-white hover:text-cribl-pink transition-colors"
              }
              end
            >
              Home
            </NavLink>
            <NavLink
              to="/configurations"
              className={({ isActive }) =>
                isActive
                  ? "text-cribl-pink font-medium"
                  : "text-white hover:text-cribl-pink transition-colors"
              }
            >
              Configurations
            </NavLink>
          </div>
        </div>

        {/* Mobile menu */}
        {isMenuOpen && (
          <div className="md:hidden py-4 space-y-4">
            <NavLink
              to="/"
              className={({ isActive }) =>
                isActive
                  ? "block text-cribl-pink font-medium"
                  : "block text-white hover:text-cribl-pink transition-colors"
              }
              end
              onClick={() => setIsMenuOpen(false)}
            >
              Home
            </NavLink>
            <NavLink
              to="/configurations"
              className={({ isActive }) =>
                isActive
                  ? "block text-cribl-pink font-medium"
                  : "block text-white hover:text-cribl-pink transition-colors"
              }
              onClick={() => setIsMenuOpen(false)}
            >
              Configurations
            </NavLink>
          </div>
        )}
      </div>
    </nav>
  );
};

export default Navbar; 