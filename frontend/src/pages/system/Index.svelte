<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { theme } from '../../includes/theme.svelte'
</script>

<Card class="mb-2" theme={$theme}>
  <CardHeader>
    <h2>System</h2>
  </CardHeader>
  <CardBody>
    <p>View and manage system information, resources, and overall health status.</p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Monitor and troubleshoot system-level issues</small>
  </CardFooter>
</Card>
