<script lang="ts">
  import Input from '../../includes/Input.svelte'
  import { Col, Row } from '@sveltestrap/sveltestrap'
  import { _ } from '../../includes/Translate.svelte'
  import type { WatchFile } from '../../api/notifiarrConfig'
  import type { ChildProps } from '../../includes/Instances.svelte'

  let {
    form = $bindable(),
    original,
    app,
    index, // unused but matches our ChildProps interface.
    validate,
  }: ChildProps<WatchFile> = $props()
</script>

<div class="watcher">
  <Row>
    <Col md={9}>
      <Input
        id={app.id + '.path'}
        bind:value={form.path}
        original={original?.path}
        {validate} />
    </Col>
    <Col md={3}>
      <Input
        type="select"
        id={app.id + '.disabled'}
        bind:value={form.disabled}
        original={original?.disabled}
        {validate} />
    </Col>
    <Col lg={12}>
      <Input
        type="textarea"
        id={app.id + '.regex'}
        bind:value={form.regex}
        original={original?.regex}
        {validate} />
    </Col>
    <Col lg={12}>
      <Input
        type="textarea"
        id={app.id + '.skip'}
        bind:value={form.skip}
        original={original?.skip}
        {validate} />
    </Col>
    <Col sm={6}>
      <Input
        type="select"
        id={app.id + '.poll'}
        bind:value={form.poll}
        original={original?.poll}
        {validate} />
    </Col>
    <Col sm={6}>
      <Input
        type="select"
        id={app.id + '.pipe'}
        bind:value={form.pipe}
        original={original?.pipe}
        {validate} />
    </Col>
    <Col sm={6}>
      <Input
        type="select"
        id={app.id + '.mustExist'}
        bind:value={form.mustExist}
        original={original?.mustExist}
        {validate} />
    </Col>
    <Col sm={6}>
      <Input
        type="select"
        id={app.id + '.logMatch'}
        bind:value={form.logMatch}
        original={original?.logMatch}
        {validate} />
    </Col>
  </Row>
</div>
