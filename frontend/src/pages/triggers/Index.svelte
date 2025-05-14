<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { theme } from '../../lib/theme.svelte'
</script>

<Card class="mb-2" {theme}>
  <CardHeader>
    <h2>Triggers</h2>
  </CardHeader>
  <CardBody>
    <p>Configure and manage event triggers for automation.</p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Set up automated responses to system events</small>
  </CardFooter>
</Card>
