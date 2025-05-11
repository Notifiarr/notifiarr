<script>
  import { Card, CardBody, CardHeader, CardTitle } from '@sveltestrap/sveltestrap'
  import { profile } from './api/profile'
  import { darkMode } from './lib/darkmode.svelte'
  $: theme = $darkMode ? 'dark' : 'light'
</script>

<Card color="notifiarr" outline class="mb-2" {theme}>
  <CardHeader>
    <CardTitle>Welcome, {$profile?.username}!</CardTitle>
  </CardHeader>
  <CardBody>Select a section from the navigation menu to get started.</CardBody>
</Card>
