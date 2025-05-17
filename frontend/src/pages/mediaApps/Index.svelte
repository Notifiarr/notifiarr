<script lang="ts" context="module">
  import { faClapperboardPlay } from '@fortawesome/sharp-duotone-light-svg-icons'
  export const page = {
    id: 'MediaApps',
    icon: faClapperboardPlay,
    c1: 'indigo',
    c2: 'blue',
    d1: 'lightseagreen',
    d2: 'antiquewhite',
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
