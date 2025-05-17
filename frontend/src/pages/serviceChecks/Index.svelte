<script lang="ts" context="module">
  import { faBuildingCircleCheck } from '@fortawesome/sharp-duotone-light-svg-icons'
  export const page = {
    id: 'Services',
    icon: faBuildingCircleCheck,
    c1: 'steelblue',
    c2: 'coral',
    d1: 'wheat',
    d2: 'blue',
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
