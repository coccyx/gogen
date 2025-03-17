import { useState, useMemo } from 'react';
import { Link } from 'react-router-dom';
import { ConfigurationSummary } from '../api/gogenApi';
import LoadingSpinner from './LoadingSpinner';

interface ConfigurationListProps {
  configurations: ConfigurationSummary[];
  loading: boolean;
  error: string | null;
}

const ConfigurationList = ({ configurations, loading, error }: ConfigurationListProps) => {
  const [currentPage, setCurrentPage] = useState(1);
  const [searchQuery, setSearchQuery] = useState('');
  const itemsPerPage = 10;

  // Filter and sort configurations
  const filteredAndSortedConfigs = useMemo(() => {
    return configurations
      .filter(config => 
        config.gogen.toLowerCase().includes(searchQuery.toLowerCase()) ||
        (config.description || '').toLowerCase().includes(searchQuery.toLowerCase())
      )
      .sort((a, b) => a.gogen.localeCompare(b.gogen));
  }, [configurations, searchQuery]);

  // Calculate pagination
  const totalPages = Math.ceil(filteredAndSortedConfigs.length / itemsPerPage);
  const startIndex = (currentPage - 1) * itemsPerPage;
  const paginatedConfigs = filteredAndSortedConfigs.slice(startIndex, startIndex + itemsPerPage);

  // Handle page changes
  const handlePageChange = (page: number) => {
    setCurrentPage(page);
    window.scrollTo(0, 0);
  };

  if (loading) return <LoadingSpinner />;
  if (error) return <div className="text-red-600" role="alert">{error}</div>;

  return (
    <div>
      {/* Search Filter */}
      <div className="mb-6">
        <input
          type="text"
          placeholder="Search configurations..."
          value={searchQuery}
          onChange={(e) => {
            setSearchQuery(e.target.value);
            setCurrentPage(1); // Reset to first page when searching
          }}
          className="w-full px-4 py-2 rounded-md border border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm bg-white text-gray-900"
        />
      </div>

      {/* Results count */}
      <div className="text-sm text-gray-600 mb-4">
        Showing {Math.min(filteredAndSortedConfigs.length, itemsPerPage)} of {filteredAndSortedConfigs.length} configurations
      </div>

      {/* Configurations Table */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="min-w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Name
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Description
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {paginatedConfigs.map((config) => (
              <tr key={config.gogen} className="hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap">
                  <Link
                    to={`/configurations/${config.gogen}`}
                    className="text-blue-800 hover:text-blue-600 font-medium"
                  >
                    {config.gogen}
                  </Link>
                </td>
                <td className="px-6 py-4">
                  <div className="text-sm text-gray-900">{config.description || '-'}</div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="mt-6 flex justify-center">
          <nav className="relative z-0 inline-flex rounded-md shadow-sm -space-x-px" aria-label="Pagination">
            <button
              onClick={() => handlePageChange(currentPage - 1)}
              disabled={currentPage === 1}
              className={`relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium ${
                currentPage === 1
                  ? 'text-gray-300 cursor-not-allowed'
                  : 'text-gray-500 hover:bg-gray-50'
              }`}
            >
              Previous
            </button>
            {[...Array(totalPages)].map((_, index) => (
              <button
                key={index + 1}
                onClick={() => handlePageChange(index + 1)}
                className={`relative inline-flex items-center px-4 py-2 border ${
                  currentPage === index + 1
                    ? 'z-10 bg-blue-800 border-blue-800 text-white'
                    : 'bg-white border-gray-300 text-gray-500 hover:bg-gray-50'
                }`}
              >
                {index + 1}
              </button>
            ))}
            <button
              onClick={() => handlePageChange(currentPage + 1)}
              disabled={currentPage === totalPages}
              className={`relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium ${
                currentPage === totalPages
                  ? 'text-gray-300 cursor-not-allowed'
                  : 'text-gray-500 hover:bg-gray-50'
              }`}
            >
              Next
            </button>
          </nav>
        </div>
      )}

      {/* No results message */}
      {filteredAndSortedConfigs.length === 0 && (
        <div className="text-center py-8 text-gray-500">
          No configurations found matching your search.
        </div>
      )}
    </div>
  );
};

export default ConfigurationList; 