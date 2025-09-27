<!--
 Wrap the sveltestrap Modal component to add a class name when the modal is open.
 Works around a bug. Also gives us a place to add business logic.
 -->
<script lang="ts">
  import { Modal } from '@sveltestrap/sveltestrap'
  import { theme } from './theme.svelte'

  type Props = { isOpen?: boolean; children?: any; [key: string]: any }
  const { isOpen, children, ...rest }: Props = $props()
  const modalClassName = $derived(rest.modalClassName + (isOpen ? ' show' : ''))
</script>

<Modal {...{ ...rest, isOpen, modalClassName, theme: $theme }}>
  {@render children?.()}
</Modal>
