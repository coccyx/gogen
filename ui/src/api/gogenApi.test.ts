import axios, { AxiosInstance } from 'axios';

// Mock axios before importing gogenApi
jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

// Create mock instance before importing gogenApi
const mockAxiosInstance = {
  get: jest.fn(),
  post: jest.fn(),
  put: jest.fn(),
  delete: jest.fn(),
  defaults: {
    baseURL: '/api',
    headers: {
      'Content-Type': 'application/json'
    }
  },
  interceptors: {
    request: { use: jest.fn(), eject: jest.fn(), clear: jest.fn() },
    response: { use: jest.fn(), eject: jest.fn(), clear: jest.fn() }
  }
} as unknown as jest.Mocked<AxiosInstance>;

// Set up axios.create mock before importing gogenApi
mockedAxios.create.mockReturnValue(mockAxiosInstance);

// Import gogenApi after mock setup
import gogenApi, { ConfigurationSummary, Configuration } from './gogenApi';

describe('gogenApi', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // Reset all mock methods
    mockAxiosInstance.get.mockReset();
    mockAxiosInstance.post.mockReset();
    mockAxiosInstance.put.mockReset();
    mockAxiosInstance.delete.mockReset();
    // Refresh axios.create mock
    mockedAxios.create.mockReturnValue(mockAxiosInstance);
  });

  describe('listConfigurations', () => {
    it('should fetch and return a list of configurations', async () => {
      const mockConfigs: ConfigurationSummary[] = [
        { gogen: 'test1', description: 'Test Config 1' },
        { gogen: 'test2', description: 'Test Config 2' },
      ];

      mockAxiosInstance.get.mockResolvedValueOnce({ data: { Items: mockConfigs } });

      const result = await gogenApi.listConfigurations();

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/list');
      expect(result).toEqual(mockConfigs);
    });

    it('should handle empty response', async () => {
      mockAxiosInstance.get.mockResolvedValueOnce({ data: { Items: [] } });

      const result = await gogenApi.listConfigurations();

      expect(result).toEqual([]);
    });

    it('should handle error', async () => {
      const error = new Error('Network error');
      mockAxiosInstance.get.mockRejectedValueOnce(error);

      await expect(gogenApi.listConfigurations()).rejects.toThrow('Network error');
    });
  });

  describe('getConfiguration', () => {
    it('should fetch and return a specific configuration', async () => {
      const mockConfig: Configuration = {
        gogen: 'test1',
        description: 'Test Config 1',
        config: 'test config content',
      };

      mockAxiosInstance.get.mockResolvedValueOnce({ data: { Item: mockConfig } });

      const result = await gogenApi.getConfiguration('test1');

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/get/test1');
      expect(result).toEqual(mockConfig);
    });

    it('should handle empty response', async () => {
      mockAxiosInstance.get.mockResolvedValueOnce({ data: { Item: {} } });

      const result = await gogenApi.getConfiguration('test1');

      expect(result).toEqual({});
    });

    it('should handle error', async () => {
      const error = new Error('Configuration not found');
      mockAxiosInstance.get.mockRejectedValueOnce(error);

      await expect(gogenApi.getConfiguration('test1')).rejects.toThrow('Configuration not found');
    });
  });

  describe('searchConfigurations', () => {
    it('should search and return matching configurations', async () => {
      const mockConfigs: ConfigurationSummary[] = [
        { gogen: 'test1', description: 'Test Config 1' },
      ];

      mockAxiosInstance.get.mockResolvedValueOnce({ data: { Items: mockConfigs } });

      const result = await gogenApi.searchConfigurations('test');

      expect(mockAxiosInstance.get).toHaveBeenCalledWith('/search', {
        params: { q: 'test' },
      });
      expect(result).toEqual(mockConfigs);
    });

    it('should handle empty search results', async () => {
      mockAxiosInstance.get.mockResolvedValueOnce({ data: { Items: [] } });

      const result = await gogenApi.searchConfigurations('nonexistent');

      expect(result).toEqual([]);
    });

    it('should handle error', async () => {
      const error = new Error('Search failed');
      mockAxiosInstance.get.mockRejectedValueOnce(error);

      await expect(gogenApi.searchConfigurations('test')).rejects.toThrow('Search failed');
    });
  });
}); 