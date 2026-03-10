import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { config } from '../config';
import { gogenApi } from '../api/gogenApi';

interface User {
  login: string;
  avatar_url: string;
  id: number;
  name?: string;
  email?: string;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: () => void;
  logout: () => void;
  handleCallback: (code: string, state: string) => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const STORAGE_KEY_TOKEN = 'github_token';
const STORAGE_KEY_USER = 'github_user';
const SESSION_KEY_STATE = 'oauth_state';

function generateRandomState(): string {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Load saved auth state on mount
  useEffect(() => {
    const savedToken = localStorage.getItem(STORAGE_KEY_TOKEN);
    const savedUser = localStorage.getItem(STORAGE_KEY_USER);

    if (savedToken && savedUser) {
      try {
        setToken(savedToken);
        setUser(JSON.parse(savedUser));
      } catch (e) {
        // Clear invalid data
        localStorage.removeItem(STORAGE_KEY_TOKEN);
        localStorage.removeItem(STORAGE_KEY_USER);
      }
    }
    setIsLoading(false);
  }, []);

  const login = () => {
    // Generate and store state for CSRF protection
    const state = generateRandomState();
    sessionStorage.setItem(SESSION_KEY_STATE, state);

    // Build GitHub OAuth URL
    const params = new URLSearchParams({
      client_id: config.githubClientId,
      redirect_uri: config.githubRedirectUri,
      state: state,
      scope: 'read:user',
    });

    // Redirect to GitHub
    window.location.href = `https://github.com/login/oauth/authorize?${params.toString()}`;
  };

  const logout = () => {
    // Clear localStorage
    localStorage.removeItem(STORAGE_KEY_TOKEN);
    localStorage.removeItem(STORAGE_KEY_USER);

    // Clear state
    setToken(null);
    setUser(null);
  };

  const handleCallback = async (code: string, state: string) => {
    // Verify state to prevent CSRF attacks
    const savedState = sessionStorage.getItem(SESSION_KEY_STATE);
    sessionStorage.removeItem(SESSION_KEY_STATE);

    if (!savedState || savedState !== state) {
      throw new Error('Invalid state parameter. Please try logging in again.');
    }

    // Exchange code for token
    const response = await gogenApi.exchangeOAuthCode(code, state);

    // Save to localStorage
    localStorage.setItem(STORAGE_KEY_TOKEN, response.access_token);
    localStorage.setItem(STORAGE_KEY_USER, JSON.stringify(response.user));

    // Update state
    setToken(response.access_token);
    setUser(response.user);
  };

  const value: AuthContextType = {
    user,
    token,
    isAuthenticated: !!token && !!user,
    isLoading,
    login,
    logout,
    handleCallback,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
