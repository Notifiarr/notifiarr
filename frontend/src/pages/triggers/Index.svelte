<script lang="ts" context="module">
  import { faStarfighter } from '@fortawesome/sharp-duotone-solid-svg-icons'
  export const page = {
    id: 'Triggers',
    icon: faStarfighter,
    c1: 'lightcoral',
    c2: 'palevioletred',
    d1: 'thistle',
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
  <small class="text-muted">Event trigger management</small>
</CardFooter>
