import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import gogenApi, { ConfigurationSummary } from '../api/gogenApi';
import LoadingSpinner from '../components/LoadingSpinner';

const MyConfigurationsPage = () => {
  useAuth(); // Ensure user is authenticated
  const [configurations, setConfigurations] = useState<ConfigurationSummary[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deleteModal, setDeleteModal] = useState<{ isOpen: boolean; configName: string | null }>({
    isOpen: false,
    configName: null,
  });
  const [isDeleting, setIsDeleting] = useState(false);

  const fetchConfigurations = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const data = await gogenApi.getMyConfigurations();
      setConfigurations(data);
    } catch (err) {
      setError('Failed to load your configurations. Please try again later.');
      console.error('Error fetching configurations:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchConfigurations();
  }, []);

  const handleDeleteClick = (configName: string) => {
    setDeleteModal({ isOpen: true, configName });
  };

  const handleDeleteConfirm = async () => {
    if (!deleteModal.configName) return;

    try {
      setIsDeleting(true);
      await gogenApi.deleteConfiguration(deleteModal.configName);
      // Remove from local state
      setConfigurations(prev => prev.filter(c => c.gogen !== deleteModal.configName));
      setDeleteModal({ isOpen: false, configName: null });
    } catch (err) {
      console.error('Error deleting configuration:', err);
      setError('Failed to delete configuration. Please try again.');
    } finally {
      setIsDeleting(false);
    }
  };

  const handleDeleteCancel = () => {
    setDeleteModal({ isOpen: false, configName: null });
  };

  // Extract just the config name from full path (owner/name)
  const getConfigDisplayName = (fullName: string) => {
    const parts = fullName.split('/');
    return parts.length > 1 ? parts[1] : fullName;
  };

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-6">
        <LoadingSpinner />
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-term-text">My Configurations</h1>
        <Link
          to="/new"
          className="btn-primary"
        >
          New Configuration
        </Link>
      </div>

      {error && (
        <div className="error-box mb-6">
          {error}
        </div>
      )}

      {configurations.length === 0 ? (
        <div className="text-center py-12 bg-term-bg-elevated border border-term-border rounded">
          <h2 className="text-lg font-semibold text-term-text-muted mb-4">
            No configurations yet
          </h2>
          <p className="text-term-text-muted mb-6 text-sm">
            Create your first configuration to get started.
          </p>
          <Link
            to="/new"
            className="btn-primary inline-block"
          >
            Create Configuration
          </Link>
        </div>
      ) : (
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
                <th className="px-6 py-2 text-right text-xs font-medium text-term-text-muted uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-term-border">
              {configurations.map((config) => (
                <tr key={config.gogen} className="hover:bg-term-bg-muted transition-colors">
                  <td className="px-6 py-3 whitespace-nowrap">
                    <Link
                      to={`/configurations/${config.gogen}`}
                      className="text-term-cyan hover:text-term-green font-mono font-medium"
                    >
                      {getConfigDisplayName(config.gogen)}
                    </Link>
                  </td>
                  <td className="px-6 py-3">
                    <span className="text-term-text-muted line-clamp-2 text-sm">
                      {config.description || 'No description'}
                    </span>
                  </td>
                  <td className="px-6 py-3 whitespace-nowrap text-right text-sm font-medium">
                    <Link
                      to={`/configurations/${config.gogen}`}
                      className="text-term-text-muted hover:text-term-text mr-4"
                    >
                      View
                    </Link>
                    <Link
                      to={`/edit/${config.gogen}`}
                      className="text-term-cyan hover:text-term-green mr-4"
                    >
                      Edit
                    </Link>
                    <button
                      onClick={() => handleDeleteClick(config.gogen)}
                      className="text-term-red hover:text-red-400"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      {deleteModal.isOpen && (
        <div className="modal-backdrop">
          <div className="modal-content">
            <h2 className="text-lg font-bold text-term-text mb-4">
              Delete Configuration
            </h2>
            <p className="text-term-text-muted mb-6">
              Are you sure you want to delete{' '}
              <span className="font-semibold text-term-text font-mono">
                {deleteModal.configName ? getConfigDisplayName(deleteModal.configName) : ''}
              </span>
              ? This action cannot be undone.
            </p>
            <div className="flex justify-end gap-4">
              <button
                onClick={handleDeleteCancel}
                disabled={isDeleting}
                className="px-4 py-1.5 text-term-text-muted hover:text-term-text disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                onClick={handleDeleteConfirm}
                disabled={isDeleting}
                className="btn-danger disabled:opacity-50"
              >
                {isDeleting ? 'Deleting...' : 'Delete'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default MyConfigurationsPage;
