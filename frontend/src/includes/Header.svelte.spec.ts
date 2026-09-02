import { render, screen, waitFor } from '@testing-library/svelte'
import { addMessages, locale } from 'svelte-i18n'
import { createRawSnippet } from 'svelte'
import { afterEach, describe, expect, it } from 'vitest'
import Header from './Header.svelte'

const ZZ_DESC = 'ZZ-BESCHREIBUNG'
const title = () => document.querySelector('h2.page-title')?.textContent ?? ''
const body = () => document.body.textContent ?? ''

describe('Header description', () => {
  afterEach(async () => {
    await locale.set('en')
  })

  it('renders the en pageDescription when no description prop is passed', async () => {
    render(Header, { props: { page: { id: 'Configuration' } } })
    await waitFor(() =>
      expect(
        screen.getByText(/Configure your Notifiarr client settings here\./),
      ).toBeTruthy(),
    )
  })

  it('updates the description when the locale changes', async () => {
    render(Header, { props: { page: { id: 'Configuration' } } })
    await waitFor(() => expect(title()).toContain('Configuration'))
    await addMessages('zz', {
      navigation: {
        titles: { Configuration: 'ZZ-TITEL' },
        pageDescription: { Configuration: ZZ_DESC },
      },
    })
    await locale.set('zz')
    await waitFor(() => expect(screen.getByText(ZZ_DESC)).toBeTruthy(), { timeout: 2000 })
    expect(title()).toContain('ZZ-TITEL')
  })

  it('renders nothing extra for a page with no pageDescription key', async () => {
    render(Header, { props: { page: { id: 'ApiDocs' } } })
    await waitFor(() => expect(title()).toContain('Client API Documentation'))
    expect(body()).not.toContain('navigation.pageDescription')
  })

  it('honors an explicit string description', async () => {
    render(Header, {
      props: { page: { id: 'Configuration' }, description: 'CUSTOM-STRING' },
    })
    await waitFor(() => expect(screen.getByText('CUSTOM-STRING')).toBeTruthy())
    expect(body()).not.toContain('Configure your Notifiarr')
  })

  it('honors an explicit snippet description', async () => {
    const snip = createRawSnippet(() => ({ render: () => '<p>SNIPPET-DESC</p>' }))
    render(Header, { props: { page: { id: 'Configuration' }, description: snip } })
    await waitFor(() => expect(screen.getByText('SNIPPET-DESC')).toBeTruthy())
    expect(body()).not.toContain('Configure your Notifiarr')
  })

  it('keeps html links from the locale string', async () => {
    render(Header, { props: { page: { id: 'SiteTunnel' } } })
    await waitFor(() =>
      expect(
        document.querySelector('a[href="https://github.com/golift/mulery"]'),
      ).toBeTruthy(),
    )
  })

  it('falls back cleanly when a locale lacks the key entirely', async () => {
    await locale.set('en')
    render(Header, { props: { page: { id: 'Configuration' } } })
    await waitFor(() => expect(title()).toContain('Configuration'))
    await addMessages('yy', { navigation: { titles: { Configuration: 'YY-TITEL' } } })
    await locale.set('yy')
    await waitFor(() => expect(title()).toContain('YY-TITEL'), { timeout: 2000 })
    expect(body()).not.toContain('navigation.pageDescription')
  })
})
