<script lang="ts" module>
  import { faChartLine } from '@fortawesome/sharp-duotone-regular-svg-icons'
  export const page = {
    id: 'Metrics',
    i: faChartLine,
    c1: 'darkgoldenrod',
    c2: 'darkorange',
    d1: 'darkkhaki',
    d2: 'peachpuff',
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
  <small class="text-muted">Detailed metrics and statistics</small>
</CardFooter>
