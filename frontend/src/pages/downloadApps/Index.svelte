<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { theme } from '../../includes/theme.svelte'
</script>

<Card class="mb-2" theme="{$theme}}">
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
