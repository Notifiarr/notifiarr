<script lang="ts" module>
  import { get } from 'svelte/store'
  import T, { _ } from '../../includes/Translate.svelte'

  export const validator = (id: string, value: any): string => {
    /* HTTP */
    if (id == 'url') {
      return value.match(/^http:\/\/../) || value.match(/^https:\/\/../)
        ? ''
        : get(_)('phrases.URLMustBeginWithHttp')
    }
    if (id === 'codes' && (!value || value.length === 0))
      return get(_)('ServiceChecks.http.codes.required')

    return ''
  }
</script>

<script lang="ts">
  import { Col, Label } from '@sveltestrap/sveltestrap'
  import type { ServiceConfig } from '../../api/notifiarrConfig'
  import type { ChildProps } from '../../includes/Instances.svelte'
  import Select from 'svelte-select'
  import { httpCodes } from './page.svelte'
  import Input from '../../includes/Input.svelte'
  import CheckedInput from '../../includes/CheckedInput.svelte'
  import { onMount } from 'svelte'
  import { deepEqual } from '../../includes/util'

  let {
    form = $bindable(),
    original,
    app,
    index,
    validate,
  }: ChildProps<ServiceConfig> = $props()

  const setData = (value: string, expect: string) => {
    return {
      url: value.split('|')[0] ?? '',
      headers: value.split('|').slice(1)?.join('\n') ?? '',
      codes: expect
        .split(',')
        .map(Number)
        .filter(c => !isNaN(c)),
      validSsl: expect.split(',').includes('SSL') ?? false,
    }
  }

  const originalHttp = $derived(setData(original.value, original.expect))
  let httpCheck = $state(setData(form.value, form.expect))
  export const reset = () => (httpCheck = setData(form.value, form.expect))

  let codeFeedback = $state<string | undefined>(undefined)

  const updateValue = (id: string, value: any) => {
    form.value = [
      httpCheck.url.trim(),
      httpCheck.headers
        .split('\n')
        .filter(h => h.trim())
        .join('|'),
    ]
      .filter(v => v.trim())
      .join('|')

    return id ? (validate?.(id, value) ?? '') : ''
  }

  const updateExpect = () => {
    form.expect =
      (httpCheck.codes?.join?.(',') ?? '') + (httpCheck.validSsl ? ',SSL' : '')
    codeFeedback = validate?.(app.id + '.http.codes', httpCheck.codes)
  }

  const merge = (index: number) => app.merge(index, form)

  // Clear the url validation if the page unmounts.
  onMount(() => () => {
    validate?.(app.id + '.url', 'https://this.is.valid')
    validate?.(app.id + '.http.codes', [200])
  })
</script>

<Col lg={6}>
  <CheckedInput
    id="url"
    app={{ ...app, merge, name: 'http' }}
    {index}
    bind:form={httpCheck}
    original={originalHttp}
    validate={(id: string, value: any) => (updateExpect(), updateValue(id, value))} />
</Col>

<Col lg={6}>
  <Input
    type="textarea"
    rows={Math.min(httpCheck.headers.split('\n').length ?? 1, 15)}
    id={app.id + '.http.headers'}
    bind:value={httpCheck.headers}
    original={originalHttp.headers}
    badge={$_('Endpoints.badge.header', {
      values: { count: httpCheck.headers.split('\n').filter(h => h.trim()).length ?? 0 },
    })}
    validate={updateValue} />
</Col>

<Col md={12}>
  <div class="http-group mb-3">
    <div class="http-check"><Label><T id={app.id + '.http.codes.label'} /></Label></div>
    <Select
      on:change={updateExpect}
      on:clear={updateExpect}
      on:input={updateExpect}
      class="form-control {httpCheck.codes?.length &&
      deepEqual(httpCheck.codes, originalHttp.codes)
        ? ''
        : 'changed ' + (httpCheck.codes?.length ? 'is-valid' : 'is-invalid')}"
      placeholder={$_(app.id + '.http.codes.label')}
      bind:justValue={httpCheck.codes}
      value={httpCheck.codes?.map?.(c => ({
        label: httpCodes.find(h => h.value === c)?.label,
        value: c,
      }))}
      multiple
      searchable
      clearable
      items={httpCodes} />
    <div class="text-danger">{codeFeedback}</div>
    <small class="text-muted"><T id={app.id + '.http.codes.description'} /></small>
  </div>
</Col>

<style>
  .http-check {
    font-family: Verdana, Geneva, Tahoma, sans-serif;
    font-weight: 550;
  }
</style>
