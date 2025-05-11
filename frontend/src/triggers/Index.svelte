<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { darkMode } from '../lib/darkmode.svelte'
  $: theme = $darkMode ? 'dark' : 'light'
</script>

<Card class="mb-2" {theme}>
  <CardHeader>
    <h2>Triggers</h2>
  </CardHeader>
  <CardBody>
    <p>Set up event-based triggers to automate actions based on system events.</p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Create automated workflows based on events</small>
  </CardFooter>
</Card>
