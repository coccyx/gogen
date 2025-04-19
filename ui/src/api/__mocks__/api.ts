import { Configuration, ConfigurationSummary } from '../gogenApi';

export const mockConfigurations: ConfigurationSummary[] = [
  {
    gogen: 'Test Config 1',
    description: 'A test configuration',
  },
  {
    gogen: 'Test Config 2',
    description: 'Another test configuration',
  },
];

export const mockConfigurationDetails: Configuration[] = mockConfigurations.map(config => ({
  ...config,
  config: 'config: test',
  samples: [],
  raters: [],
  mix: [],
  generators: [],
  global: {},
  templates: [],
}));

export const api = {
  listConfigurations: jest.fn().mockResolvedValue(mockConfigurations),
  getConfiguration: jest.fn().mockImplementation((name: string) => 
    Promise.resolve(mockConfigurationDetails.find(c => c.gogen === name))
  ),
  searchConfigurations: jest.fn().mockImplementation((query: string) => 
    Promise.resolve(mockConfigurations.filter(c => 
      c.gogen.toLowerCase().includes(query.toLowerCase()) || 
      c.description.toLowerCase().includes(query.toLowerCase())
    ))
  ),
}; 