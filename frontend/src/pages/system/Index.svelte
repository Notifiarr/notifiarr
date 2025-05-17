<script lang="ts" context="module">
  import { faServer } from '@fortawesome/sharp-duotone-regular-svg-icons'
  export const page = {
    id: 'System',
    icon: faServer,
    c1: 'darkslategray',
    c2: 'dimgray',
    d1: 'lavender',
    d2: 'gainsboro',
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
  <small class="text-muted">System information and status</small>
</CardFooter>
