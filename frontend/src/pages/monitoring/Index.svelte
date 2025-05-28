<script lang="ts" module>
  import { faDisplayChartUpCircleCurrency } from '@fortawesome/sharp-duotone-regular-svg-icons'
  export const page = {
    id: 'Monitoring',
    i: faDisplayChartUpCircleCurrency,
    c1: 'darkslateblue',
    c2: 'gainsboro',
    d1: 'papayawhip',
    d2: 'paleturquoise',
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
  <small class="text-muted">System and application monitoring</small>
</CardFooter>
