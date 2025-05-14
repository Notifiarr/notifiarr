<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { theme } from '../../lib/theme.svelte'
</script>

<Card class="mb-2" {theme}>
  <CardHeader>
    <h2>Media Apps</h2>
  </CardHeader>
  <CardBody>
    <p>Configure and manage media server applications like Plex, Jellyfin, and Emby.</p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Manage your media server integrations</small>
  </CardFooter>
</Card>
