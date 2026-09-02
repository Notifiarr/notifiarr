declare module '*.svelte' {
  const component: any
  export default component
}

declare module 'rapidoc'

declare namespace svelteHTML {
  interface IntrinsicElements {
    'rapi-doc': Record<string, unknown>
  }
}
