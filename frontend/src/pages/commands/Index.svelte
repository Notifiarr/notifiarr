<script lang="ts">
  import { Card, CardHeader, CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { theme } from '../../includes/theme.svelte'
</script>

<Card class="mb-2" theme={$theme}>
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
