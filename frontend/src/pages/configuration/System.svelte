<script lang="ts">
  import { Row, Col } from '@sveltestrap/sveltestrap'
  import Input from '../../includes/Input.svelte'
  import { _ } from '../../includes/Translate.svelte'
  import { profile } from '../../api/profile.svelte'
  import type { Config } from '../../api/notifiarrConfig'

  type Props = { config: Config; original: Config }
  const { config = $bindable(), original }: Props = $props()
</script>

<!-- System Section -->
<h4>{$_('config.titles.System')}</h4>
<Row>
  <Col md={4}>
    <Input
      id="config.serial"
      envVar="SERIAL"
      type="select"
      bind:value={config.serial}
      original={original.serial} />
  </Col>
  {#if $profile.isWindows}
    <Col md={$profile.clientInfo?.user.devAllowed ? 4 : 8}>
      <Input
        id="config.autoUpdate"
        envVar="AUTO_UPDATE"
        type="select"
        bind:value={config.autoUpdate}
        original={original.autoUpdate}
        options={[
          { value: 'off', name: $_('words.select-option.Disabled') },
          { value: 'daily', name: $_('words.select-option.Daily') },
          { value: '12h', name: $_('phrases.Every12Hours') },
          { value: '6h', name: $_('phrases.Every6Hours') },
          { value: '3h', name: $_('phrases.Every3Hours') },
        ]} />
    </Col>
    {#if $profile.clientInfo?.user.devAllowed}
      <Col md={4}>
        <Input
          id="config.unstableCh"
          envVar="UNSTABLE_CH"
          type="select"
          bind:value={config.unstableCh}
          original={original.unstableCh} />
      </Col>
    {/if}
  {:else}
    <Col md={4}>
      <Input
        envVar="FILE_MODE"
        type="select"
        id="config.fileMode"
        bind:value={config.fileMode}
        original={original.fileMode}
        options={[
          { value: '0600', name: '0600 -rw-------' },
          { value: '0640', name: '0640 -rw-r-----' },
          { value: '0644', name: '0644 -rw-r--r--' },
          { value: '0604', name: '0604 -rw----r--' },
          { value: '0660', name: '0660 -rw-rw----' },
          { value: '0664', name: '0664 -rw-rw-r--' },
        ]} />
    </Col>
    <Col md={4}>
      <Input
        id="config.apt"
        envVar="APT"
        type="select"
        bind:value={config.apt}
        original={original.apt} />
    </Col>
  {/if}
</Row>
