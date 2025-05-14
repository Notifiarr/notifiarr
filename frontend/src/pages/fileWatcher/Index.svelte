<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { theme } from '../../includes/theme.svelte'
</script>

<Card class="mb-2" theme={$theme}>
  <CardHeader>
    <h2>File Watcher</h2>
  </CardHeader>
  <CardBody>
    <p>Monitor and respond to changes in specified files and directories.</p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Track file system changes and trigger actions</small>
  </CardFooter>
</Card>
