import { fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import Login from './Login.svelte'

const { login } = vi.hoisted(() => ({
  login: vi.fn<(name: string, password: string) => Promise<string | null>>(),
}))

vi.mock('./api/profile.svelte', () => ({ profile: { login } }))

vi.mock('./header/Index.svelte', () => ({ showMsg: vi.fn() }))

async function typeField(placeholder: string, value: string) {
  await fireEvent.input(screen.getByPlaceholderText(placeholder), { target: { value } })
}

describe('Login', () => {
  beforeEach(() => {
    login.mockReset()
  })

  it('asks for both fields and does not call login when either is empty', async () => {
    render(Login)

    await fireEvent.click(screen.getByRole('button', { name: 'Login' }))

    await waitFor(() =>
      expect(screen.getByText('Please enter both username and password.')).toBeTruthy(),
    )
    expect(login).not.toHaveBeenCalled()
  })

  it('submits the username and password', async () => {
    login.mockResolvedValue('nope')
    render(Login)

    await typeField('Username', 'david')
    await typeField('Password', 'secret')
    await fireEvent.click(screen.getByRole('button', { name: 'Login' }))

    await waitFor(() => expect(login).toHaveBeenCalledWith('david', 'secret'))
    expect(await screen.findByText('Login failed: nope')).toBeTruthy()
  })
})
