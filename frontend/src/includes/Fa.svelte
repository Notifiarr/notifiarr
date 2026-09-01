<script lang="ts" module>
  import type { Component, Snippet } from 'svelte'
  import type { IconWeight } from 'phosphor-svelte'

  /** Phosphor Svelte component, or a PNG/SVG URL (Vite import). */
  export type IconSource = Component<any> | string
  export type Bullet = true | 'circle' | 'square' | 'triangle'
  export type Flip = boolean | 'h' | 'v' | 'both' | 'horizontal' | 'vertical'

  export interface Props {
    /** Phosphor component or image URL. Optional when `help` or `bullet` is set. */
    i?: IconSource
    /** Dark-mode glyph or image URL. */
    d?: IconSource
    c1?: string
    c2?: string
    d1?: string
    d2?: string
    scale?: string | number
    /** Phosphor weight. Default bold — fill when `bullet`, duotone only if you ask. */
    weight?: IconWeight
    spin?: boolean
    /**
     * Title-row mark: 1.4em box. PNG and glyph share it, so headers cannot
     * drift out of alignment.
     */
    logo?: boolean
    /** Toolbar / input-addon mark: 1.5em. Replaces ad-hoc `scale={1.5}` / `"1.5x"`. */
    btn?: boolean
    /** Filled list disc. `true`/`circle`, or `square`/`triangle`. */
    bullet?: Bullet
    /** Circled help ? — gray ring, orange mark. */
    help?: boolean
    /** Mirror the glyph. `true`/`h` is horizontal (ICMP paddle). */
    flip?: Flip
    class?: string
    style?: string
    id?: string
    href?: string
    onclick?: () => void
    /** In-app navigation: wrap the mark in `<go-to>`. */
    page?: string
    /**
     * Label beside the mark. Locks icon+text into one centered row
     * (`inline-flex`, `0.4em` gap) so callers cannot forget alignment.
     */
    children?: Snippet
  }
</script>

<script lang="ts">
  import { theme } from './theme.svelte'
  import Circle from 'phosphor-svelte/lib/Circle'
  import Square from 'phosphor-svelte/lib/Square'
  import Triangle from 'phosphor-svelte/lib/Triangle'
  import Question from 'phosphor-svelte/lib/Question'
  import QuestionMark from 'phosphor-svelte/lib/QuestionMark'
  import CheckCircle from 'phosphor-svelte/lib/CheckCircle'
  import Check from 'phosphor-svelte/lib/Check'
  import XCircle from 'phosphor-svelte/lib/XCircle'
  import X from 'phosphor-svelte/lib/X'

  const bullets = { circle: Circle, square: Square, triangle: Triangle } as const

  const {
    d,
    c1,
    c2,
    d1,
    d2,
    i,
    spin = false,
    weight,
    scale,
    logo = false,
    btn = false,
    help = false,
    bullet,
    flip,
    class: className = '',
    style = '',
    id,
    href,
    onclick,
    page,
    children,
  }: Props = $props()

  const isDisc = $derived(bullet === true || bullet === 'circle')
  const c1r = $derived(c1 ?? (help ? 'gray' : isDisc ? 'royalblue' : undefined))
  const d1r = $derived(
    d1 ?? (help && c1 == null ? 'gainsboro' : isDisc && c1 == null ? 'orange' : c1r),
  )
  const c2r = $derived(c2 ?? (help ? 'orange' : undefined))
  const d2r = $derived(d2 ?? c2r)
  const primaryColor = $derived(theme.isDark ? d1r : c1r)
  const secondaryColor = $derived(theme.isDark ? d2r : c2r)
  const resolvedWeight = $derived(weight ?? (bullet ? 'fill' : 'bold'))
  const source = $derived.by(() => {
    const raw = theme.isDark && d != null ? d : i
    if (typeof raw === 'string') return raw
    if (raw) return raw
    if (help) return Question
    if (bullet) return bullets[bullet === true ? 'circle' : bullet]
    return Question
  })
  const src = $derived(typeof source === 'string' ? source : undefined)
  const Ph = $derived(typeof source === 'string' ? undefined : source)
  // Phosphor circled glyphs are one path, so duotone cannot color the ring
  // separately. When c1+c2 are set, draw Circle (c1) + inner mark (c2).
  const Inner = $derived(
    Ph === Question
      ? QuestionMark
      : Ph === CheckCircle
        ? Check
        : Ph === XCircle
          ? X
          : undefined,
  )
  const ringed = $derived(!!Ph && !!Inner && !!primaryColor && !!secondaryColor)
  const size = $derived.by(() => {
    if (scale != null && scale !== '') {
      const n = typeof scale === 'number' ? scale : parseFloat(String(scale))
      return Number.isFinite(n) ? `${n}em` : '1em'
    }
    if (bullet) return '0.55em'
    if (logo) return '1.4em'
    if (btn) return '1.5em'
    return '1em'
  })
  const flipH = $derived(
    flip === true || flip === 'h' || flip === 'horizontal' || flip === 'both',
  )
  const flipV = $derived(flip === 'v' || flip === 'vertical' || flip === 'both')
  const handleClick = (e: Event) => {
    if (onclick) {
      e.preventDefault()
      onclick()
      return
    }
    if (!href || href === '#anotherPage') e.preventDefault()
  }
