<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { theme } from '../../lib/theme.svelte'
</script>

<Card class="mb-2" {theme}>
  <CardHeader>
    <h2>Service Checks</h2>
  </CardHeader>
  <CardBody>
    <p>Monitor the health and status of your services and applications.</p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Track service availability and performance</small>
  </CardFooter>
</Card>
