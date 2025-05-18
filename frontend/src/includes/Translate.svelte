<!-- this component allows a smoother syntax for html translations with values or href links. -->

<script module lang="ts">
  import { derived } from 'svelte/store'
  import { _, date as dt, isLoading, locale } from 'svelte-i18n'
  export { _ } // pass it through
  export const isReady = derived(locale, $locale => typeof $locale === 'string')

  let formatDate: (date: Date | number) => string // type is DateFormatter, but not sure how to import it.
  dt.subscribe(val => (formatDate = val))

  export function date(date: string | Date | number): string {
    if (typeof date !== 'string') return formatDate(date)
    return formatDate(new Date(0).setUTCMilliseconds(Date.parse(date)))
  }
</script>

<script lang="ts">
  interface Props {
    id: string
    [key: string]: any
  }

  let { ...props }: Props = $props()
</script>

{#if $isReady === true && $isLoading === false}
  {@html $_(props.id, { values: { ...props } })}
{/if}
