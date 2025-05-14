<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { theme } from '../../includes/theme.svelte'
</script>

<Card class="mb-2" theme={$theme}>
  <CardHeader>
    <h2>Metrics</h2>
  </CardHeader>
  <CardBody>
    <p>View and analyze system and application performance metrics.</p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Monitor and analyze system performance data</small>
  </CardFooter>
</Card>
