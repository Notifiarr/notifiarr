<script lang="ts">
  import Header from '../../includes/Header.svelte'
  import { faSplotch } from '@fortawesome/sharp-duotone-regular-svg-icons'
  import Fa from '../../includes/Fa.svelte'
  import T, { _ } from '../../includes/Translate.svelte'
  import { Card, Row } from '@sveltestrap/sveltestrap'
  import { slide } from 'svelte/transition'
  import { profile } from '../../api/profile.svelte'
  import { page } from './Index.svelte'

  type Props = { addit: (e: Event) => void }
  let { addit }: Props = $props()

  let showMore = $state(false)
  const toggleMore = (e: Event) => (e.preventDefault(), (showMore = !showMore))
  // goes into a translation.
  const upstreamIp = $derived(
    '<span class="text-danger">' + $profile?.upstreamIp + '</span>',
  )
</script>

<Header {page}>
  <!-- Bullet Points-->
  <ul class="mb-2 list-unstyled">
    <li>
      <Fa i={faSplotch} c1="gray" d1="gainsboro" c2="orange" />
      {@html $_('profile.phrase.PasswordOnlyUsed')}
    </li>
    <li>
      <Fa i={faSplotch} c1="gray" d1="gainsboro" c2="orange" />
      {@html $_('profile.phrase.HeaderOnlyUsed')}
    </li>
    <li>
      <Fa i={faSplotch} c1="gray" d1="gainsboro" c2="orange" />
      {@html $_('profile.phrase.AuthTypeChange')}
    </li>
    <li>
      <Fa i={faSplotch} c1="gray" d1="gainsboro" c2="orange" />
      {@html $_('profile.phrase.SeeWikiForAuthProxyHelp')}
    </li>
    <!-- Missing/Add upstream section -->
    {#if !$profile?.upstreamAllowed}
      <li>
        <Fa i={faSplotch} c1="maroon" c2="salmon" d1="salmon" d2="maroon" />
        <b><T id="profile.phrase.ProxyAuthDisabled" {upstreamIp} /></b>
        <a href="#addit" onclick={addit}>
          {$_('profile.phrase.Addit')}
        </a>
      </li>
    {/if}

    <!-- Show more toggle -->
    <li>
      <Fa i={faSplotch} c1="darkblue" d1="lightblue" c2="royalblue" />
      <a href="#toggle-more" onclick={toggleMore}>
        {showMore ? $_('profile.phrase.ShowMeLess') : $_('profile.phrase.ShowMeMore')}
      </a>
    </li>
  </ul>

  <!-- Show more section -->
  <Row>
    {#if showMore}
      <div transition:slide={{ duration: 900 }}>
        <Card body color="warning-subtle" class="my-1 pb-0">
          <p>{@html $_('profile.phrase.ShowMoreDesc')}</p>
          <h5>{$_('profile.authType.options.password')}</h5>
          <p>{@html $_('profile.authType.option_desc.password')}</p>
          <h5>{$_('profile.authType.options.header')}</h5>
          <p>{@html $_('profile.authType.option_desc.header')}</p>
          <h5>{$_('profile.authType.options.noauth')}</h5>
          <p>{@html $_('profile.authType.option_desc.noauth')}</p>
        </Card>
      </div>
    {/if}
  </Row>
</Header>
