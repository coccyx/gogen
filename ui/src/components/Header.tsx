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
    <header className="bg-blue-900 text-white p-4 shadow-md">
      <div className="container mx-auto px-4 flex justify-between items-center">
        <Link to="/" className="text-2xl font-bold no-underline text-white">Gogen UI</Link>
        <nav>
          <ul className="flex items-center space-x-4">
            <li>
              <Link to="/" className="hover:text-cyan-400 transition-colors">Home</Link>
            </li>
            {!isLoading && (
              <>
                {isAuthenticated && user ? (
                  <>
                    <li>
                      <Link
                        to="/my-configurations"
                        className="hover:text-cyan-400 transition-colors"
                      >
                        My Configs
                      </Link>
                    </li>
                    <li>
                      <Link
                        to="/new"
                        className="bg-cyan-600 hover:bg-cyan-700 px-3 py-1 rounded transition-colors"
                      >
                        New Config
                      </Link>
                    </li>
                    <li className="relative" ref={dropdownRef}>
                      <button
                        onClick={() => setIsDropdownOpen(!isDropdownOpen)}
                        className="flex items-center space-x-2 hover:text-cyan-400 transition-colors"
                      >
                        <img
                          src={user.avatar_url}
                          alt={user.login}
                          className="w-8 h-8 rounded-full border-2 border-white"
                        />
                        <span className="hidden sm:inline">{user.login}</span>
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
                        <div className="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg py-1 z-50">
                          <Link
                            to="/my-configurations"
                            className="block px-4 py-2 text-gray-700 hover:bg-gray-100"
                            onClick={() => setIsDropdownOpen(false)}
                          >
                            My Configurations
                          </Link>
                          <Link
                            to="/new"
                            className="block px-4 py-2 text-gray-700 hover:bg-gray-100"
                            onClick={() => setIsDropdownOpen(false)}
                          >
                            New Configuration
                          </Link>
                          <hr className="my-1" />
                          <button
                            onClick={() => {
                              setIsDropdownOpen(false);
                              logout();
                            }}
                            className="block w-full text-left px-4 py-2 text-gray-700 hover:bg-gray-100"
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
                      className="bg-cyan-600 hover:bg-cyan-700 px-4 py-2 rounded transition-colors"
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
