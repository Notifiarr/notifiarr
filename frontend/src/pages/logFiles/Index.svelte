<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { darkMode } from '../../lib/darkmode.svelte'
  $: theme = $darkMode ? 'dark' : 'light'
</script>

<Card class="mb-2" {theme}>
  <CardHeader>
    <h2>Log Files</h2>
  </CardHeader>
  <CardBody>
    <p>View and search through application and system log files for troubleshooting.</p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Access and analyze log data for diagnostics</small>
  </CardFooter>
</Card>
