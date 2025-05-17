<script lang="ts" module>
  import { faTerminal } from '@fortawesome/sharp-duotone-solid-svg-icons'
  export const page = {
    id: 'Commands',
    i: faTerminal,
    c1: 'green',
    c2: 'darkgreen',
    d1: 'seagreen',
    d2: 'green',
  }
</script>

<script lang="ts">
  import { CardBody } from '@sveltestrap/sveltestrap'
  import { _ } from '../../includes/Translate.svelte'
  import { warning } from '../../includes/util'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'

  // Handle form submission
  function submit() {
    warning('Not implemented')
    // profile.writeConfig(c)
  }
</script>

<Header {page} />
<CardBody>
  <p>TODO</p>
</CardBody>
<Footer {submit} />
