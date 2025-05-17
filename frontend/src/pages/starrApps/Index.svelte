<script lang="ts" context="module">
  import { faStars } from '@fortawesome/sharp-duotone-regular-svg-icons'
  export const page = {
    id: 'StarrApps',
    icon: faStars,
    c1: 'darkgoldenrod',
    c2: 'gold',
    d1: 'gold',
    d2: 'darkgoldenrod',
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
