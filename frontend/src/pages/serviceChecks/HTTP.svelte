<script lang="ts" module>
  import { get } from 'svelte/store'
  import T, { _ } from '../../includes/Translate.svelte'

  export const validator = (id: string, value: any): string => {
    /* HTTP */
    if (
      id == 'url' &&
      (!value || (!value.match(/^http:\/\/../) && !value.match(/^https:\/\/../)))
    )
      return get(_)('phrases.URLMustBeginWithHttp')
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

  const originalHttp = $derived(setData(original?.value ?? '', original?.expect ?? ''))
  let httpCheck = $state(setData(form?.value ?? '', form?.expect ?? ''))
  export const reset = () => (httpCheck = setData(form?.value ?? '', form?.expect ?? ''))

  let codeFeedback = $state<string | undefined>(undefined)

  const updateValue = (id: string, value: any): string => {
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

  const codesFromSelect = (value: unknown): number[] => {
    if (value == null) return []
    const list = Array.isArray(value) ? value : [value]
    return list.map(Number).filter(c => !Number.isNaN(c))
  }

  const updateExpect = (codes = httpCheck.codes) => {
    const list = Array.isArray(codes) ? codes : []
    httpCheck.codes = list
    codeFeedback = validate?.(app.id + '.http.codes', list)
  }

  const onCodes = (value: unknown) => updateExpect(codesFromSelect(value))

  const merge = (index: number) => app.merge(index, form)

  // Clear the url validation if the page unmounts.
  onMount(() => () => {
    validate?.(app.id + '.url', 'https://this.is.valid')
    validate?.(app.id + '.http.codes', [200])
  })

  updateExpect()

  // CheckedInput toggles validSsl on https URLs. Sync form.expect from codes + SSL
  // without rewriting codes (that would fight bind:value).
  $effect(() => {
    const codes = httpCheck.codes ?? []
    form.expect = codes.join(',') + (httpCheck.validSsl ? ',SSL' : '')
  })
</script>

<Col lg={6}>
  <CheckedInput
    id="url"
    envVar={`${app.envPrefix}_${index}_VALUE`}
    app={{ ...app, merge, name: 'http' }}
    {index}
    bind:form={httpCheck}
    original={originalHttp}
    validate={updateValue} />
</Col>

<Col lg={6}>
  <Input
    type="textarea"
    rows={Math.min(httpCheck.headers.split('\n').length ?? 1, 15)}
    id={app.id + '.http.headers'}
    envVar={`${app.envPrefix}_${index}_VALUE`}
    bind:value={httpCheck.headers}
    original={originalHttp.headers}
    validate={updateValue}
    badge={$_('Endpoints.badge.header', {
      values: { count: httpCheck.headers.split('\n').filter(h => h.trim()).length ?? 0 },
    })} />
</Col>

<Col md={12}>
  <div class="http-group mb-3">
    <div class="http-check"><Label><T id={app.id + '.http.codes.label'} /></Label></div>
    <Select
      class="form-control {httpCheck.codes?.length &&
      deepEqual(httpCheck.codes, originalHttp.codes)
        ? ''
        : 'changed ' + (httpCheck.codes?.length ? 'is-valid' : 'is-invalid')}"
      placeholder={$_(app.id + '.http.codes.label')}
      valueMode="id"
      bind:value={httpCheck.codes}
      oninput={onCodes}
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

  .http-group :global(.changed) {
    background-color: rgba(205, 92, 92, 0.322) !important;
  }
</style>
