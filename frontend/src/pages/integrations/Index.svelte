<script lang="ts" module>
  import { faLinkHorizontal } from '@fortawesome/sharp-duotone-solid-svg-icons'
  export const page = {
    id: 'Integrations',
    i: faLinkHorizontal,
    c1: 'purple',
    c2: 'gray',
    d1: 'magenta',
    d2: 'white',
  }
</script>

<script lang="ts">
  import { CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { _ } from '../../includes/Translate.svelte'
  import Header from '../../includes/Header.svelte'
</script>

<Header {page} />

<CardBody>
  <p>TODO</p>
</CardBody>

<CardFooter>
  <small class="text-muted">Integration statistics</small>
</CardFooter>
