<script lang="ts">
  import { Card, CardBody, CardHeader, CardTitle } from '@sveltestrap/sveltestrap'
  import { profile } from './api/profile.svelte'
  import { theme } from './includes/theme.svelte'
</script>

<Card color="notifiarr" outline class="mb-2" theme={$theme}>
  <CardHeader>
    <CardTitle>Welcome, {$profile?.username}!</CardTitle>
  </CardHeader>
  <CardBody>Select a section from the navigation menu to get started.</CardBody>
</Card>
