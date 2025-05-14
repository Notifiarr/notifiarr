<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { theme } from '../../lib/theme.svelte'
</script>

<Card class="mb-2" {theme}>
  <CardHeader>
    <h2>Monitoring</h2>
  </CardHeader>
  <CardBody>
    <p>Set up and view system and application monitoring dashboards.</p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Monitor the health and status of your systems</small>
  </CardFooter>
</Card>
