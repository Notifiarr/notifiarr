<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { theme } from '../../includes/theme.svelte'
</script>

<Card class="mb-2" theme={$theme}>
  <CardHeader>
    <h2>Snapshot Apps</h2>
  </CardHeader>
  <CardBody>
    <p>Configure applications that provide system and application state snapshots.</p>
  </CardBody>
  <CardFooter>
    <small class="text-muted">Set up and manage snapshot capabilities</small>
  </CardFooter>
</Card>
