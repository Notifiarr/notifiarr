<script lang="ts" module>
  import { faSatellite } from '@fortawesome/sharp-duotone-light-svg-icons'
  export const page = {
    id: 'SiteTunnel',
    i: faSatellite,
    c1: 'blue',
    c2: 'lightblue',
    d1: 'steelblue',
    d2: 'brown',
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