</script>

{#snippet glyph()}
  <span
    class={[
      'nr-icon',
      !children && className,
      {
        spin,
        ringed,
        img: !!src,
        'flip-h': flipH,
        'flip-v': flipV,
        'has-c2': resolvedWeight === 'duotone' && !!secondaryColor,
        'has-c1': !!primaryColor,
      },
    ]}
    id,
    style="--icon-c1: {primaryColor ??
      'currentColor'}; --icon-c2: {secondaryColor ??
      'currentColor'}; font-size: {size}; {style}">
    {#if src}
      <img {src} alt="" />
    {:else if ringed && Inner}
      <Circle weight={resolvedWeight} color={primaryColor} size="1em" />
      <Inner weight={resolvedWeight} color={secondaryColor} size="1em" />
    {:else if Ph}
      <Ph weight={resolvedWeight} color={primaryColor ?? 'currentColor'} size="1em" />
    {/if}
  </span>
{/snippet}

{#snippet mark()}
  {#if page}
    <go-to class="nr-go" {page}>{@render glyph()}</go-to>
  {:else if onclick || href}
    <a href={href ?? '#anotherPage'} onclick={handleClick}>
      {@render glyph()}
    </a>
  {:else}
    {@render glyph()}
  {/if}
{/snippet}

{#if children}
  <span class={['nr-mark', className]}>
    {@render mark()}
    {@render children()}
  </span>
{:else}
  {@render mark()}
{/if}

<style>
  .nr-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    vertical-align: middle;
    line-height: 0;
    width: 1em;
    height: 1em;
    flex-shrink: 0;
    margin: 0;
  }
  .nr-icon :global(svg) {
    display: block;
    width: 1em;
    height: 1em;
    flex-shrink: 0;
  }
  /* PNG/SVG logos are wide, not square — same em-box height as glyphs. */
  .nr-icon.img {
    width: auto;
    max-width: 1.6em;
  }
  .nr-icon.img img {
    display: block;
    height: 1em;
    width: auto;
    max-height: 1em;
    max-width: 1.6em;
    object-fit: contain;
  }
  .nr-icon.ringed {
    position: relative;
  }
  .nr-icon.ringed :global(svg:last-of-type) {
    position: absolute;
    width: 0.52em;
    height: 0.52em;
    inset: 0;
    margin: auto;
  }
  .nr-icon.flip-h {
    transform: scaleX(-1);
  }
  .nr-icon.flip-v {
    transform: scaleY(-1);
  }
  .nr-icon.flip-h.flip-v {
    transform: scale(-1);
  }
  /* Duotone only: empty <rect>, then a 0.2-opacity fill, then the glyph.
     Leave that opacity alone — forcing 1 paints a solid blob. */
  .nr-icon.has-c2 :global(svg path:nth-of-type(1)) {
    fill: var(--icon-c2);
  }
  .nr-icon.has-c1.has-c2 :global(svg path:nth-of-type(2)) {
    fill: var(--icon-c1);
  }
  .nr-icon.spin :global(svg) {
    animation: nr-spin 1s linear infinite;
  }
  @keyframes nr-spin {
    to {
      transform: rotate(360deg);
    }
  }
  :global(go-to.nr-go) {
    display: inline-flex;
    align-items: center;
    line-height: 0;
    flex-shrink: 0;
  }
  /* Icon + label: one locked row. Callers must not re-flex this. */
  .nr-mark {
    display: inline-flex;
    align-items: center;
    gap: 0.4em;
    line-height: 1;
    vertical-align: middle;
    min-width: 0;
    max-width: 100%;
  }
</style>
