// Environment-specific configuration
interface Config {
  apiBaseUrl: string;
  githubClientId: string;
  githubRedirectUri: string;
}

export const config: Config = {
  apiBaseUrl: process.env.VITE_API_URL ?? '',
  githubClientId: process.env.VITE_GITHUB_CLIENT_ID ?? '',
  githubRedirectUri: process.env.VITE_GITHUB_REDIRECT_URI ?? '',
};
