/// <reference types="vitest" />
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import { viteSingleFile } from 'vite-plugin-singlefile'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), viteSingleFile()],
  build: {
    target: 'es2018',
  },
  test: {
    globals: true,
    environment: 'jsdom',
  },
})
