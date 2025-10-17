<svelte:options customElement={{ tag: 'spoiler-toggle', shadow: 'open' }} />

<script lang="ts">
  import { get } from 'svelte/store'
  import { _ } from './Translate.svelte'

  let isOpen = $state(false)
  const toggle = (e: MouseEvent | KeyboardEvent) => {
    e.preventDefault()
    e.stopPropagation()
    isOpen = !isOpen
  }
</script>

<a
  href="#toggle-spoiler-visibility"
  title={get(_)('phrases.ToggleSpoilerVisibility')}
  class="wrapper"
  style={isOpen ? 'border:none' : ''}
  onclick={toggle}
  onkeydown={toggle}>
  <span class="spoiler" style={isOpen ? 'opacity:1' : ''}>
    <slot />
  </span>
</a>

<style>
  .wrapper {
    text-decoration: none;
    color: inherit;
    border: 1px solid var(--bs-border-color);
    border-radius: 0.25rem;
    transition: border 0.3s ease-in-out;
  }
  .wrapper:hover {
    border: none;
  }
  .spoiler {
    opacity: 0;
    transition: opacity 1s ease;
  }
  .spoiler:hover {
    opacity: 1;
    transition: opacity 0.4s ease;
  }
</style>
