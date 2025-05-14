<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { theme } from '../../includes/theme.svelte'
</script>

<Card class="mb-2" theme={$theme}>
  <CardHeader>
    <h2>Log Files</h2>
  </CardHeader>
  <CardBody>
    <p>View and search through application and system log files for troubleshooting.</p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Access and analyze log data for diagnostics</small>
  </CardFooter>
</Card>
