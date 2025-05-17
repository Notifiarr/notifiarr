<script lang="ts" context="module">
  import { faHouseChimneyHeart } from '@fortawesome/sharp-duotone-light-svg-icons'
  export const page = {
    id: 'Landing',
    icon: faHouseChimneyHeart,
    c1: 'seagreen',
    c2: 'darkgreen',
    d1: 'lightblue',
    d2: 'silver',
  }
</script>

<script lang="ts">
  import { CardBody } from '@sveltestrap/sveltestrap'
  import { profile } from './api/profile.svelte'
  import Header from './includes/Header.svelte'
</script>

<Header {page} />
<CardBody>
  <h4>Welcome, {$profile?.username}!</h4>
  <p>Select a section from the navigation menu to get started.</p>
</CardBody>
