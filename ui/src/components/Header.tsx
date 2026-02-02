import { useState, useRef, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const Header = () => {
  const { user, isAuthenticated, logout, isLoading } = useAuth();
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLLIElement>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsDropdownOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <header className="bg-term-bg-elevated text-term-text px-4 py-2 border-b border-term-border">
      <div className="container mx-auto px-4 flex justify-between items-center">
        <Link to="/" className="font-mono text-xl font-semibold no-underline text-term-text hover:text-term-green transition-colors">
          gogen
        </Link>
        <nav>
          <ul className="flex items-center space-x-4">
            <li>
              <Link to="/" className="text-sm hover:text-term-green transition-colors">Home</Link>
            </li>
            {!isLoading && (
              <>
                {isAuthenticated && user ? (
                  <>
                    <li>
                      <Link
                        to="/my-configurations"
                        className="text-sm hover:text-term-green transition-colors"
                      >
                        My Configs
                      </Link>
                    </li>
                    <li>
                      <Link
                        to="/new"
                        className="bg-term-green hover:bg-opacity-90 text-term-bg px-3 py-1 rounded text-sm font-medium transition-colors"
                      >
                        New Config
                      </Link>
                    </li>
                    <li className="relative" ref={dropdownRef}>
                      <button
                        onClick={() => setIsDropdownOpen(!isDropdownOpen)}
                        className="flex items-center space-x-2 hover:text-term-green transition-colors"
                      >
                        <img
                          src={user.avatar_url}
                          alt={user.login}
                          className="w-7 h-7 rounded-full border border-term-border"
                        />
                        <span className="hidden sm:inline text-sm">{user.login}</span>
                        <svg
                          className={`w-4 h-4 transition-transform ${isDropdownOpen ? 'rotate-180' : ''}`}
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                        </svg>
                      </button>
                      {isDropdownOpen && (
                        <div className="absolute right-0 mt-2 w-48 bg-term-bg-elevated border border-term-border rounded shadow-lg py-1 z-50">
                          <Link
                            to="/my-configurations"
                            className="block px-4 py-2 text-sm text-term-text hover:bg-term-bg-muted"
                            onClick={() => setIsDropdownOpen(false)}
                          >
                            My Configurations
                          </Link>
                          <Link
                            to="/new"
                            className="block px-4 py-2 text-sm text-term-text hover:bg-term-bg-muted"
                            onClick={() => setIsDropdownOpen(false)}
                          >
                            New Configuration
                          </Link>
                          <hr className="my-1 border-term-border" />
                          <button
                            onClick={() => {
                              setIsDropdownOpen(false);
                              logout();
                            }}
                            className="block w-full text-left px-4 py-2 text-sm text-term-text hover:bg-term-bg-muted"
                          >
                            Sign out
                          </button>
                        </div>
                      )}
                    </li>
                  </>
                ) : (
                  <li>
                    <Link
                      to="/login"
                      className="bg-term-green hover:bg-opacity-90 text-term-bg px-3 py-1.5 rounded text-sm font-medium transition-colors"
                    >
                      Login
                    </Link>
                  </li>
                )}
              </>
            )}
          </ul>
        </nav>
      </div>
    </header>
  );
};

export default Header;
