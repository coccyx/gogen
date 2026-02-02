import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import gogenApi, { Configuration } from '../api/gogenApi';
import LoadingSpinner from '../components/LoadingSpinner';
import ExecutionComponent from '../components/ExecutionComponent';
import Editor from '@monaco-editor/react';

const ConfigurationDetailPage = () => {
  const { owner, configName } = useParams<{ owner: string; configName: string }>();
  const navigate = useNavigate();
  const { user, isAuthenticated } = useAuth();
  const [configuration, setConfiguration] = useState<Configuration | null>(null);
  const [editedConfig, setEditedConfig] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deleteModal, setDeleteModal] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  const isOwner = isAuthenticated && user?.login === owner;

  useEffect(() => {
    const fetchConfiguration = async () => {
      if (!owner || !configName) {
        setError('Configuration details are incomplete');
        setIsLoading(false);
        return;
      }

      try {
        setIsLoading(true);
        const fullConfigName = `${owner}/${configName}`;
        const data = await gogenApi.getConfiguration(fullConfigName);
        setConfiguration(data);
        setEditedConfig(data.config || '');
        setError(null);
      } catch (err) {
        setError('Failed to load configuration. Please try again later.');
        console.error(`Error fetching configuration ${owner}/${configName}:`, err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchConfiguration();
  }, [owner, configName]);

  // Create a modified configuration object with the edited config
  const getExecutionConfiguration = (): Configuration => {
    if (!configuration) throw new Error('Configuration not loaded');
    return {
      ...configuration,
      config: editedConfig
    };
  };

  const handleDelete = async () => {
    if (!owner || !configName) return;

    try {
      setIsDeleting(true);
      await gogenApi.deleteConfiguration(`${owner}/${configName}`);
      navigate('/my-configurations');
    } catch (err) {
      console.error('Error deleting configuration:', err);
      setError('Failed to delete configuration. Please try again.');
    } finally {
      setIsDeleting(false);
      setDeleteModal(false);
    }
  };

  const handleFork = () => {
    // Navigate to new config page with pre-filled config
    // Store the config in sessionStorage to pass to the new page
    if (configuration) {
      sessionStorage.setItem('fork_config', JSON.stringify({
        config: editedConfig,
        description: configuration.description ? `Forked from ${owner}/${configName}: ${configuration.description}` : `Forked from ${owner}/${configName}`,
      }));
      navigate('/new');
    }
  };

  const renderContent = () => {
    if (isLoading) {
      return <LoadingSpinner />;
    }

    if (error || !configuration) {
      return (
        <div className="text-center py-8">
          <p className="text-red-600 mb-4">{error || 'Configuration not found'}</p>
          <button
            onClick={() => navigate('/')}
            className="btn-primary"
          >
            Back to Home
          </button>
        </div>
      );
    }

    return (
      <>
        <div className="flex justify-between items-center mb-6">
          <div>
            <h1 className="text-3xl font-bold text-gray-800">{configuration.gogen}</h1>
            <p className="text-gray-500 text-sm mt-1">by {owner}</p>
          </div>
          <div className="flex items-center gap-3">
            {isOwner ? (
              <>
                <Link
                  to={`/edit/${owner}/${configName}`}
                  className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
                >
                  Edit
                </Link>
                <button
                  onClick={() => setDeleteModal(true)}
                  className="bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition-colors"
                >
                  Delete
                </button>
              </>
            ) : isAuthenticated ? (
              <button
                onClick={handleFork}
                className="bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 transition-colors"
              >
                Fork
              </button>
            ) : null}
            <Link to="/" className="btn-primary">
              Back to List
            </Link>
          </div>
        </div>

        {configuration.description && (
          <div className="mb-8">
            <h2 className="text-xl font-semibold mb-2">Description</h2>
            <p className="text-gray-700">{configuration.description}</p>
          </div>
        )}

        <div className="mb-8">
          <h2 className="text-xl font-semibold mb-2">Configuration</h2>
          <div className="border rounded-lg overflow-hidden shadow-md">
            <Editor
              height="400px"
              defaultLanguage="yaml"
              value={editedConfig}
              onChange={(value: string | undefined) => setEditedConfig(value || '')}
              theme="vs-light"
              options={{
                minimap: { enabled: false },
                scrollBeyondLastLine: false,
                fontSize: 14,
                lineNumbers: 'on',
                renderLineHighlight: 'all',
                automaticLayout: true,
              }}
            />
          </div>
        </div>

        {configuration.samples && configuration.samples.length > 0 && (
          <div className="mb-8">
            <h2 className="text-xl font-semibold mb-2">Samples</h2>
            <div className="bg-gray-100 p-4 rounded-md overflow-x-auto">
              <pre>
                <code>{JSON.stringify(configuration.samples, null, 2)}</code>
              </pre>
            </div>
          </div>
        )}

        {configuration.raters && configuration.raters.length > 0 && (
          <div className="mb-8">
            <h2 className="text-xl font-semibold mb-2">Raters</h2>
            <div className="bg-gray-100 p-4 rounded-md overflow-x-auto">
              <pre>
                <code>{JSON.stringify(configuration.raters, null, 2)}</code>
              </pre>
            </div>
          </div>
        )}

        <ExecutionComponent configuration={getExecutionConfiguration()} />
      </>
    );
  };

  return (
    <div className="container mx-auto px-4 py-8" role="main">
      {renderContent()}

      {/* Delete Confirmation Modal */}
      {deleteModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h2 className="text-xl font-bold text-gray-800 mb-4">
              Delete Configuration
            </h2>
            <p className="text-gray-600 mb-6">
              Are you sure you want to delete{' '}
              <span className="font-semibold">{configName}</span>? This action cannot be undone.
            </p>
            <div className="flex justify-end gap-4">
              <button
                onClick={() => setDeleteModal(false)}
                disabled={isDeleting}
                className="px-4 py-2 text-gray-600 hover:text-gray-800 disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                onClick={handleDelete}
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

export default ConfigurationDetailPage;
