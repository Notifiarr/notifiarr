<script lang="ts" module>
  import { get } from 'svelte/store'

  let notification = $state('')
  let spin = $state(false)

  export function showMsg(msg: string) {
    notification = msg
  }

  export async function updateBackend(e?: Event) {
    e?.preventDefault()
    showMsg(`<span class="text-warning">${get(_)('phrases.UpdatingBackEnd')}</span>`)
    spin = true
    try {
      await profile.refresh()
      await delay(2345)
      showMsg('')
    } catch (err) {
      showMsg(
        `<span class="text-danger">
        ${get(_)('phrases.FailedToUpdateBackEnd', { values: { error: `${err}` } })}
        </span>`,
      )
    } finally {
      spin = false
    }
  }
</script>

<script lang="ts">
  import { theme } from '../includes/theme.svelte'
  import { profile } from '../api/profile.svelte'
  import { urlbase } from '../api/fetch'
  import { nav } from '../navigation/nav.svelte'
  import { faArrowsRepeat } from '@fortawesome/sharp-duotone-solid-svg-icons'
  import Fa from '../includes/Fa.svelte'
  import { Navbar, NavbarBrand, Row, Col, Card } from '@sveltestrap/sveltestrap'
  import logo from '../assets/notifiarr.svg?inline'
  import T, { _ } from '../includes/Translate.svelte'
  import { age, delay } from '../includes/util'
  import Reload from './Reload.svelte'
  import Shutdown from './Shutdown.svelte'
</script>

<!-- Top of the page. Logo and reload / shutdown buttons. -->
<Navbar theme={$theme} class="mb-0 pb-0">
  {#if $profile.loggedIn}
    <span style="position: absolute; right: 0;" class="fs-3">
      <!-- Reload Confirmation Link + Modal -->
      <Reload />
      <!-- Shutdown Confirmation Link + Modal -->
      <Shutdown />
    </span>
  {/if}
  <NavbarBrand href={$urlbase} onclick={e => nav.goto(e, '')} class="mb-0 pb-0">
    <h1 class="m-0 lh-1" style="font-size: 40px;">
      <img src={logo} height="45" alt="Logo" />
      <span class="title-notifiarr">Notifiarr Client</span>
    </h1>
  </NavbarBrand>
</Navbar>

<!-- Notification Center-->
<Row class="mt-0 mb-1 lh-1">
  <Col class="fs-6 fs-lighter ms-3 fst-italic">
    <Card color="transparent border-0" theme={$theme}>
      <span class="text-nowrap">
        {#if $profile?.loggedIn}
          <a href="#reload" onclick={updateBackend} class="text-decoration-none">
            <Fa i={faArrowsRepeat} c1="#3cd2a5" d1="green" class="me-1" {spin} />
            {#if notification}
              {@html notification}
            {:else if profile.updated}
              <T
                id="phrases.BackEndUpdated"
                timeDuration={age(profile.now - new Date($profile.updated).getTime())} />
            {/if}
          </a>
        {/if}
      </span>
    </Card>
  </Col>
</Row>
