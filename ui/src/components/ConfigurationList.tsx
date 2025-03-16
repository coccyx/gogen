import { Link } from 'react-router-dom';
import { ConfigurationSummary } from '../api/gogenApi';

interface ConfigurationListProps {
  configurations: ConfigurationSummary[];
  loading: boolean;
  error: string | null;
}

const ConfigurationList = ({ configurations, loading, error }: ConfigurationListProps) => {
  if (loading) {
    return (
      <div className="flex justify-center items-center h-40">
        <p className="text-lg">Loading configurations...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
        <p>{error}</p>
      </div>
    );
  }

  const getConfigPath = (gogen: string) => {
    // Split the gogen string into owner and config name
    const [owner, ...configParts] = gogen.split('/');
    return `/configurations/${owner}/${configParts.join('/')}`;
  };

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full bg-white rounded-lg overflow-hidden shadow-md">
        <thead className="bg-gray-100 text-gray-700">
          <tr>
            <th className="py-3 px-4 text-left">Name</th>
            <th className="py-3 px-4 text-left">Description</th>
            <th className="py-3 px-4 text-left">Actions</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-200">
          {configurations.map((config) => (
            <tr key={config.gogen} className="hover:bg-gray-50">
              <td className="py-3 px-4">{config.gogen}</td>
              <td className="py-3 px-4">{config.description}</td>
              <td className="py-3 px-4">
                <Link 
                  to={getConfigPath(config.gogen)}
                  className="text-blue-500 hover:underline"
                >
                  View Details
                </Link>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default ConfigurationList; 