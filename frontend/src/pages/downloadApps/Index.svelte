<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { darkMode } from '../../lib/darkmode.svelte'
  $: theme = $darkMode ? 'dark' : 'light'
</script>

<Card class="mb-2" {theme}>
  <CardHeader>
    <h2>Download Apps</h2>
  </CardHeader>
  <CardBody>
    <p>
      Manage download clients such as SABnzbd, NZBGet, Transmission, qBittorrent, and
      others.
    </p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Configure and monitor your download applications</small>
  </CardFooter>
</Card>
