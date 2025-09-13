<script lang="ts">
  import { Card, CardHeader, Popover, Table } from '@sveltestrap/sveltestrap'
  import { age } from '../../includes/util'
  import { color, getLogo, title } from './data'
  import type { TautulliConfig, Info, TautulliUser } from '../../api/notifiarrConfig'
  import Modal from './Modal.svelte'
  import T, { _ } from '../../includes/Translate.svelte'
  import { profile } from '../../api/profile.svelte'

  type Props = {
    index?: number
    config: TautulliConfig
    status?: Info
    statusAge?: number | Date
    users?: TautulliUser[]
    usersAge?: number | Date
  }

  const { index = 0, config, status, statusAge, users, usersAge }: Props = $props()
  const app = 'tautulli'
  let statusModal: Modal | null = $state(null)
  let usersModal: Modal | null = $state(null)
  let usersJson: Modal | null = $state(null)
</script>

<Card outline color={color(app)}>
  <CardHeader>
    <h5 class="m-0">
      <img
        src={getLogo(app)}
        alt="{app} Logo"
        class="float-start me-2"
        height="32"
        width="32" />
      {title(app)}
      {index + 1}
    </h5>
  </CardHeader>
  <Table class="rounded-bottom mb-0" size="sm">
    <tbody class="table-body">
      <tr>
        <td class="text-break">{config.name}</td>
        <td class="text-break"><a href={config.url}>{config.url}</a></td>
      </tr>
      {#if status}
        <tr>
          <td class="text-nowrap">
            <Modal pageId="{app}Status" data={status} bind:this={statusModal} />
            <a href="#{app}Status" onclick={statusModal?.toggle}>
              <T id="Integrations.titles.StatusAge" /></a>
          </td>
          <td class="text-break">
            {age(profile.now - new Date(statusAge ?? 0).getTime())}</td>
        </tr>
        <tr>
          <td class="text-nowrap"><T id="Integrations.titles.Version" /></td>
          <td class="text-break">{status?.tautulli_version}</td>
        </tr>
        <tr>
          <td class="text-nowrap"><T id="Integrations.titles.Branch" /></td>
          <td class="text-break">{status?.tautulli_branch}</td>
        </tr>
        <tr>
          <td class="text-nowrap"><T id="system.OperatingSystem.Platform" /></td>
          <td class="text-break">{status?.tautulli_platform}</td>
        </tr>
      {/if}

      {#if status && users}<tr><td colspan="2"></td></tr>{/if}

      {#if users && users.length > 0}
        <tr>
          <td class="text-nowrap"><T id="Integrations.mediaTitles.Users" /></td>
          <td>
            <a href="#{app}Users" onclick={usersModal?.toggle}>{users.length}</a>
            <Modal pageId="{app}Users" bind:this={usersModal}>
              <Table responsive striped>
                <thead>
                  <tr>
                    <th class="text-nowrap">
                      <T id="Integrations.mediaTitles.PlexUsername" /></th>
                    <th class="text-nowrap">
                      <T id="Integrations.mediaTitles.CustomName" /></th>
                    <th><T id="Integrations.mediaTitles.Email" /></th>
                  </tr>
                </thead>
                <tbody>
                  {#each users as user}
                    <tr>
                      <td class="text-nowrap">
                        <img
                          referrerpolicy="no-referrer"
                          class="float-start me-2"
                          width="25"
                          height="25"
                          src={user.thumb}
                          alt={user.username} />{user.username}</td>
                      <td>{user.friendly_name}</td>
                      <td class="text-break">
                        <a href="mailto:{user.email}">{user.email}</a></td>
                    </tr>
                  {/each}
                </tbody>
              </Table>
            </Modal>
          </td>
        </tr>
        <tr>
          <td class="text-nowrap">
            <a href="#Integrations.{app}UsersJson" onclick={usersJson?.toggle}>
              <T id="Integrations.titles.CacheAge" />
            </a>
            <Modal pageId="{app}UsersJson" data={users} bind:this={usersJson} />
          </td>
          <td class="text-break">
            {age(profile.now - new Date(usersAge ?? 0).getTime())}</td>
        </tr>
      {/if}
    </tbody>
  </Table>
</Card>

<style>
  .table-body :global(tr:last-of-type td) {
    border-bottom: 0 !important;
  }
</style>
