import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vite.dev/config/
export default defineConfig({
  build: {
    chunkSizeWarningLimit: 2000,
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks: {
          bootstrap: ['@sveltestrap/sveltestrap', 'bootstrap/dist/css/bootstrap.min.css'],
          fontawesome: [
            '@fortawesome/sharp-duotone-light-svg-icons',
            '@fortawesome/sharp-duotone-regular-svg-icons',
            '@fortawesome/sharp-duotone-solid-svg-icons',
            '@fortawesome/free-brands-svg-icons',
            // 'svelte-fa',
          ],
          // If the app grows too big, this is a good place to split it:
          // includes: [
          //   './src/includes/formsTracker.svelte.ts',
          //   './src/includes/instanceValidator.ts',
          //   './src/includes/util.ts',
          //   './src/includes/theme.svelte.ts',
          // ],
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
