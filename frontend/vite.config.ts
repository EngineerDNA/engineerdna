import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          // React core and related
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],

          // TanStack Query
          'query-vendor': ['@tanstack/react-query'],

          // Recharts (heavy library for charts)
          'charts-vendor': ['recharts'],
        },
      },
    },
    chunkSizeWarningLimit: 500,
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:3847',
        changeOrigin: true,
      },
    },
  },
});
