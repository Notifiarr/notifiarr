<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { darkMode } from '../lib/darkmode.svelte'
  $: theme = $darkMode ? 'dark' : 'light'
</script>

<Card class="mb-2" {theme}>
  <CardHeader>
    <h2>System</h2>
  </CardHeader>
  <CardBody>
    <p>View and manage system information, resources, and overall health status.</p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Monitor and troubleshoot system-level issues</small>
  </CardFooter>
</Card>
