<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { darkMode } from '../lib/darkmode.svelte'
  $: theme = $darkMode ? 'dark' : 'light'
</script>

<Card class="mb-2" {theme}>
  <CardHeader>
    <h2>Service Checks</h2>
  </CardHeader>
  <CardBody>
    <p>Configure and monitor service health checks for various applications.</p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Ensure your services are running properly</small>
  </CardFooter>
</Card>
