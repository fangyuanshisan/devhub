import { defineConfig } from 'astro/config';
import sitemap from '@astrojs/sitemap';
import vue from '@astrojs/vue';

export default defineConfig({
  site: process.env.FRONTEND_SITE_URL || 'http://127.0.0.1:8090',
  output: 'static',
  outDir: '../frontend',
  integrations: [
    vue(),
    sitemap(),
  ],
  vite: {
    cacheDir: process.env.VITE_CACHE_DIR || 'node_modules/.vite',
    build: {
      assetsInlineLimit: 0,
    },
  },
});
