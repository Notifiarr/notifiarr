<script lang="ts" module>
  import { faStarfighter } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import { CardBody } from '@sveltestrap/sveltestrap'
  import Header from '../../includes/Header.svelte'
  import { profile } from '../../api/profile.svelte'
  import Triggers from './Triggers.svelte'

  export const page = {
    id: 'Actions',
    i: faStarfighter,
    c1: 'lightcoral',
    c2: 'palevioletred',
    d1: 'thistle',
    d2: 'gainsboro',
  }
</script>

<Header {page} />

<CardBody>
  {#if $profile.timers && Object.keys($profile.timers).length > 0}
    <Triggers type="Timers" obj={$profile.timers ?? {}} />
  {/if}
  {#if $profile.schedules && Object.keys($profile.schedules).length > 0}
    <Triggers type="Schedules" obj={$profile.schedules ?? {}} />
  {/if}
  {#if $profile.triggers && Object.keys($profile.triggers).length > 0}
    <Triggers type="Triggers" obj={$profile.triggers ?? {}} />
  {/if}
</CardBody>
