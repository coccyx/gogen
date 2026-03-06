import { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import gogenApi from '../api/gogenApi';
import LoadingSpinner from '../components/LoadingSpinner';
import ExecutionComponent from '../components/ExecutionComponent';
import Editor from '@monaco-editor/react';

const DEFAULT_CONFIG = `# Gogen Configuration
# See https://github.com/coccyx/gogen for documentation

global:
  output:
    outputter: stdout

samples:
  - name: sample
    interval: 1
    count: 1
    tokens:
      - name: ts
        type: timestamp
        format: "2006-01-02T15:04:05.000Z"
    lines:
      - _raw: "ts={{ .ts }} message=Hello World"
`;

const EditConfigurationPage = () => {
  const { owner, configName } = useParams<{ owner: string; configName: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();

  const isEditMode = !!(owner && configName);

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [config, setConfig] = useState(DEFAULT_CONFIG);
  const [isLoading, setIsLoading] = useState(isEditMode);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isEditMode && owner && configName) {
      fetchConfiguration();
    } else {
      // Check for forked config in sessionStorage
      const forkData = sessionStorage.getItem('fork_config');
      if (forkData) {
        try {
          const { config: forkedConfig, description: forkedDescription } = JSON.parse(forkData);
          setConfig(forkedConfig || DEFAULT_CONFIG);
          setDescription(forkedDescription || '');
          sessionStorage.removeItem('fork_config');
        } catch (e) {
          console.error('Error parsing fork data:', e);
        }
      }
    }
  }, [owner, configName, isEditMode]);

  const fetchConfiguration = async () => {
    if (!owner || !configName) return;

    try {
      setIsLoading(true);
      setError(null);
      const fullConfigName = `${owner}/${configName}`;
      const data = await gogenApi.getConfiguration(fullConfigName);

      // Verify ownership
      if (data.owner !== user?.login) {
        setError('You can only edit your own configurations.');
        return;
      }

      setName(configName);
      setDescription(data.description || '');
      setConfig(data.config || DEFAULT_CONFIG);
    } catch (err) {
      setError('Failed to load configuration. Please try again.');
      console.error('Error fetching configuration:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const validateForm = (): string | null => {
    if (!name.trim()) {
      return 'Configuration name is required.';
    }
    if (!/^[a-zA-Z0-9_-]+$/.test(name)) {
      return 'Configuration name can only contain letters, numbers, hyphens, and underscores.';
    }
    if (!config.trim()) {
      return 'Configuration content is required.';
    }
    return null;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const validationError = validateForm();
    if (validationError) {
      setError(validationError);
      return;
    }

    if (!user) {
      setError('You must be logged in to save configurations.');
      return;
    }

    try {
      setIsSaving(true);
      setError(null);

      await gogenApi.upsertConfiguration({
        name: name.trim(),
        description: description.trim(),
        config: config,
      });

      // Redirect to the configuration page
      navigate(`/configurations/${user.login}/${name.trim()}`);
    } catch (err: any) {
      const errorMessage = err?.response?.data?.error || 'Failed to save configuration. Please try again.';
      setError(errorMessage);
      console.error('Error saving configuration:', err);
    } finally {
      setIsSaving(false);
    }
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
      <div className="max-w-4xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-bold text-term-text">
            {isEditMode ? 'Edit Configuration' : 'New Configuration'}
          </h1>
          <Link
            to={isEditMode ? `/configurations/${owner}/${configName}` : '/my-configurations'}
            className="text-term-text-muted hover:text-term-text"
          >
            Cancel
          </Link>
        </div>

        {error && (
          <div className="error-box mb-6">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-term-text-muted mb-1">
              Name
            </label>
            <input
              type="text"
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              disabled={isEditMode}
              placeholder="my-configuration"
              className="input disabled:bg-term-bg-muted disabled:cursor-not-allowed"
            />
            {isEditMode && (
              <p className="mt-1 text-sm text-term-text-muted">
                Configuration name cannot be changed after creation.
              </p>
            )}
          </div>

          <div>
            <label htmlFor="description" className="block text-sm font-medium text-term-text-muted mb-1">
              Description
            </label>
            <textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="A brief description of your configuration..."
              rows={3}
              className="input resize-none"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-term-text-muted mb-1">
              Configuration (YAML)
            </label>
            <div className="border border-term-border rounded overflow-hidden">
              <Editor
                height="400px"
                defaultLanguage="yaml"
                value={config}
                onChange={(value) => setConfig(value || '')}
                theme="vs-dark"
                options={{
                  minimap: { enabled: false },
                  scrollBeyondLastLine: false,
                  fontSize: 13,
                  fontFamily: 'JetBrains Mono, ui-monospace, SFMono-Regular, monospace',
                  lineNumbers: 'on',
                  renderLineHighlight: 'all',
                  automaticLayout: true,
                  tabSize: 2,
                }}
              />
            </div>
          </div>

          <div className="flex justify-end gap-4 pt-4">
            <Link
              to={isEditMode ? `/configurations/${owner}/${configName}` : '/my-configurations'}
              className="px-4 py-1.5 text-term-text-muted hover:text-term-text"
            >
              Cancel
            </Link>
            <button
              type="submit"
              disabled={isSaving}
              className="btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isSaving ? 'Saving...' : isEditMode ? 'Save Changes' : 'Create Configuration'}
            </button>
          </div>
        </form>

        <ExecutionComponent
          configuration={{
            gogen: name || 'preview',
            description: description,
            config: config,
          }}
        />
      </div>
    </div>
  );
};

export default EditConfigurationPage;
