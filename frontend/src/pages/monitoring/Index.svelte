<script lang="ts" module>
  import { page } from './page.svelte'
  export { page }
</script>

<script lang="ts">
  import {
    CardBody,
    Button,
    Dropdown,
    DropdownToggle,
    DropdownMenu,
    DropdownItem,
    Popover,
  } from '@sveltestrap/sveltestrap'
  import T, { _ } from '../../includes/Translate.svelte'
  import Header from '../../includes/Header.svelte'
  import { profile } from '../../api/profile.svelte'
  import Table from './Table.svelte'
  import { age, warning } from '../../includes/util'
  import { getUi } from '../../api/fetch'
  import { Monitor as chk } from './page.svelte'
  import Cards from './Cards.svelte'
  import Fa from '../../includes/Fa.svelte'
  import {
    faListCheck,
    faTableCellsLarge,
    faArrowDownToBracket,
    faArrowUpFromBracket,
    faClock,
  } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import Cookies from 'js-cookie'
  import { slide } from 'svelte/transition'
  import { theme as thm } from '../../includes/theme.svelte'

  const theme = $derived($thm)

  const setRefreshInterval = (interval: number) => {
    refreshInterval = interval
    Cookies.set('refreshInterval', interval.toString())
    nextRefresh = refreshInterval ? profile.now + refreshInterval : 0
  }

  let disabled = $state(false)
  const start = async (e: Event) => await toggle(e, 'start')
  const stop = async (e: Event) => await toggle(e, 'stop')
  const toggle = async (e: Event, action: string) => {
    e.preventDefault()
    disabled = true
    const resp = await getUi('services/' + action, false)
    if (!resp.ok) warning($_('monitoring.' + action + 'Failed'))
    else await chk.updateBackend(e)
    disabled = false
  }

  const toggleCards = () => Cookies.set('showCards', (showCards = !showCards).toString())
  let showCards = $state(Cookies.get('showCards') === 'true')
  const icon = $derived(showCards ? faListCheck : faTableCellsLarge)

  let showOutput = $state<Record<string, boolean>>({})
  let showOutputAll = $state(false)
  const showAllI = $derived(showOutputAll ? faArrowUpFromBracket : faArrowDownToBracket)

  const toggleOutput = (e: Event) => {
    e.preventDefault()
    showOutputAll = !showOutputAll
    Object.keys(showOutput).forEach(id => (showOutput[id] = showOutputAll))
  }

  let refreshInterval = $state(Number(Cookies.get('refreshInterval')) ?? 0)
  // svelte-ignore state_referenced_locally
  let nextRefresh = $state(refreshInterval ? profile.now + refreshInterval : 0)

  $effect(() => {
    if (nextRefresh && profile.now >= nextRefresh) {
      chk.updateBackend(new Event('refresh'))
      nextRefresh = profile.now + refreshInterval
    }
  })
</script>

<Header {page}>
  <T id="monitoring.ConfigureServices" />
  <div class="float-end m-2 d-inline-block toggle-buttons">
    {#if showCards}
      <Popover target="showAll-button" placement="left" trigger="hover" {theme}>
        <T id={`monitoring.${showOutputAll ? 'hideAllOutputs' : 'showAllOutputs'}`} />
      </Popover>
      <Button color="info" outline onclick={toggleOutput} id="showAll-button">
        <Fa i={showAllI} c1="blue" d1="slateblue" scale="1.5" />
      </Button>
    {/if}
    <Popover target="refresh-button" placement="left" trigger="hover" {theme}>
      <T id="monitoring.refreshInterval" />
    </Popover>
    <Dropdown class="d-inline-block">
      <DropdownToggle color="warning" outline class="interval-picker" id="refresh-button">
        {#if refreshInterval}
          {age(refreshInterval, true)}
        {:else}
          <Fa i={faClock} c1="slateblue" d1="wheat" scale="1.5" />
        {/if}
      </DropdownToggle>
      <DropdownMenu>
        <DropdownItem header><T id="monitoring.refreshInterval" /></DropdownItem>
        <DropdownItem
          onclick={() => setRefreshInterval(0)}
          active={refreshInterval === 0}>
          <T id="words.select-option.Off" />
        </DropdownItem>
        {#each [10, 15, 45, 30, 60, 120, 240, 300, 600, 900, 1200, 1800] as interval}
          <DropdownItem
            onclick={() => setRefreshInterval(interval * 1000)}
            active={refreshInterval === interval * 1000}>
            {age(interval * 1000, true)}
          </DropdownItem>
        {/each}
      </DropdownMenu>
    </Dropdown>
    <Popover target="cards-button" placement="left" trigger="hover" {theme}>
      <T id={`monitoring.${showCards ? 'cardsView' : 'classicView'}`} />
    </Popover>
    <Button color="info" outline onclick={toggleCards} id="cards-button">
      <Fa i={icon} c1="blue" d1="slateblue" scale="1.5" />
    </Button>
  </div>
  {#if nextRefresh}
    <br />
    <b class="text-primary">
      <T
        id="monitoring.nextRefresh"
        timeDuration={age(nextRefresh - profile.now, true)} />
    </b>
  {/if}
  <div style="clear:both"></div>
</Header>

<CardBody>
  {#if $profile.checkDisabled}
    <h5 class="text-danger"><T id="monitoring.ChecksDisabled" /></h5>
    <p><T id="monitoring.EnableOnConfig" /></p>
  {:else if !$profile.checkRunning}
    <h5 class="text-danger d-inline-block"><T id="monitoring.ChecksStopped" /></h5>
    <Button color="success" outline class="float-end" onclick={start} {disabled}>
      <T id="monitoring.StartChecks" />
    </Button>
    <div style="clear:both"></div>
  {:else}
    <Button color="primary" outline onclick={chk.updateBackend} disabled={chk.refresh}>
      <T id={`phrases.${chk.refresh ? 'UpdatingBackEnd' : 'UpdateBackend'}`} />
    </Button>
    <Button color="danger" outline class="float-end" onclick={stop} {disabled}>
      <T id="monitoring.StopChecks" />
    </Button>
  {/if}

  <div class="mt-2">
    {#if showCards}
      <div transition:slide={{ duration: 420 }}><Cards bind:showOutput /></div>
    {:else}
      <div transition:slide={{ duration: 420 }}><Table /></div>
    {/if}
  </div>
</CardBody>

<style>
  .toggle-buttons :global(.interval-picker) {
    min-width: 45px !important;
    padding: 0.375rem !important;
  }
</style>
