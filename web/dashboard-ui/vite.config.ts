/// <reference types="vitest" />
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react(),
  ],
  server: {
    host: '0.0.0.0',
    port: 3000,
    proxy: {
      '/dashboard.v1alpha1': {
        target: 'http://localhost:8443',
      },
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/setupTests.ts',
    coverage: {
      provider: 'v8',
      exclude: [
        'src/__tests__/**',
        'src/proto/**',
        'src/services/DashboardServices.ts'
      ],
      //reporter: ['text', 'json', 'html'],
    },
  },
})
