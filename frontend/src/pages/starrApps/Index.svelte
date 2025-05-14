<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { theme } from '../../lib/theme.svelte'
</script>

<Card class="mb-2" {theme}>
  <CardHeader>
    <h2>Starr Apps</h2>
  </CardHeader>
  <CardBody>
    <p>
      Monitor and manage your Starr application suite including Sonarr, Radarr, Lidarr,
      and more.
    </p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">View and control your media automation applications</small>
  </CardFooter>
</Card>
