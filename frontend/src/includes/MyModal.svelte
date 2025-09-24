<!--
 Wrap the sveltestrap Modal component to add a class name when the modal is open.
 Works around a bug. Also gives us a place to add business logic.
 -->
<script lang="ts">
  import { Modal } from '@sveltestrap/sveltestrap'
  import type { ComponentProps, Snippet } from 'svelte'
  import { theme } from './theme.svelte'

  type Props = {
    isOpen?: boolean
    modalClassName?: string
    children?: Snippet
    [key: string]: any
  }

  const { isOpen, modalClassName, children, ...rest }: Props = $props()
</script>

<Modal
  {isOpen}
  modalClassName={modalClassName + (isOpen ? ' show' : '')}
  theme={$theme}
  {...rest}>
  {@render children?.()}
</Modal>
