import { readable, type Readable, type Unsubscriber } from 'svelte/store'
import Cookies from 'js-cookie'

const themes = ['dark', 'light']
const cookieName = 'theme'

/** We save theme in a cookie and a svelte store, so it persists across page loads. */
class ThemeClass {
  private theme: Readable<string>
  public change: (value: string) => void = () => {}
  public list = themes
  public isDark = $state(true)

  constructor() {
    this.theme = readable('dark', set => {
      this.change = (theme: string) => (
        set(theme), this.setVars(theme), Cookies.set(cookieName, theme)
      )
      this.change(Cookies.get(cookieName) ?? 'dark')
    })
  }

  public subscribe = (run: (value: string) => void): Unsubscriber =>
    this.theme.subscribe(run)

  private setVars = (newTheme: string) => {
    this.isDark = newTheme.includes('dark')
    this.isDark
      ? window.document.body.classList.add('dark-mode')
      : window.document.body.classList.remove('dark-mode')
  }

  /** Use this to toggle the current theme.
   *  This will go away once we have multiple themes.
   */
  public toggle = (e: Event) => (
    e.preventDefault(), theme.change(this.isDark ? 'light' : 'dark')
  )
}

/** Use this to get or change the current theme or list all themes. */
export const theme = new ThemeClass()
