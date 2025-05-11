<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { darkMode } from '../lib/darkmode.svelte'
  $: theme = $darkMode ? 'dark' : 'light'
</script>

<Card class="mb-2" {theme}>
  <CardHeader>
    <h2>File Watcher</h2>
  </CardHeader>
  <CardBody>
    <p>Set up and monitor file system events and triggers based on file changes.</p>
  </CardBody>
  <CardFooter>
    <small class="text-muted"
      >Configure automated actions based on file system events</small>
  </CardFooter>
</Card>
