<script>
  import { Card } from '@sveltestrap/sveltestrap'
  import { profile } from './lib/login'
</script>

<Card body color="success" outline>
    <h1>Welcome, {$profile?.username}!</h1>
</Card>
