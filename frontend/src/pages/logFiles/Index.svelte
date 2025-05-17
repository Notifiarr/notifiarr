<script lang="ts" context="module">
  import { faPrintMagnifyingGlass } from '@fortawesome/sharp-duotone-solid-svg-icons'
  export const page = {
    id: 'LogFiles',
    icon: faPrintMagnifyingGlass,
    c1: 'midnightblue',
    c2: 'darkslategray',
    d1: 'lightgreen',
    d2: 'gold',
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
  <small class="text-muted">Log file management</small>
</CardFooter>
