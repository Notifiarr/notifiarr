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
        envVar={`${app.envPrefix}_${index}_PATH`}
        bind:value={form.path}
        original={original?.path}
        {validate} />
    </Col>
    <Col md={3}>
      <Input
        type="select"
        id={app.id + '.disabled'}
        envVar={`${app.envPrefix}_${index}_DISABLED`}
        bind:value={form.disabled}
        original={original?.disabled}
        {validate} />
    </Col>
    <Col lg={12}>
      <Input
        type="textarea"
        id={app.id + '.regex'}
        envVar={`${app.envPrefix}_${index}_REGEX`}
        bind:value={form.regex}
        original={original?.regex}
        {validate} />
    </Col>
    <Col lg={12}>
      <Input
        type="textarea"
        id={app.id + '.skip'}
        envVar={`${app.envPrefix}_${index}_SKIP`}
        bind:value={form.skip}
        original={original?.skip}
        {validate} />
    </Col>
    <Col sm={6}>
      <Input
        type="select"
        id={app.id + '.poll'}
        envVar={`${app.envPrefix}_${index}_POLL`}
        bind:value={form.poll}
        original={original?.poll}
        {validate} />
    </Col>
    <Col sm={6}>
      <Input
        type="select"
        id={app.id + '.pipe'}
        envVar={`${app.envPrefix}_${index}_PIPE`}
        bind:value={form.pipe}
        original={original?.pipe}
        {validate} />
    </Col>
    <Col sm={6}>
      <Input
        type="select"
        id={app.id + '.mustExist'}
        envVar={`${app.envPrefix}_${index}_MUST_EXIST`}
        bind:value={form.mustExist}
        original={original?.mustExist}
        {validate} />
    </Col>
    <Col sm={6}>
      <Input
        type="select"
        id={app.id + '.logMatch'}
        envVar={`${app.envPrefix}_${index}_LOG_MATCH`}
        bind:value={form.logMatch}
        original={original?.logMatch}
        {validate} />
    </Col>
  </Row>
</div>
