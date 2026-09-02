<script lang="ts" module>
  export const page = { id: 'ApiDocs' }

  let uiInit: Promise<void> | undefined

  const ensureUi = async () => {
    uiInit ??= import('rapidoc').then(() => undefined)
    try {
      await uiInit
    } catch (error) {
      uiInit = undefined
      throw error
    }
  }
</script>

<script lang="ts">
  import { onDestroy, onMount, tick } from 'svelte'
  import T from '../../includes/Translate.svelte'
  import { theme } from '../../includes/theme.svelte'
  import { urlbase } from '../../api/fetch'
  import { CardBody, Input } from '@sveltestrap/sveltestrap'
  import Fa from '../../includes/Fa.svelte'
  import Warning from 'phosphor-svelte/lib/Warning'
  import CircleNotch from 'phosphor-svelte/lib/CircleNotch'
  import { profile } from '../../api/profile.svelte'
  import Header from '../../includes/Header.svelte'
  import { lockWindowScroll } from './apiDocsScroll'

  type SpecDoc = { id: string; file: string; path: string }

  interface RapiDocEl extends HTMLElement {
    loadSpec: (spec: Record<string, unknown> | string) => void
    theme: string
    bgColor: string
    textColor: string
  }

  const apiDocs: SpecDoc[] = [
    { id: 'api', file: 'api_openapi.json', path: 'api' },
    { id: 'ui', file: 'ui_openapi.json', path: 'ui' },
  ]

  let loadError = $state('')
  let doc = $state(apiDocs[0])
  let ready = $state(false)
  let showViewer = $state(false)
  let viewer = $state<RapiDocEl | undefined>()
  let loadSeq = 0
  let fetchAbort: AbortController | undefined

  const palette = $derived(
    theme.isDark
      ? { theme: 'dark', bg: '#212529', fg: '#dee2e6' }
      : { theme: 'light', bg: '#ffffff', fg: '#212529' },
  )

  const specVersion = () =>
    [$profile.version, $profile.revision].filter(Boolean).join('-') || '0.0.0'

  const stampSpec = (spec: Record<string, unknown>, selected: SpecDoc) => ({
    ...spec,
    info: {
      ...((spec.info as Record<string, unknown> | undefined) ?? {}),
      version: specVersion(),
    },
    servers: [{ url: $urlbase + selected.path }],
  })

  const fetchSpec = async (selected: SpecDoc, signal: AbortSignal) => {
    const res = await fetch($urlbase + selected.file, { signal })
    if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
    return stampSpec((await res.json()) as Record<string, unknown>, selected)
  }

  const applyTheme = (el: RapiDocEl) => {
    const { theme: name, bg, fg } = palette
    el.setAttribute('theme', name)
    el.setAttribute('bg-color', bg)
    el.setAttribute('text-color', fg)
    el.theme = name
    el.bgColor = bg
    el.textColor = fg
  }

  const growHost = (el: HTMLElement) => {
    el.style.setProperty('height', 'auto', 'important')
    el.style.setProperty('max-height', 'none', 'important')
    const inject = () => {
      const root = el.shadowRoot
      if (!root) return false
      if (root.getElementById('nr-openapi-grow')) return true
      const style = document.createElement('style')
      style.id = 'nr-openapi-grow'
      style.textContent =
        ':host{height:auto!important;max-height:none!important;min-height:0!important;' +
        'overflow:hidden!important;flex:none!important;' +
        'border-bottom-left-radius:var(--bs-card-inner-border-radius,0.375rem)!important;' +
        'border-bottom-right-radius:var(--bs-card-inner-border-radius,0.375rem)!important}' +
        '.body,.main-content,.m-endpoint,summary{height:auto!important;max-height:none!important;' +
        'overflow:visible!important;overflow-anchor:none!important}' +
        '.main-content{border-bottom-left-radius:inherit;border-bottom-right-radius:inherit}'
      root.appendChild(style)
      return true
    }
    if (inject()) return
    let frames = 0
    const retry = () => {
      if (inject() || ++frames > 60) return
      requestAnimationFrame(retry)
    }
    requestAnimationFrame(retry)
  }

  let unlockScroll: (() => void) | undefined

  const lockScroll = (el: RapiDocEl) => {
    unlockScroll?.()
    const unlock = lockWindowScroll(el)
    unlockScroll = () => {
      unlock()
      unlockScroll = undefined
    }
  }

  const attachViewer = (node: HTMLElement) => {
    const el = node as RapiDocEl
    viewer = el
    growHost(el)
    $effect(() => {
      applyTheme(el)
    })
    return () => {
      unlockScroll?.()
      if (viewer === el) viewer = undefined
    }
  }

  const loadSelected = async () => {
    const selected = doc
    const seq = ++loadSeq
    fetchAbort?.abort()
    fetchAbort = new AbortController()
    const { signal } = fetchAbort
    loadError = ''
    ready = false
    try {
      const specP = fetchSpec(selected, signal)
      const uiP = ensureUi()
      const settled = await Promise.allSettled([specP, uiP])
      if (seq !== loadSeq) return
      if (settled[0].status === 'rejected') throw settled[0].reason
      if (settled[1].status === 'rejected') throw settled[1].reason
      const spec = settled[0].value
      showViewer = true
      await tick()
      if (seq !== loadSeq) return
      if (viewer) {
        applyTheme(viewer)
        growHost(viewer)
        viewer.loadSpec(spec)
        growHost(viewer)
        lockScroll(viewer)
        applyTheme(viewer)
      }
      ready = true
    } catch (error) {
      if (signal.aborted || seq !== loadSeq) return
      loadError = error instanceof Error ? error.message : `${error}`
    }
  }

  onDestroy(() => {
    fetchAbort?.abort()
    unlockScroll?.()
  })

  onMount(async () => {
    await tick()
    await new Promise<void>(resolve => requestAnimationFrame(() => resolve()))
    await loadSelected()
  })
