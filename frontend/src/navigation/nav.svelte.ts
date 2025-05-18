import { get } from 'svelte/store'
import { urlbase } from '../api/fetch'
import type { Component } from 'svelte'
import Landing from '../Landing.svelte'
import { iequals, ltrim } from '../includes/util'
import type { Props as FaProps } from '../includes/Fa.svelte'
import { allPages } from './pages'

// Page represents the data to render a page link.
export interface Page extends FaProps {
  id: string
  component: Component
}

class Navigator {
  public ActivePage: Component = $state(Landing)
  private activePage = $state('')
  private pages = allPages

  /** Call this in the onMount function of the parent component to set the initial page. */
  public onMount() {
    // Navigate to the initial page based on the URL when the content mounts.
    const parts = ltrim(window.location.pathname, get(urlbase)).split('/')
    this.setActivePage(parts.length > 0 ? parts[0] : '')
  }

  /**
   * Used to navigate to a page.
   * @param event - from an onclick handler, optional.
   * @param pid - the id of the page to navigate to, ie profile, configuration, etc.
   */
  public goto(event: Event | null, pid: string, subPages: string[] = []): void {
    event?.preventDefault()
    this.setActivePage(pid)
    const params = new URLSearchParams(window.location.search).toString()
    const path = [pid, ...subPages].join('/').toLowerCase()
    const uri = `${get(urlbase)}${path}${params ? `?${params}` : ''}`
    window.history.pushState({ uri: this.activePage }, '', uri)
  }

  // popstate is split from goto(), so we can call it from popstate.
  /**  Call this only when the back button is clicked. */
  public popstate(e: PopStateEvent) {
    e.preventDefault()
    this.setActivePage(e.state?.uri ?? '')
  }

  public active(check: string): boolean {
    return iequals(this.activePage, check)
  }

  private setActivePage(newPage: string) {
    const page = this.pages.find(p => iequals(p.id, newPage))
    this.activePage = page ? newPage : ''
    this.ActivePage = page?.component || Landing
  }
}

export const nav = new Navigator()
