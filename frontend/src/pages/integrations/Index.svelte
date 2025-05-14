<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { darkMode } from '../../lib/darkmode.svelte'
  $: theme = $darkMode ? 'dark' : 'light'
</script>

<Card class="mb-2" {theme}>
  <CardHeader>
    <h2>Integrations</h2>
  </CardHeader>
  <CardBody>
    <p>Configure third-party integrations and connectivity with external services.</p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Connect Notifiarr with other platforms and services</small>
  </CardFooter>
</Card>
