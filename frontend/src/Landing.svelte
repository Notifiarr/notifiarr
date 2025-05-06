<script>
  import { Card } from '@sveltestrap/sveltestrap'
  import { profile } from './api/profile'
</script>

<Card body color="success" outline class="mb-3">
  <h2>Welcome, {$profile?.username}!</h2>
  <p>Select a section from the navigation menu to get started.</p>
</Card>
