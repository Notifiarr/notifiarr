<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { theme } from '../../includes/theme.svelte'
</script>

<Card class="mb-2" theme={$theme}>
  <CardHeader>
    <h2>Endpoints</h2>
  </CardHeader>
  <CardBody>
    <p>View and manage API endpoints for integration with other systems.</p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Configure API access and integration points</small>
  </CardFooter>
</Card>
