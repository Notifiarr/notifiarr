import { get, writable } from 'svelte/store'
import { urlbase } from '../api/fetch'
import type { Component } from 'svelte'
import Landing from '../Landing.svelte'
import { iequals, ltrim } from '../includes/util'
import type { Props as FaProps } from '../includes/Fa.svelte'
import { allPages } from './pages'
import { closeSidebar } from './Index.svelte'

// Page represents the data to render a page link.
export interface Page extends FaProps {
  id: string
  component: Component
}

class Navigator {
  /** This is the current page component on screen. */
  public ActivePage: Component = $state(Landing)
  /** This is the current page id on screen. */
  private activePage = $state('')
  /** This is set to true if any form has changes. */
  public formChanged = $state(false)
  /** This is set to the page id if there are unsaved changes. */
  public showUnsavedAlert = $state('')
  /** This is used to force a navigation event even if there are unsaved changes. */
  public forceEvent = { type: 'force', preventDefault: () => {} } as Event

  /** Call this in the onMount function of the parent component to set the initial page. */
  public onMount = () => {
    // Navigate to the initial page based on the URL when the content mounts.
    const parts = ltrim(window.location.pathname, get(urlbase)).split('/')
    const page = this.setActivePage(parts.length > 0 ? parts[0] : '')
    // Fix the url in the browser if it doesn't match a page.
    if (page === '') window.history.replaceState({ uri: '' }, '', get(urlbase))
    // Otherwise, update the url in the browser to the current page.
    // else window.history.replaceState({ uri: page }, '', get(urlbase) + page)
  }

  /**
   * Used to navigate to a page.
   * @param event - from an onclick handler, optional.
   * @param pid - the id of the page to navigate to, ie profile, configuration, etc.
   */
  public goto = (event: Event | null, pid: string, subPages: string[] = []): void => {
    event?.preventDefault()

    if (this.formChanged && event?.type !== 'force' && pid !== this.activePage) {
      this.showUnsavedAlert = pid
      return
    }

    this.showUnsavedAlert = ''
    this.formChanged = false
    pid = this.setActivePage(pid)
    closeSidebar()
    const params = new URLSearchParams(window.location.search).toString()
    const path = [pid, ...subPages].join('/').toLowerCase()
    const uri = `${get(urlbase)}${path}${params ? `?${params}` : ''}`
    // Push the new page into the browser history.
    window.history.pushState({ uri: this.activePage }, '', uri)
  }

  /**
   * Used to set a query parameter in the url.
   * @param key - the key to set.
   * @param value - the value to set.
   */
  public setQuery = async (key: string, value: string) => {
    const query = new URLSearchParams(window.location.search)
    if (value === '') await query.delete(key)
    else await query.set(key, value)
    const uri = `${window.location.pathname}?${query.toString()}`
    // Replace the current state with the new url + query.
    await window.history.replaceState({ uri: this.activePage }, '', uri)
  }

  public getQuery = (key: string) => new URLSearchParams(window.location.search).get(key)

  // popstate is split from goto(), so we can call it from popstate.
  /**  Call this only when the back button is clicked. */
  public popstate = (e: PopStateEvent) => (
    e.preventDefault(), this.setActivePage(e.state?.uri ?? '')
  )

  /** active returns true if the provided page id is currently selected. */
  public active = (page: string): boolean => iequals(this.activePage, page)

  private setActivePage = (newPage: string): string => {
    const page = allPages.find(p => iequals(p.id, newPage))
    this.ActivePage = page?.component || Landing
    return (this.activePage = page ? newPage : '')
  }
}

export const nav = new Navigator()
