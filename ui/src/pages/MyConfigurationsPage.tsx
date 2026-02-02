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
      <div className="container mx-auto px-4 py-8">
        <LoadingSpinner />
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold text-gray-800">My Configurations</h1>
        <Link
          to="/new"
          className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
        >
          New Configuration
        </Link>
      </div>

      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-6">
          {error}
        </div>
      )}

      {configurations.length === 0 ? (
        <div className="text-center py-12 bg-gray-50 rounded-lg">
          <h2 className="text-xl font-semibold text-gray-600 mb-4">
            No configurations yet
          </h2>
          <p className="text-gray-500 mb-6">
            Create your first configuration to get started.
          </p>
          <Link
            to="/new"
            className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition-colors inline-block"
          >
            Create Configuration
          </Link>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Name
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Description
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {configurations.map((config) => (
                <tr key={config.gogen} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Link
                      to={`/configurations/${config.gogen}`}
                      className="text-blue-600 hover:text-blue-800 font-medium"
                    >
                      {getConfigDisplayName(config.gogen)}
                    </Link>
                  </td>
                  <td className="px-6 py-4">
                    <span className="text-gray-600 line-clamp-2">
                      {config.description || 'No description'}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <Link
                      to={`/configurations/${config.gogen}`}
                      className="text-gray-600 hover:text-gray-900 mr-4"
                    >
                      View
                    </Link>
                    <Link
                      to={`/edit/${config.gogen}`}
                      className="text-blue-600 hover:text-blue-800 mr-4"
                    >
                      Edit
                    </Link>
                    <button
                      onClick={() => handleDeleteClick(config.gogen)}
                      className="text-red-600 hover:text-red-800"
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
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h2 className="text-xl font-bold text-gray-800 mb-4">
              Delete Configuration
            </h2>
            <p className="text-gray-600 mb-6">
              Are you sure you want to delete{' '}
              <span className="font-semibold">
                {deleteModal.configName ? getConfigDisplayName(deleteModal.configName) : ''}
              </span>
              ? This action cannot be undone.
            </p>
            <div className="flex justify-end gap-4">
              <button
                onClick={handleDeleteCancel}
                disabled={isDeleting}
                className="px-4 py-2 text-gray-600 hover:text-gray-800 disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                onClick={handleDeleteConfirm}
                disabled={isDeleting}
                className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 disabled:opacity-50"
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
