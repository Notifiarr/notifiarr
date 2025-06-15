import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vite.dev/config/
export default defineConfig({
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
