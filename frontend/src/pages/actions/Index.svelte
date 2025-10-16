<script lang="ts" module>
  import {
    faStarfighter,
    faX,
    faListCheck,
    faTableCellsLarge,
  } from '@fortawesome/sharp-duotone-solid-svg-icons'

  export const page = {
    id: 'Actions',
    i: faStarfighter,
    c1: 'lightcoral',
    c2: 'palevioletred',
    d1: 'thistle',
    d2: 'gainsboro',
  }
</script>

<script lang="ts">
  import { theme as thm } from '../../includes/theme.svelte'
  import { Button, CardBody, Input, InputGroup, Popover } from '@sveltestrap/sveltestrap'
  import Header from '../../includes/Header.svelte'
  import { profile } from '../../api/profile.svelte'
  import Triggers from './Triggers.svelte'
  import Fa from '../../includes/Fa.svelte'
  import Cookies from 'js-cookie'
  import TableView from './TableView.svelte'
  import T, { _ } from '../../includes/Translate.svelte'
  import type { TriggerInfo } from '../../api/notifiarrConfig'
  import { val } from './run'
  import { slide } from 'svelte/transition'

  const toggleCards = () =>
    Cookies.set('showActionCards', (showCards = !showCards).toString())
  let showCards = $state(Cookies.get('showActionCards') === 'true')
  const icon = $derived(showCards ? faListCheck : faTableCellsLarge)

  let filter = $state('')

  const sift = (rows?: TriggerInfo[]): TriggerInfo[] =>
    rows?.filter(row => val(row).toLowerCase().includes(filter.toLowerCase())) ?? []
</script>

<Header {page}>
  <InputGroup>
    {#if filter}
      <Button color="warning" outline onclick={() => (filter = '')}>
        <Fa i={faX} scale="1.2" />
      </Button>
    {/if}
    <!-- Filter input. Clear button is above.-->
    <Input bind:value={filter} placeholder={$_('Actions.titles.Filter')} />
    <!-- Toggle cards view. -->
    <Button color="info" outline onclick={toggleCards} id="cards-btn">
      <Fa i={icon} c1="lightblue" d1="slateblue" scale="1.5" />
    </Button>
    <Popover target="cards-btn" placement="left" trigger="hover" theme={$thm}>
      <T id={`monitoring.${showCards ? 'classicView' : 'cardsView'}`} />
    </Popover>
  </InputGroup>
</Header>

{#if showCards}
  <div transition:slide>
    <CardBody class="py-0">
      <Triggers type="Timers" rows={sift($profile.timers)} />
      <Triggers type="Schedules" rows={sift($profile.schedules)} />
      <Triggers type="Triggers" rows={sift($profile.triggers)} />
    </CardBody>
  </div>
{:else}
  <div transition:slide>
    <TableView
      rows={[
        sift($profile.timers),
        sift($profile.schedules),
        sift($profile.triggers),
      ].flat()} />
  </div>
{/if}
