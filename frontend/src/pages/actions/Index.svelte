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
  import type { Row } from './run'
  import T from '../../includes/Translate.svelte'

  const toggleCards = () =>
    Cookies.set('showActionCards', (showCards = !showCards).toString())
  let showCards = $state(Cookies.get('showActionCards') === 'true')
  const icon = $derived(showCards ? faListCheck : faTableCellsLarge)

  let filter = $state('')

  const timers = $derived<Row[]>(
    Object.values($profile.timers ?? []).map(t => ({ ...t, type: 'Timer' })),
  )
  const schedules = $derived<Row[]>(
    Object.values($profile.schedules ?? []).map(s => ({ ...s, type: 'Schedule' })),
  )
  const triggers = $derived<Row[]>(
    Object.values($profile.triggers ?? []).map(t => ({ ...t, type: 'Trigger' })),
  )
</script>

<Header {page}>
  <InputGroup>
    {#if filter}
      <Button color="warning" outline onclick={() => (filter = '')}>
        <Fa i={faX} scale="1.5" />
      </Button>
    {/if}
    <Input bind:value={filter} placeholder="Filter" />
    <Button color="info" outline onclick={toggleCards} id="cards-btn">
      <Fa i={icon} c1="lightblue" d1="slateblue" scale="1.5" />
    </Button>
    <Popover target="cards-btn" placement="left" trigger="hover" theme={$thm}>
      <T id={`monitoring.${showCards ? 'classicView' : 'cardsView'}`} />
    </Popover>
  </InputGroup>
</Header>

<CardBody>
  {#if showCards}
    {#if timers.length > 0}<Triggers type="Timers" rows={timers} {filter} />{/if}
    {#if schedules.length > 0}<Triggers type="Schedules" rows={schedules} {filter} />{/if}
    {#if triggers.length > 0}<Triggers type="Triggers" rows={triggers} {filter} />{/if}
  {:else}
    <TableView rows={[...timers, ...schedules, ...triggers]} {filter} />
  {/if}
</CardBody>
