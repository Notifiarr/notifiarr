<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { darkMode } from '../lib/darkmode.svelte'
  $: theme = $darkMode ? 'dark' : 'light'
</script>

<Card class="mb-2" {theme}>
  <CardHeader>
    <h2>Site Tunnel</h2>
  </CardHeader>
  <CardBody>
    <p>Configure and manage secure tunnels for remote access.</p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Set up secure connections to your Notifiarr client</small>
  </CardFooter>
</Card>
