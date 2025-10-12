<script lang="ts">
  import type { IconDefinition } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import { theme } from './theme.svelte'
  import Fa from 'svelte-fa'
  import type { ComponentProps } from 'svelte'

  export interface Props extends Omit<ComponentProps<Fa>, 'icon'> {
    /** The icon to display */
    i: IconDefinition
    /**
     * The icon to display in dark mode.
     */
    d?: IconDefinition
    /**
     * The primary and default color of the icon.
     */
    c1?: string
    /**
     * The secondary color of the icon.
     */
    c2?: string
    /**
     * The primary color of the icon in dark mode.
     */
    d1?: string
    /**
     * The secondary color of the icon in dark mode.
     */
    d2?: string
    /** Every other prop. */
    [key: string]: any
  }

  const { d, c1, c2, d1 = c1, d2 = c2, ...rest }: Props = $props()
  const primaryColor = $derived(theme.isDark ? d1 : c1)
  const secondaryColor = $derived(theme.isDark ? d2 : c2)
  const icon = $derived(theme.isDark && d ? d : rest.i)
  const onclick = $derived((e: Event) => (e.preventDefault(), rest.onclick()))
</script>

{#if rest.onclick || rest.href}
  <a href={rest.href ?? '#anotherPage'} {onclick}>
    <Fa {...rest} {icon} {primaryColor} {secondaryColor} id={'fa-icon' + rest.id} />
  </a>
{:else}
  <Fa {...rest} {icon} {primaryColor} {secondaryColor} id={'fa-icon' + rest.id} />
{/if}
