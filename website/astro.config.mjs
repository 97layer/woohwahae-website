import { defineConfig } from 'astro/config';
import vercel from '@astrojs/vercel/static';

// 97LAYER OS Website Configuration
// Minimalist archive following WOOHWAHAE aesthetic principles

export default defineConfig({
  output: 'static',
  adapter: vercel({
    webAnalytics: {
      enabled: false  // No tracking, anti-algorithm
    }
  }),
  site: 'https://woohwahae.com',  // Update with actual domain
  markdown: {
    shikiConfig: {
      theme: 'min-light',  // Minimal syntax highlighting
      wrap: false
    }
  }
});