<script lang="ts" module>
  import { CardBody, CardFooter } from '@sveltestrap/sveltestrap'
  import { profile } from './api/profile.svelte'
  import Header from './includes/Header.svelte'
  import T from './includes/Translate.svelte'
  import golift from './assets/golift.png?inline'
  export const page = { id: 'Landing' }
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
