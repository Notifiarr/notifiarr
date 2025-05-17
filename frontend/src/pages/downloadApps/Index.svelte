<script lang="ts" context="module">
  import { faDownload } from '@fortawesome/sharp-duotone-regular-svg-icons'
  export const page = {
    id: 'Downloaders',
    icon: faDownload,
    c1: 'brown',
    c2: 'lightsalmon',
    d1: 'coral',
    d2: 'lightpink',
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
<CardBody>TODO</CardBody>
<Footer {submit} />
