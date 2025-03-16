import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import gogenApi, { Configuration } from '../api/gogenApi';
import LoadingSpinner from '../components/LoadingSpinner';
import ExecutionComponent from '../components/ExecutionComponent';
import Editor from '@monaco-editor/react';

const ConfigurationDetailPage = () => {
  const { owner, configName } = useParams<{ owner: string; configName: string }>();
  const navigate = useNavigate();
  const [configuration, setConfiguration] = useState<Configuration | null>(null);
  const [editedConfig, setEditedConfig] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

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
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-gray-800">{configuration.gogen}</h1>
        <Link to="/" className="btn-primary">
          Back to List
        </Link>
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
    </div>
  );
};

export default ConfigurationDetailPage; 