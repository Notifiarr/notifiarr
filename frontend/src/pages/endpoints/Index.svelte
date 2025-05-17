<script lang="ts" module>
  import { faWebhook } from '@fortawesome/sharp-duotone-solid-svg-icons'
  export const page = {
    id: 'Endpoints',
    i: faWebhook,
    c1: 'sandybrown',
    c2: 'saddlebrown',
    d1: 'lightcoral',
    d2: 'tan',
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
