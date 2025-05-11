<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { darkMode } from '../lib/darkmode.svelte'
  $: theme = $darkMode ? 'dark' : 'light'
</script>

<Card class="mb-2" {theme}>
  <CardHeader>
    <h2>Commands</h2>
  </CardHeader>
  <CardBody>
    <p>Set up and execute custom commands and scripts from the Notifiarr interface.</p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Manage automation through custom commands</small>
  </CardFooter>
</Card>