</script>

<Header {page} badge="v{$profile.version}">
  <p><T id="ApiDocs.Contrast" /></p>
  <Input type="select" bind:value={doc} onchange={loadSelected} class="mb-2">
    <option value={null} disabled><T id="ApiDocs.Choose" /></option>
    {#each apiDocs as ad (ad.id)}
      <option value={ad}>
        <T id={`ApiDocs.${ad.id}.title`} basePath={$urlbase + ad.path} />
      </option>
    {/each}
  </Input>
  <T id={`ApiDocs.${doc.id}.body`} />
  <ul class="mb-0 mt-2">
    <li><T id="ApiDocs.BasePath" basePath={$urlbase + doc.path} /></li>
  </ul>
</Header>

<div class="openapi-wrap">
  {#if loadError || !ready}
    <CardBody>
      <h5>
        {#if loadError}
          <Fa i={Warning} btn c1="red">
            <T id="phrases.ERROR" />
          </Fa><br />
          {loadError}
        {:else}
          <Fa i={CircleNotch} spin weight="bold" btn c1="orange">
            <T id="phrases.Loading" />
          </Fa>
        {/if}
      </h5>
    </CardBody>
  {/if}
  {#if showViewer}
    <rapi-doc
      {@attach attachViewer}
      class={['openapi-host', ready && !loadError ? 'is-ready' : 'is-pending']}
      style="height:auto!important;max-height:none!important;min-height:0!important"
      theme={palette.theme}
      bg-color={palette.bg}
      text-color={palette.fg}
      primary-color="#2fa582"
      render-style="view"
      layout="column"
      allow-try="false"
      allow-authentication="false"
      show-header="false"
      show-info="false"
      allow-spec-url-load="false"
      allow-spec-file-load="false"
      allow-spec-file-download="false"
      allow-server-selection="false"
      update-route="false"
      load-fonts="false"
      show-method-in-nav-bar="as-colored-text"></rapi-doc>
  {/if}
</div>

<style>
  /* Clip RapiDoc, not the card. overflow:hidden on .card squares the teal outline. */
  .openapi-wrap {
    position: relative;
    height: auto;
    min-height: 0;
    overflow: clip;
    border-bottom-left-radius: var(--bs-card-inner-border-radius, 0.375rem);
    border-bottom-right-radius: var(--bs-card-inner-border-radius, 0.375rem);
  }

  .openapi-host {
    display: block !important;
    width: 100%;
    height: auto !important;
    max-height: none !important;
    min-height: 0 !important;
    overflow: clip;
    border-bottom-left-radius: inherit;
    border-bottom-right-radius: inherit;
  }

  /* RapiDoc's :host { height:100% } must not take layout space until it is ready. */
  .openapi-host.is-pending {
    position: absolute !important;
    width: 0 !important;
    height: 0 !important;
    min-height: 0 !important;
    overflow: hidden !important;
    visibility: hidden !important;
    pointer-events: none !important;
  }
</style>
