import { get, readable, type Readable, type Unsubscriber } from 'svelte/store'
import Cookies from 'js-cookie'

const themes = ['dark', 'light']
const cookieName = 'theme'

/** We save theme in a cookie and a svelte store, so it persists across page loads. */
class ThemeClass {
  private theme: Readable<string>
  private set: (value: string) => void = () => {}
  public list = themes

  constructor() {
    this.theme = readable('dark', set => {
      this.set = set
      const theme = Cookies.get(cookieName) ?? 'dark'
      this.set(theme)
      this.setBG(theme)
    })
  }

  public subscribe(run: (value: string) => void): Unsubscriber {
    return this.theme.subscribe(run)
  }

  /** Use this to toggle the current theme.
   *  This will go away once we have multiple themes.
   */
  public toggle(e: Event) {
    e.preventDefault()
    theme.change(get(theme).includes('dark') ? 'light' : 'dark')
  }

  public change(newTheme: string) {
    Cookies.set(cookieName, newTheme)
    this.set(newTheme)
    this.setBG(newTheme)
  }

  private setBG(newTheme: string) {
    if (newTheme.includes('dark')) {
      window.document.body.classList.add('dark-mode')
    } else {
      window.document.body.classList.remove('dark-mode')
    }
  }
}

/** Use this to get or change the current theme or list all themes. */
export const theme = new ThemeClass()
