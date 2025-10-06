<script lang="ts" module>
  import { CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { profile } from './api/profile.svelte'
  import Header from './includes/Header.svelte'
  import T from './includes/Translate.svelte'
  import golift from './assets/golift.png?inline'
  import { faHouseChimneyHeart } from '@fortawesome/sharp-duotone-light-svg-icons'
  export const page = {
    id: 'Landing',
    i: faHouseChimneyHeart,
    c1: 'seagreen',
    c2: 'darkgreen',
    d1: 'lightblue',
    d2: 'silver',
  }
</script>

<Header {page} />

<CardBody>
  <h5 class="mt-0"><T id="phrases.Landing.Welcome" username={$profile?.username} /></h5>
  <p><T id="phrases.Landing.SelectASection" /></p>
  <ul class="mb-0">
    <li><go-to page="apidocs"><T id="navigation.titles.ApiDocs" /></go-to></li>
  </ul>
</CardBody>

<CardFooter class="clearfix">
  <img src={golift} alt="Go Lift Logo" class="float-start me-2" />
  <pre>{$profile.fortune}</pre>
</CardFooter>
