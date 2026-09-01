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

  let icon: Props = { i: Bug, c1: 'green', c2: 'lightgreen', d2: 'white' }

  if ($profile.clientInfo?.user?.subscriber) {
    icon.i = $profile.clientInfo?.user?.devAllowed ? Horse : Cat
    icon.c1 = 'darkgoldenrod'
    icon.c2 = 'lightcoral'
    icon.d1 = 'goldenrod'
    icon.d2 = 'azure'
  } else if ($profile.clientInfo?.user?.patron) {
    icon.i = $profile.clientInfo?.user?.devAllowed ? BugBeetle : Butterfly
    icon.c1 = 'orange'
    icon.c2 = 'wheat'
  } else if ($profile.clientInfo?.user?.devAllowed) {
    icon.i = Bird
    icon.c1 = 'purple'
    icon.c2 = 'coral'
  }
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
