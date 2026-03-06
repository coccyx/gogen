// Environment-specific configuration
interface Config {
  apiBaseUrl: string;
  githubClientId: string;
  githubRedirectUri: string;
}

function getEnvValue(name: string): string {
  const value = process.env[name];
  return value ?? '';
}

export const config: Config = {
  apiBaseUrl: getEnvValue('VITE_API_URL'),
  githubClientId: getEnvValue('VITE_GITHUB_CLIENT_ID'),
  githubRedirectUri: getEnvValue('VITE_GITHUB_REDIRECT_URI'),
};
