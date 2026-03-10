import axios from 'axios';
import { config } from '../config';

// Create an axios instance with default config
const apiClient = axios.create({
  baseURL: config.apiBaseUrl,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add request interceptor to include auth token
apiClient.interceptors.request.use((requestConfig) => {
  const token = localStorage.getItem('github_token');
  if (token) {
    requestConfig.headers.Authorization = `token ${token}`;
  }
  return requestConfig;
});

// Define interfaces for API responses
export interface ConfigurationSummary {
  gogen: string;
  description: string;
  owner?: string;
}

export interface Configuration extends ConfigurationSummary {
  config?: string;
  samples?: any[];
  raters?: any[];
  mix?: any[];
  generators?: any[];
  global?: any;
  templates?: any[];
  s3Path?: string;
}

export interface OAuthResponse {
  access_token: string;
  token_type: string;
  user: {
    login: string;
    avatar_url: string;
    id: number;
    name?: string;
    email?: string;
  };
}

export interface UpsertRequest {
  name: string;
  description: string;
  config: string;
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

  // Exchange OAuth code for access token
  exchangeOAuthCode: async (code: string, state: string): Promise<OAuthResponse> => {
    try {
      const response = await apiClient.post('/auth/github', { code, state });
      return response.data;
    } catch (error) {
      console.error('Error exchanging OAuth code:', error);
      throw error;
    }
  },

  // Get current user's configurations
  getMyConfigurations: async (): Promise<ConfigurationSummary[]> => {
    try {
      const response = await apiClient.get('/my-configs');
      return response.data.Items || [];
    } catch (error) {
      console.error('Error fetching my configurations:', error);
      throw error;
    }
  },

  // Create or update a configuration
  upsertConfiguration: async (data: UpsertRequest): Promise<any> => {
    try {
      const response = await apiClient.post('/upsert', data);
      return response.data;
    } catch (error) {
      console.error('Error upserting configuration:', error);
      throw error;
    }
  },

  // Delete a configuration
  deleteConfiguration: async (configPath: string): Promise<any> => {
    try {
      const response = await apiClient.delete(`/delete/${configPath}`);
      return response.data;
    } catch (error) {
      console.error(`Error deleting configuration ${configPath}:`, error);
      throw error;
    }
  },
};

export default gogenApi;
