<script lang="ts" module>
  import { faFileWaveform } from '@fortawesome/sharp-duotone-light-svg-icons'
  export const page = {
    id: 'FileWatcher',
    i: faFileWaveform,
    d1: 'thistle',
    d2: 'blue',
    c1: 'sienna',
    c2: 'moccasin',
  }
</script>

<script lang="ts">
  import { CardBody } from '@sveltestrap/sveltestrap'
  import { _ } from '../../includes/Translate.svelte'
  import Footer from '../../includes/Footer.svelte'
  import Header from '../../includes/Header.svelte'

  // Handle form submission
  function submit(e: Event) {
    e.preventDefault()
    // profile.writeConfig(c)
  }
</script>

<Header {page} />
<CardBody>
  <p>TODO</p>
</CardBody>
<Footer {submit} />
