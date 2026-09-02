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

  const selectId = (v: unknown): number | undefined => {
    if (v && typeof v === 'object' && 'value' in v)
      return selectId((v as { value: unknown }).value)
    if (typeof v === 'number' && Number.isInteger(v)) return v
    if (typeof v === 'string' && v.trim() !== '') {
      const n = Number(v)
      return Number.isInteger(n) ? n : undefined
    }
    return undefined
  }

  const selectIds = (value: unknown): number[] => {
    if (value == null) return []
    const list = Array.isArray(value) ? value : [value]
    return list.map(selectId).filter((n): n is number => n !== undefined)
  }

  const setData = (value: string, expect: string) => {
    const tokens = expect
      .split(',')
      .map(s => s.trim())
      .filter(Boolean)
    return {
      url: value.split('|')[0] ?? '',
      headers: value.split('|').slice(1)?.join('\n') ?? '',
      codes: selectIds(tokens.filter(s => s !== 'SSL')),
      validSsl: tokens.includes('SSL'),
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

  const updateExpect = (codes = httpCheck.codes, validSsl = httpCheck.validSsl) => {
    const list = selectIds(codes)
    form.expect = [...list, ...(validSsl ? ['SSL'] : [])].join(',')
    codeFeedback = validate?.(app.id + '.http.codes', list)
  }

  const onCodes = (value: unknown) => updateExpect(selectIds(value), httpCheck.validSsl)

  const merge = (index: number) => app.merge(index, form)

  // Clear the url validation if the page unmounts.
  onMount(() => () => {
    validate?.(app.id + '.url', 'https://this.is.valid')
    validate?.(app.id + '.http.codes', [200])
  })

  // CheckedInput toggles validSsl on https URLs. Do not assign httpCheck.codes here
  // (that fights bind:value); only rewrite the persisted expect string.
  $effect(() => {
    updateExpect(httpCheck.codes, httpCheck.validSsl)
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
