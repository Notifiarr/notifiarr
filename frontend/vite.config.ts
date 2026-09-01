import { svelteTesting } from '@testing-library/svelte/vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import { defineConfig } from 'vitest/config'

// https://vite.dev/config/
export default defineConfig({
  build: {
    chunkSizeWarningLimit: 2000,
    sourcemap: false,
    rolldownOptions: {
      output: {
        // Vite 8 / Rolldown dropped object-form manualChunks.
        codeSplitting: {
          groups: [
            {
              name: 'bootstrap',
              test: /[\\/]node_modules[\\/](?:@sveltestrap[\\/]sveltestrap|bootstrap)(?:[\\/]|$)/,
            },
            // If the app grows too big, this is a good place to split it:
            // {
            //   name: 'includes',
            //   test: /[\\/]src[\\/]includes[\\/]/,
            // },
          ],
        },
      },
    },
  },
  base: './',
  plugins: [
    svelte({
      dynamicCompileOptions: ({ filename }) => {
        // Enable custom element compilation for files that end with element.svelte.
        return { customElement: filename.endsWith('element.svelte') }
      },
    }),
    svelteTesting(),
  ],
  test: {
    expect: { requireAssertions: true },
    environment: 'jsdom',
    include: ['src/**/*.{test,spec}.{js,ts}'],
    setupFiles: ['./src/test-setup.ts'],
  },
})
