<script lang="ts" module>
  export type Result = TestResult & { app: string; idx: number; id: number }
</script>

<script lang="ts">
  import type { TestResult } from '../../api/notifiarrConfig'
  type Props = { row: Result; expand: (idx: number) => void }
  const { row, expand }: Props = $props()

  const onclick = (e: MouseEvent) => {
    e.preventDefault()
    expand(row.idx)
  }
</script>

<a href="#expand" {onclick} class="ahref">{row.app} {row.id + 1}</a>

<style>
  .ahref {
    color: inherit !important;
    text-decoration: none;
  }
</style>
