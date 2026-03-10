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
  if (error) return <div className="text-term-red" role="alert">{error}</div>;

  return (
    <div>
      {/* Search Filter */}
      <div className="mb-4">
        <input
          type="text"
          placeholder="Search configurations..."
          value={searchQuery}
          onChange={(e) => {
            setSearchQuery(e.target.value);
            setCurrentPage(1); // Reset to first page when searching
          }}
          className="input"
        />
      </div>

      {/* Results count */}
      <div className="text-sm text-term-text-muted mb-4">
        Showing {Math.min(filteredAndSortedConfigs.length, itemsPerPage)} of {filteredAndSortedConfigs.length} configurations
      </div>

      {/* Configurations Table */}
      <div className="bg-term-bg-elevated rounded border border-term-border overflow-hidden">
        <table className="min-w-full">
          <thead className="bg-term-bg-muted">
            <tr>
              <th className="px-6 py-2 text-left text-xs font-medium text-term-text-muted uppercase tracking-wider">
                Name
              </th>
              <th className="px-6 py-2 text-left text-xs font-medium text-term-text-muted uppercase tracking-wider">
                Description
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-term-border">
            {paginatedConfigs.map((config) => (
              <tr key={config.gogen} className="hover:bg-term-bg-muted transition-colors">
                <td className="px-6 py-3 whitespace-nowrap">
                  <Link
                    to={`/configurations/${config.gogen}`}
                    className="text-term-cyan hover:text-term-green font-mono font-medium"
                  >
                    {config.gogen}
                  </Link>
                </td>
                <td className="px-6 py-3">
                  <div className="text-sm text-term-text-muted">{config.description || '-'}</div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="mt-4 flex justify-center">
          <nav className="relative z-0 inline-flex rounded -space-x-px" aria-label="Pagination">
            <button
              onClick={() => handlePageChange(currentPage - 1)}
              disabled={currentPage === 1}
              className={`relative inline-flex items-center px-2 py-1.5 rounded-l border border-term-border bg-term-bg-elevated text-sm font-medium ${
                currentPage === 1
                  ? 'text-term-text-muted cursor-not-allowed'
                  : 'text-term-text hover:bg-term-bg-muted'
              }`}
            >
              Previous
            </button>
            {[...Array(totalPages)].map((_, index) => (
              <button
                key={index + 1}
                onClick={() => handlePageChange(index + 1)}
                className={`relative inline-flex items-center px-3 py-1.5 border ${
                  currentPage === index + 1
                    ? 'z-10 bg-term-green border-term-green text-term-bg'
                    : 'bg-term-bg-elevated border-term-border text-term-text hover:bg-term-bg-muted'
                }`}
              >
                {index + 1}
              </button>
            ))}
            <button
              onClick={() => handlePageChange(currentPage + 1)}
              disabled={currentPage === totalPages}
              className={`relative inline-flex items-center px-2 py-1.5 rounded-r border border-term-border bg-term-bg-elevated text-sm font-medium ${
                currentPage === totalPages
                  ? 'text-term-text-muted cursor-not-allowed'
                  : 'text-term-text hover:bg-term-bg-muted'
              }`}
            >
              Next
            </button>
          </nav>
        </div>
      )}

      {/* No results message */}
      {filteredAndSortedConfigs.length === 0 && (
        <div className="text-center py-8 text-term-text-muted">
          No configurations found matching your search.
        </div>
      )}
    </div>
  );
};

export default ConfigurationList;
