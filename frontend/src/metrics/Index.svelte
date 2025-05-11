<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { darkMode } from '../lib/darkmode.svelte'
  $: theme = $darkMode ? 'dark' : 'light'
</script>

<Card class="mb-2" {theme}>
  <CardHeader>
    <h2>Metrics</h2>
  </CardHeader>
  <CardBody>
    <p>
      View and analyze performance metrics and statistics for your system and
      applications.
    </p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Track system performance and application usage</small>
  </CardFooter>
</Card>
