import { writable } from 'svelte/store'
import Cookies from 'js-cookie'

export const urlbase = writable<string>(Cookies.get('urlbase') || '/')
