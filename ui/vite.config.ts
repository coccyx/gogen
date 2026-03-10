import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');

  return {
    plugins: [
      react(),
      {
        name: 'configure-server',
        configureServer(server) {
          server.middlewares.use((req, res, next) => {
            res.setHeader('Cross-Origin-Embedder-Policy', 'require-corp');
            res.setHeader('Cross-Origin-Opener-Policy', 'same-origin');
            next();
          });
        }
      }
    ],
    define: {
      'process.env.VITE_API_URL': JSON.stringify(env.VITE_API_URL ?? ''),
      'process.env.VITE_GITHUB_CLIENT_ID': JSON.stringify(env.VITE_GITHUB_CLIENT_ID ?? ''),
      'process.env.VITE_GITHUB_REDIRECT_URI': JSON.stringify(env.VITE_GITHUB_REDIRECT_URI ?? ''),
    },
    server: {
      port: 3000,
      open: true,
      proxy: {
        '/api': {
          target: mode === 'staging'
            ? 'https://staging-api.gogen.io'
            : mode === 'production'
              ? 'https://api.gogen.io'
              : 'http://localhost:4000',
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api/, '/v1'),
          secure: mode !== 'development',
        }
      },
      fs: {
        allow: ['..']
      }
    },
    build: {
      outDir: 'dist',
      sourcemap: true,
      rollupOptions: {
        output: {
          manualChunks: {
            monaco: ['@monaco-editor/react'],
          },
        },
      },
    },
    optimizeDeps: {
      exclude: ['@wasmer/sdk']
    },
    assetsInclude: ['**/*.wasm'],
  };
});
