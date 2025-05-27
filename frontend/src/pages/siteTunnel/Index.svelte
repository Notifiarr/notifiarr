<script lang="ts" module>
  import { faSatellite } from '@fortawesome/sharp-duotone-light-svg-icons'
  export const page = {
    id: 'SiteTunnel',
    i: faSatellite,
    c1: 'blue',
    c2: 'lightblue',
    d1: 'steelblue',
    d2: 'brown',
  }
</script>

<script lang="ts">
  import { Button, Card, CardBody, Input, Spinner, Table } from '@sveltestrap/sveltestrap'
  import T, { _ } from '../../includes/Translate.svelte'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'
  import { profile } from '../../api/profile.svelte'
  import { delay, failure, warning } from '../../includes/util'
  import { faSplotch } from '@fortawesome/sharp-duotone-light-svg-icons'
  import Fa from '../../includes/Fa.svelte'
  import { nav } from '../../navigation/nav.svelte'
  import { getUi } from '../../api/fetch'
  import Stats from './Stats.svelte'

  let pinging = $state(false)
  // Tunnel configuration comes from the website.
  let primaryTunnel = $derived($profile.clientInfo?.user.tunnels?.[0])
  let backupTunnel = $derived($profile.clientInfo?.user.tunnels?.[1])
  // Track ping button and response.
  let pingOutput: Record<number, string> = $state({})
  let pingError = $state('')

  // Tunnel functions
  const pingTunnels = async (e?: Event) => {
    e?.preventDefault()
    pinging = true
    pingError = ''
    const resp = await getUi('tunnel/ping', true)
    if (!resp.ok) {
      warning('Ping Error: ' + resp.body)
      pingError = resp.body
    } else {
      pingOutput = resp.body
    }
    pinging = false
  }

  // Handle form submission
  const submit = async (e: Event) => {
    e.preventDefault()
    if (!primaryTunnel || !backupTunnel)
      return (profile.error = 'phrases.TunnelSaveError')
    await profile.saveTunnels(primaryTunnel, backupTunnel)
  }

  $effect(() => {
    nav.formChanged =
      primaryTunnel !== $profile.clientInfo?.user.tunnels?.[0] ||
      backupTunnel !== $profile.clientInfo?.user.tunnels?.[1]
  })

  const primaryChanged = (socket: any, primary: any): boolean =>
    socket == primary && $profile.clientInfo?.user.tunnels?.[0] != socket
</script>

<Header {page} />

<CardBody>
  <!-- Active Tunnel Card -->
  <Card body color="success" outline>
    <p>{@html $_('SiteTunnel.subText')}</p>
    <p class="mb-0">
      <Fa i={faSplotch} c1="gray" d1="gainsboro" c2="orange" />
      <b><T id="SiteTunnel.phrases.ActiveTunnel" tunnel={$profile.activeTunnel} /></b>
    </p>
  </Card>

  <!-- Primary Tunnel Table -->
  <h4><T id="SiteTunnel.phrases.PrimaryTunnel" /></h4>
  <Table size="sm" hover={true} borderless>
    <tbody class="primary-tunnels">
      {#each $profile.clientInfo?.user.mulery ?? [] as mule, idx}
        <tr
          onclick={() => (primaryTunnel = mule?.socket)}
          style="cursor:pointer;"
          class:changed={primaryChanged(mule!.socket, primaryTunnel)}>
          <td style="width:30px;">
            <Input
              type="radio"
              style="cursor:pointer;"
              bind:group={primaryTunnel}
              value={mule?.socket}
              checked={primaryTunnel === mule?.socket} />
          </td>
          <td class="location">{mule?.location}</td>
          <td style="width:1%;" class="text-info">{pingOutput[idx]}</td>
          <td class="shrink-column">{mule?.socket}</td>
        </tr>
      {/each}
    </tbody>
  </Table>

  <!-- Ping Tunnels Button-->
  <Button color="success" onclick={pingTunnels} disabled={pinging} class="d-inline-block">
    {#if pinging}
      <Spinner size="sm" color="white" />
      <T id="SiteTunnel.phrases.Pinging" />
    {:else}
      <T id="SiteTunnel.phrases.Ping" />
    {/if}
  </Button>
  <div class="text-danger d-inline-block ms-2">{pingError}</div>

  <!-- Backup Tunnel Input -->
  <div class="backup-tunnel">
    <h4><T id="SiteTunnel.phrases.BackupTunnel" /></h4>
    <Input
      type="select"
      bind:value={backupTunnel}
      placeholder={$_('SiteTunnel.phrases.SelectTunnel')}
      class={backupTunnel != $profile.clientInfo?.user.tunnels?.[1] ? 'changed' : ''}>
      {#each $profile.clientInfo?.user.mulery ?? [] as mule}
        <option selected={backupTunnel === mule!.socket} value={mule!.socket}>
          {mule?.location} &nbsp; {mule?.socket}
        </option>
      {/each}
    </Input>
  </div>

  <!-- Tunnel Stats -->
  <Stats />
</CardBody>

<Footer
  {submit}
  successText="SiteTunnel.phrases.TunnelsSaved"
  saveDisabled={!nav.formChanged} />

<style>
  .location {
    width: 1%;
    white-space: nowrap;
    padding-right: 0.5rem;
  }

  .shrink-column {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 0;
  }

  .primary-tunnels :global(.changed td),
  .backup-tunnel :global(.changed) {
    background-color: rgba(205, 92, 92, 0.322) !important;
  }
</style>
