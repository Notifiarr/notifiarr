<script lang="ts">
  import { Card, CardBody, CardHeader, Table } from '@sveltestrap/sveltestrap'
  import T, { _ } from '../../includes/Translate.svelte'
  import Input from '../../includes/Input.svelte'
  import { slide } from 'svelte/transition'
  import { escapeHtml } from '../../includes/util'

  const testRegex = (): string => {
    try {
      const m = new RegExp(test).exec(pattern)
      if (!m) return `<b class="text-warning">${$_('RegexTester.NoMatch')}</b>`
      const highlightMatch = '<b class="text-info">' + escapeHtml(m[0]) + '</b>'
      const location = pattern.search(m[0])
      return (
        escapeHtml(pattern.slice(0, location)) +
        highlightMatch +
        escapeHtml(pattern.slice(location + m[0].length))
      )
    } catch (e) {
      return `<b class="text-danger">${escapeHtml(`${e}`)}</b>`
    }
  }

  /** Get the regex from the form. */
  export const regex = (): string => test

  let test = $state('')
  let pattern = $state('')
</script>

<h4><T id={'RegexTester.RegularExpressionTester'} /></h4>
<Input
  id={'RegexTester.testRegex'}
  type="textarea"
  bind:value={test}
  rows={1}
  original={null} />
<Input
  id={'RegexTester.testPattern'}
  type="textarea"
  bind:value={pattern}
  rows={1}
  original={null} />

{#if pattern && test}
  <div transition:slide>
    <Card color="info" outline>
      <CardHeader><b class="mb-0"><T id={'RegexTester.TestResult'} /></b></CardHeader>
      <CardBody>
        <Table size="sm" class="mb-0 pb-0" responsive>
          <tbody class="fit">
            <tr>
              <th><T id={'RegexTester.Test'} /></th>
              <td>{pattern}</td>
            </tr>
            <tr>
              <th><T id={'RegexTester.Expression'} /></th>
              <td><span class="text-info">{test}</span></td>
            </tr>
            <tr>
              <th><T id={'RegexTester.Result'} /></th>
              <td>{@html testRegex()}</td>
            </tr>
          </tbody>
        </Table>
      </CardBody>
    </Card>
  </div>
{/if}

<style>
  /* Make the left-side headers fit the max-content length. */
  tbody.fit th {
    white-space: nowrap;
    width: 1%;
    padding-right: 1rem;
  }

  tbody.fit td {
    word-break: break-all;
  }
</style>
