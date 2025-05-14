<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { darkMode } from '../../lib/darkmode.svelte'
  $: theme = $darkMode ? 'dark' : 'light'
</script>

<Card class="mb-2" {theme}>
  <CardHeader>
    <h2>Endpoints</h2>
  </CardHeader>
  <CardBody>
    <p>View and manage API endpoints for integration with other systems.</p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Configure API access and integration points</small>
  </CardFooter>
</Card>
