import { Link } from 'react-router-dom';

const NotFoundPage = () => {
  return (
    <div className="flex flex-col items-center justify-center py-16" data-testid="not-found-container">
      <h1 className="text-6xl font-bold text-cribl-primary mb-4">404</h1>
      <h2 className="text-2xl font-semibold text-cribl-primary mb-6">Page Not Found</h2>
      <p className="text-lg mb-8 text-center max-w-md">
        The page you are looking for might have been removed, had its name changed, or is temporarily unavailable.
      </p>
      <Link to="/" className="btn-primary">
        Go to Homepage
      </Link>
    </div>
  );
};

export default NotFoundPage; 