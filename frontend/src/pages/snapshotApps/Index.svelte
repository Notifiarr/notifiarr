<script lang="ts" context="module">
  import { faCameraRetro } from '@fortawesome/sharp-duotone-light-svg-icons'
  export const page = {
    id: 'SnapshotApps',
    icon: faCameraRetro,
    c1: 'burlywood',
    c2: 'darkgray',
    d1: 'burlywood',
    d2: 'silver',
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
