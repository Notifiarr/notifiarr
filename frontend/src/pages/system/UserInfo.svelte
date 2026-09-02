<script lang="ts">
  import Bug from 'phosphor-svelte/lib/Bug'
  import Butterfly from 'phosphor-svelte/lib/Butterfly'
  import BugBeetle from 'phosphor-svelte/lib/BugBeetle'
  import Cat from 'phosphor-svelte/lib/Cat'
  import Horse from 'phosphor-svelte/lib/Horse'
  import Bird from 'phosphor-svelte/lib/Bird'
  import { Table } from '@sveltestrap/sveltestrap'
  import { profile } from '../../api/profile.svelte'
  import Header from '../../includes/Helem.svelte'
  import T from '../../includes/Translate.svelte'
  import type { Props } from '../../includes/Fa.svelte'

  const invalidAPIKey = $derived(
    $profile.apiKeyValid === false && $profile.config?.apiKey?.length === 36,
  )

  /** Fa.logo is a boolean size flag; Helem.logo is an image URL. Do not spread logo. */
  type Icon = Omit<Props, 'logo' | 'page' | 'children'>

  const icon = $derived.by((): Icon => {
    const user = $profile.clientInfo?.user
    if (user?.subscriber) {
      return {
        i: user.devAllowed ? Horse : Cat,
        c1: 'darkgoldenrod',
        c2: 'lightcoral',
        d1: 'goldenrod',
        d2: 'azure',
      }
    }
    if (user?.patron) {
      return { i: user.devAllowed ? BugBeetle : Butterfly, c1: 'orange', c2: 'wheat' }
    }
    if (user?.devAllowed) {
      return { i: Bird, c1: 'purple', c2: 'coral' }
    }
    return { i: Bug, c1: 'green', c2: 'lightgreen', d2: 'white' }
  })
</script>

<!-- User Section -->
<Header id="UserInformation" {...icon} page="ClientInfo" />

<Table>
  <tbody>
    {#if invalidAPIKey}
      <tr>
        <td colspan="2"><T id="system.UserInformation.InvalidAPIKey" /></td>
      </tr>
    {:else}
      <tr>
        <th><T id="system.UserInformation.Patron" /></th>
        <td>{$profile.clientInfo?.user?.patron ? 'Yes' : 'No'}</td>
      </tr>
      <tr>
        <th><T id="system.UserInformation.Subscriber" /></th>
        <td>{$profile.clientInfo?.user?.subscriber ? 'Yes' : 'No'}</td>
      </tr>
      {#if $profile.clientInfo?.user?.subscriber}
        <tr>
          <th><T id="system.UserInformation.AbsoluteBadAss" /></th>
          <td><T id="system.UserInformation.YesYesYouAre" /></td>
        </tr>
      {:else if $profile.clientInfo?.user?.patron}
        <tr>
          <th><T id="system.UserInformation.HellaAwesome" /></th>
          <td><T id="system.UserInformation.YoureSoGifted" /></td>
        </tr>
      {/if}
      <tr>
        <th><T id="system.UserInformation.DateFormat" /></th>
        <td>
          {$profile.clientInfo?.user.dateFormat.fmt}
          from {$profile.clientInfo?.user.dateFormat.php}
        </td>
      </tr>
    {/if}
  </tbody>
</Table>
