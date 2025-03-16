import axios from 'axios';

// Add Vite env type definition
interface ImportMetaEnv {
  VITE_API_URL?: string;
  // Add other environment variables as needed
}

// Extend ImportMeta
interface ImportMeta {
  readonly env: ImportMetaEnv;
}

// Define the base URL for the API
// Use a direct approach to avoid process.env issues
const API_BASE_URL = '/api';  // This will be proxied to the configured API URL

// Create an axios instance with default config
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Define interfaces for API responses
export interface ConfigurationSummary {
  gogen: string;
  description: string;
}

export interface Configuration extends ConfigurationSummary {
  config?: string;
  samples?: any[];
  raters?: any[];
  mix?: any[];
  generators?: any[];
  global?: any;
  templates?: any[];
}

// API functions
export const gogenApi = {
  // Get a list of all configurations
  listConfigurations: async (): Promise<ConfigurationSummary[]> => {
    try {
      const response = await apiClient.get('/list');
      return response.data.Items || [];
    } catch (error) {
      console.error('Error fetching configurations:', error);
      throw error;
    }
  },

  // Get a specific configuration by name
  getConfiguration: async (configName: string): Promise<Configuration> => {
    try {
      const response = await apiClient.get(`/get/${configName}`);
      return response.data.Item || {};
    } catch (error) {
      console.error(`Error fetching configuration ${configName}:`, error);
      throw error;
    }
  },

  // Search for configurations
  searchConfigurations: async (query: string): Promise<ConfigurationSummary[]> => {
    try {
      const response = await apiClient.get('/search', {
        params: { q: query },
      });
      return response.data.Items || [];
    } catch (error) {
      console.error('Error searching configurations:', error);
      throw error;
    }
  },
};

export default gogenApi; 