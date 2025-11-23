import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [svelte()],
  server: {
    port: 3000,
    host: true,  // This allows external connections
    open: false  // Don't automatically open browser
  },
  build: {
    outDir: '../dist'
  }
});