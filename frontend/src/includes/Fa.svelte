<script lang="ts">
  import type { IconDefinition } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import { theme } from './theme.svelte'
  import Fa from 'svelte-fa'
  import type { ComponentProps } from 'svelte'

  type FaProps = ComponentProps<Fa>

  interface Props extends FaProps {
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
  }

  const dark = $derived($theme.includes('dark'))
  const { d, c1, c2, d1, d2, ...rest }: Props = $props()
  const primaryColor = $derived(dark && d1 ? d1 : c1)
  const secondaryColor = $derived(dark && d2 ? d2 : c2)
  const icon = $derived(dark && d ? d : rest.icon)
</script>

<Fa {...rest} {icon} {primaryColor} {secondaryColor} />
