<script lang="ts" module>
  export type WatchFiles = (WatchFile | null)[]
</script>

<script lang="ts">
  import Input from '../../includes/Input.svelte'
  import { Col, Row } from '@sveltestrap/sveltestrap'
  import { _ } from '../../includes/Translate.svelte'
  import type { WatchFile } from '../../api/notifiarrConfig'
  import type { App } from '../../includes/Instance.svelte'

  type Props = {
    form: WatchFile
    original: WatchFile
    app: App
    index: number
    validate?: (id: string, value: any) => string
  }

  let { form = $bindable(), original, app, index, validate }: Props = $props()
</script>

<div class="watcher">
  <!-- Top row, shows name and url or hostname/ip. -->
  <Row>
    <!-- Name is required for all integrations except Nvidia. -->
    <Col lg={6}>
      <Input
        id={app.id + '.path'}
        bind:value={form.path}
        original={original?.path}
        {validate} />
    </Col>
  </Row>
</div>
