<script lang="ts" module>
  export const page = { id: 'ApiDocs' }
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
  let viewer = $state<RapiDocEl | undefined>()
  let loadSeq = 0
  let fetchAbort: AbortController | undefined
  let uiInit: Promise<void> | undefined

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
    const root = el.shadowRoot
    if (!root || root.getElementById('nr-openapi-grow')) return
    const style = document.createElement('style')
    style.id = 'nr-openapi-grow'
    style.textContent =
      ':host{height:auto!important;max-height:none!important;overflow:hidden!important;' +
      'border-bottom-left-radius:var(--bs-card-inner-border-radius,0.375rem);' +
      'border-bottom-right-radius:var(--bs-card-inner-border-radius,0.375rem)}' +
      '.body,.main-content,.m-endpoint,summary{height:auto!important;max-height:none!important;' +
      'overflow:visible!important;overflow-anchor:none!important}'
    root.appendChild(style)
  }

  let unlockScroll: (() => void) | undefined

  const lockScroll = (el: RapiDocEl) => {
    unlockScroll?.()
    let pinned = window.scrollY
    let freeze = 0
    const html = document.documentElement
    const prevAnchor = html.style.overflowAnchor
    html.style.overflowAnchor = 'none'

    const inside = (node: Node | null) => {
      if (!node) return false
      const root = node.getRootNode()
      return root === el.shadowRoot || el.contains(node) || node === el
    }

    const restore = () => {
      if (window.scrollY !== pinned) window.scrollTo(0, pinned)
    }

    const onUserScroll = () => {
      if (freeze) {
        restore()
        return
      }
      pinned = window.scrollY
    }
    window.addEventListener('scroll', onUserScroll, { passive: true })

    const origIntoView = Element.prototype.scrollIntoView
    Element.prototype.scrollIntoView = function (
      this: Element,
      arg?: boolean | ScrollIntoViewOptions,
    ) {
      if (inside(this)) return
      origIntoView.call(this, arg as boolean)
    }

    const origFocus = HTMLElement.prototype.focus
    HTMLElement.prototype.focus = function (this: HTMLElement, opts?: FocusOptions) {
      origFocus.call(this, inside(this) ? { ...opts, preventScroll: true } : opts)
    }

    const hold = () => {
      freeze++
      pinned = freeze === 1 ? window.scrollY : pinned
      el.style.minHeight = `${el.offsetHeight}px`
      restore()
      const start = performance.now()
      const tick = () => {
        restore()
        if (performance.now() - start < 400) {
          requestAnimationFrame(tick)
          return
        }
        el.style.minHeight = ''
        freeze--
      }
      requestAnimationFrame(tick)
    }

    el.addEventListener('click', hold, true)
    el.addEventListener('toggle', hold, true)

    const ro = new ResizeObserver(restore)
    ro.observe(el)

    unlockScroll = () => {
      window.removeEventListener('scroll', onUserScroll)
      Element.prototype.scrollIntoView = origIntoView
      HTMLElement.prototype.focus = origFocus
      el.removeEventListener('click', hold, true)
      el.removeEventListener('toggle', hold, true)
      ro.disconnect()
      el.style.minHeight = ''
      html.style.overflowAnchor = prevAnchor
      unlockScroll = undefined
    }
  }

  const attachViewer = (node: HTMLElement) => {
    const el = node as RapiDocEl
    viewer = el
    $effect(() => {
      applyTheme(el)
    })
    return () => {
      unlockScroll?.()
      if (viewer === el) viewer = undefined
    }
  }

  const ensureUi = async () => {
    uiInit ??= import('rapidoc').then(() => undefined)
    try {
      await uiInit
    } catch (error) {
      uiInit = undefined
      throw error
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
      const spec = await fetchSpec(selected, signal)
      if (seq !== loadSeq) return
      await ensureUi()
      if (seq !== loadSeq) return
      if (viewer) {
        applyTheme(viewer)
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
  <rapi-doc
    {@attach attachViewer}
    class={['openapi-host', loadError || !ready ? 'd-none' : '']}
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
    show-method-in-nav-bar="as-colored-text"
  ></rapi-doc>
</div>

<style>
  :global(.card:has(.openapi-wrap)) {
    overflow: hidden;
  }

  .openapi-wrap {
    overflow: hidden;
    border-bottom-left-radius: var(--bs-card-inner-border-radius, 0.375rem);
    border-bottom-right-radius: var(--bs-card-inner-border-radius, 0.375rem);
  }

  .openapi-host {
    display: block;
    width: 100%;
    overflow: hidden;
    border-bottom-left-radius: inherit;
    border-bottom-right-radius: inherit;
  }
</style>
