import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import gogenApi, { ConfigurationSummary } from '../api/gogenApi';
import LoadingSpinner from '../components/LoadingSpinner';

const ConfigurationsPage = () => {
  const [configurations, setConfigurations] = useState<ConfigurationSummary[]>([]);
  const [filteredConfigurations, setFilteredConfigurations] = useState<ConfigurationSummary[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Fetch configurations on component mount
  useEffect(() => {
    const fetchConfigurations = async () => {
      try {
        setIsLoading(true);
        const data = await gogenApi.listConfigurations();
        setConfigurations(data);
        setFilteredConfigurations(data);
        setError(null);
      } catch (err) {
        setError('Failed to load configurations. Please try again later.');
        console.error('Error fetching configurations:', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchConfigurations();
  }, []);

  // Filter configurations based on search query
  useEffect(() => {
    if (searchQuery.trim() === '') {
      setFilteredConfigurations(configurations);
    } else {
      const filtered = configurations.filter(
        (config) =>
          config.gogen.toLowerCase().includes(searchQuery.toLowerCase()) ||
          (config.description && config.description.toLowerCase().includes(searchQuery.toLowerCase()))
      );
      setFilteredConfigurations(filtered);
    }
  }, [searchQuery, configurations]);

  // Handle search input change
  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchQuery(e.target.value);
  };

  if (isLoading) {
    return <LoadingSpinner />;
  }

  if (error) {
    return (
      <div className="text-center py-8">
        <p className="text-red-600 mb-4">{error}</p>
        <button
          onClick={() => window.location.reload()}
          className="btn-primary"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div>
      <div className="flex flex-col md:flex-row justify-between items-center mb-8">
        <h1 className="text-3xl font-bold text-cribl-primary mb-4 md:mb-0">
          Gogen Configurations
        </h1>
        <div className="w-full md:w-64">
          <input
            type="text"
            placeholder="Search configurations..."
            value={searchQuery}
            onChange={handleSearchChange}
            className="w-full px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-cribl-purple"
          />
        </div>
      </div>

      {filteredConfigurations.length === 0 ? (
        <div className="text-center py-8">
          <p className="text-lg">No configurations found.</p>
          {searchQuery && (
            <p className="mt-2">
              Try adjusting your search query or{' '}
              <button
                onClick={() => setSearchQuery('')}
                className="text-cribl-purple hover:underline"
              >
                clear the search
              </button>
              .
            </p>
          )}
        </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {filteredConfigurations.map((config) => (
            <div key={config.gogen} className="card hover:shadow-lg transition-shadow">
              <h2 className="text-xl font-semibold text-cribl-primary mb-2 truncate">
                {config.gogen}
              </h2>
              <p className="text-gray-600 mb-4 h-12 overflow-hidden">
                {config.description || 'No description available'}
              </p>
              <div className="flex justify-between mt-4">
                <Link
                  to={`/configurations/${config.gogen}`}
                  className="text-blue-800 hover:text-blue-700"
                >
                  View Details
                </Link>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default ConfigurationsPage; 