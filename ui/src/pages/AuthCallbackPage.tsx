import { useEffect, useState, useRef } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const AuthCallbackPage = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { handleCallback, isAuthenticated } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const processedRef = useRef(false);

  useEffect(() => {
    // Prevent double execution in React StrictMode
    if (processedRef.current) return;

    // If already authenticated, redirect to home
    if (isAuthenticated) {
      navigate('/', { replace: true });
      return;
    }

    const code = searchParams.get('code');
    const state = searchParams.get('state');
    const errorParam = searchParams.get('error');
    const errorDescription = searchParams.get('error_description');

    // Check for OAuth errors from GitHub
    if (errorParam) {
      setError(errorDescription || errorParam);
      return;
    }

    // Validate required parameters
    if (!code || !state) {
      setError('Missing required parameters. Please try logging in again.');
      return;
    }

    // Mark as processed before async operation
    processedRef.current = true;

    // Exchange the code for an access token
    handleCallback(code, state)
      .then(() => {
        // Redirect to home on success
        navigate('/', { replace: true });
      })
      .catch((err) => {
        console.error('OAuth callback error:', err);
        setError(err.message || 'Authentication failed. Please try again.');
        // Reset so user can try again
        processedRef.current = false;
      });
  }, [searchParams, handleCallback, navigate, isAuthenticated]);

  if (error) {
    return (
      <div className="container mx-auto px-4 py-16">
        <div className="max-w-md mx-auto bg-white rounded-lg shadow-md p-8">
          <h1 className="text-2xl font-bold text-red-600 mb-4 text-center">
            Authentication Error
          </h1>
          <p className="text-gray-600 mb-6 text-center">{error}</p>
          <button
            onClick={() => navigate('/login', { replace: true })}
            className="w-full bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 transition-colors"
          >
            Back to Login
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-16">
      <div className="max-w-md mx-auto bg-white rounded-lg shadow-md p-8 text-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-900 mx-auto mb-6"></div>
        <h1 className="text-xl font-semibold text-gray-800 mb-2">
          Completing sign in...
        </h1>
        <p className="text-gray-600">
          Please wait while we complete your authentication.
        </p>
      </div>
    </div>
  );
};

export default AuthCallbackPage;
