<script lang="ts">
  import {
    faUserAlt,
    faUserSecret,
    faUserBountyHunter,
    faUserNinja,
    faUserCowboy,
  } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import { faUserHairMullet } from '@fortawesome/sharp-duotone-light-svg-icons'
  import { Table } from '@sveltestrap/sveltestrap'
  import { profile } from '../../api/profile.svelte'
  import Header from './Header.svelte'
  import T from '../../includes/Translate.svelte'
  import type { Props } from '../../includes/Fa.svelte'

  let icon: Props = { i: faUserAlt, c1: 'green', c2: 'lightgreen', d2: 'white' }

  if ($profile.clientInfo?.user?.subscriber) {
    icon.i = $profile.clientInfo?.user?.devAllowed ? faUserBountyHunter : faUserNinja
    icon.c1 = 'darkgoldenrod'
    icon.c2 = 'lightcoral'
    icon.d1 = 'goldenrod'
    icon.d2 = 'azure'
  } else if ($profile.clientInfo?.user?.patron) {
    icon.i = $profile.clientInfo?.user?.devAllowed ? faUserCowboy : faUserHairMullet
    icon.c1 = 'orange'
    icon.c2 = 'wheat'
  } else if ($profile.clientInfo?.user?.devAllowed) {
    icon.i = faUserSecret
    icon.c1 = 'purple'
    icon.c2 = 'coral'
  }
</script>

<!-- User Section -->
<Header id="UserInformation" {...icon} page="ClientInfo" />

<Table>
  <tbody>
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
  </tbody>
</Table>
