import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

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
            {
              name: 'fontawesome',
              test: /[\\/]node_modules[\\/]@fortawesome[\\/]/,
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
  ],
})
