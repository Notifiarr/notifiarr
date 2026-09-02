import { render, screen } from '@testing-library/svelte'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { profile } from '../api/profile.svelte'
import Input from './Input.svelte'

vi.mock('../api/profile.svelte', async () => {
  const { writable } = await import('svelte/store')
  return { profile: writable({ environment: {} as Record<string, string> }) }
})

function setProfile(environment: Record<string, string>) {
  ;(profile as unknown as { set: (v: { environment: Record<string, string> }) => void }).set({
    environment,
  })
}

describe('Input', () => {
  beforeEach(() => {
    setProfile({})
  })

  it('keeps a caller class when the value is unchanged', () => {
    render(Input, {
      props: { id: 'test.plain', value: 'same', original: 'same', class: 'mb-0' },
    })

    const el = screen.getByRole('textbox')
    expect(el).toHaveClass('mb-0')
    expect(el).not.toHaveClass('changed')
  })

  it('keeps a caller class next to dirty highlighting', () => {
    render(Input, {
      props: { id: 'test.plain', value: 'new', original: 'old', class: 'mb-0' },
    })

    const el = screen.getByRole('textbox')
    expect(el).toHaveClass('mb-0')
    expect(el).toHaveClass('changed')
    expect(el).toHaveClass('is-valid')
  })

  it('does not show the env-var tooltip when envVar is missing', () => {
    render(Input, { props: { id: 'test.plain', value: '', original: '' } })

    expect(screen.queryByTitle('Show more information')).toBeNull()
    expect(screen.queryByText(/DN_undefined/)).toBeNull()
  })

  it('shows the env-var tooltip when envVar is set and the field is enabled', async () => {
    setProfile({ DN_API_KEY: 'set' })

    render(Input, {
      props: { id: 'test.plain', value: '', original: '', envVar: 'API_KEY' },
    })

    const toggle = screen.getByTitle('Show more information')
    toggle.click()
    expect(await screen.findByText(/DN_API_KEY/)).toBeTruthy()
  })

  it('does not show the env-var tooltip when the field is disabled', () => {
    render(Input, {
      props: {
        id: 'test.plain',
        value: '',
        original: '',
        envVar: 'API_KEY',
        disabled: true,
      },
    })

    expect(screen.queryByTitle('Show more information')).toBeNull()
  })

  it('inverts Enabled/Disabled when invert is true', () => {
    render(Input, {
      props: {
        id: 'config.debug',
        type: 'select',
        value: true,
        original: true,
        invert: true,
      },
    })

    const select = screen.getByRole('combobox') as HTMLSelectElement
    expect(select.value).toBe('true')
    expect(select.selectedOptions[0].textContent?.trim()).toBe('Disabled')
  })

  it('does not invert Enabled/Disabled when invert is false', () => {
    render(Input, {
      props: {
        id: 'apps.sonarr.disabled',
        type: 'select',
        value: true,
        original: true,
        invert: false,
      },
    })

    const select = screen.getByRole('combobox') as HTMLSelectElement
    expect(select.value).toBe('true')
    expect(select.selectedOptions[0].textContent?.trim()).toBe('Enabled')
  })

  it('derives validation feedback without mutating caller state', () => {
    const seen: string[] = []
    render(Input, {
      props: {
        id: 'test.plain',
        value: 'bad',
        original: 'bad',
        validate: (_id: string, value: string) => {
          seen.push(value)
          return value === 'bad' ? 'nope' : ''
        },
      },
    })

    expect(screen.getByText('nope')).toBeTruthy()
    expect(screen.getByRole('textbox')).toHaveAttribute('aria-invalid', 'true')
    expect(seen).toContain('bad')
  })
})

