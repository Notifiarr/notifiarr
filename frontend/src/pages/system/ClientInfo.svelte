<script lang="ts">
  import logo from '../../assets/notifiarr.svg?inline'
  import { Table } from '@sveltestrap/sveltestrap'
  import { profile } from '../../api/profile.svelte'
  import Header from './Header.svelte'
  import { age } from '../../includes/util'
  import { date, time } from '../../includes/Translate.svelte'
  import T from '../../includes/Translate.svelte'
  // make this dynamic?
  const uptime = $derived(age(Date.now() - new Date($profile.started).getTime()))
</script>

<!-- Client Section -->
<Header id="ClientInformation" {logo} />
<Table>
  <tbody>
    <tr>
      <th><T id="system.ClientInformation.Application" /></th>
      <td>{$profile.program}</td>
    </tr>
    <tr>
      <th><T id="system.ClientInformation.Version" /></th>
      <td>{$profile.version}-{$profile.revision}</td>
    </tr>
    <tr>
      <th><T id="system.ClientInformation.Started" /></th>
      <td>{date($profile.started)} {time($profile.started)}</td>
    </tr>
    <tr>
      <th><T id="system.ClientInformation.Uptime" /></th>
      <td>{uptime}</td>
    </tr>
  </tbody>
</Table>
