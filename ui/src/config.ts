// Environment-specific configuration
interface Config {
  apiBaseUrl: string;
  githubClientId: string;
  githubRedirectUri: string;
}

export const config: Config = {
  apiBaseUrl: import.meta.env.VITE_API_URL,
  githubClientId: import.meta.env.VITE_GITHUB_CLIENT_ID,
  githubRedirectUri: import.meta.env.VITE_GITHUB_REDIRECT_URI,
}; 