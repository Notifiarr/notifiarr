<svelte:options customElement={{ tag: 'spoiler-toggle', shadow: 'open' }} />

<span class="wrapper">
  <span class="spoiler">
    <slot />
  </span>
</span>

<style>
  .wrapper {
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
