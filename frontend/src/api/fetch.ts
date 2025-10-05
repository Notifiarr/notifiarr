import { get, writable } from 'svelte/store'
import { delay, failure, ltrim, rtrim, success } from '../includes/util'
import { _ } from 'svelte-i18n'
import Cookies from 'js-cookie'
import { locale } from '../includes/locale/index.svelte'
import { profile } from './profile.svelte'

export const LoggedOut = new Error('logged out')
export const TimedOut = new Error('request timed out')

/** The base URL of the backend's http interface. */
export const urlbase = writable<string>(Cookies.get('urlbase') || '/')

/** The response from the backend. We avoid throwing exceptions in the wrapper methods, and return this object instead. */
export type BackendResponse = { ok: boolean; body: any }

/**
 * Get a UI resource.
 * @param uri The URI of the resource to get.
 * @param json Whether to parse the response as JSON.
 * @returns A promise that resolves to the response body as either text or JSON.
 */
export async function getUi(
  uri: string,
  json: boolean = true,
  timeout: number = 5000,
): Promise<BackendResponse> {
  return await request('ui/' + uri, 'GET', null, json)
}

/**
 * Post a UI resource.
 * @param uri The URI of the resource to post.
 * @param body The body of the resource to post.
 * @param json Whether to parse the response as JSON.
 * @returns A promise that resolves to the response body as either text or JSON.
 */
export async function postUi(
  uri: string,
  body: BodyInit,
  json: boolean = true,
  timeout: number = 5000,
): Promise<BackendResponse> {
  return await request('ui/' + uri, 'POST', body, json, timeout)
}

/**
 * Get an API resource.
 * @param uri The URI of the resource to get.
 * @param json Whether to parse the response as JSON.
 * @returns A promise that resolves to the response body as either text or JSON.
 */
export async function getApi(
  uri: string,
  json: boolean = true,
): Promise<BackendResponse> {
  return await request('api/' + uri, 'GET', null, json)
}

/**
 * Post an API resource.
 * @param uri The URI of the resource to post.
 * @param body The body of the resource to post.
 * @param json Whether to parse the response as JSON.
 * @returns A promise that resolves to the response body as either text or JSON.
 */
export async function postApi(
  uri: string,
  body: BodyInit,
  json: boolean = true,
): Promise<BackendResponse> {
  return await request('api/' + uri, 'POST', body, json)
}

/**
 * Check if the server has finished reloading its configuration. This is an UP check.
 * This is used to wait for the server to reload the page after a configuration change.
 * @returns A promise that resolves when the server has reloaded and is available for requests.
 */
export async function checkReloaded(): Promise<void> {
  const checkReload = async () => {
    if ((await getUi('ping', false)).ok) {
      success(get(_)('phrases.ReloadSuccess'))
      return true
    }

    return false
  }

  return new Promise(async (resolve, reject) => {
    await delay(800) // initial delay (this + the first delay)
    // This waits up to about 40 seconds, and makes 50 attempts.
    for (let i = 0; i < 50; i++) {
      // delay between checks
      await delay(i < 5 ? 250 : i < 12 ? 400 : i < 18 ? 500 : i < 25 ? 600 : 1000)
      if (await checkReload()) return resolve()
    }

    const err = get(_)('phrases.ReloadCheckTimedOut')
    failure(err)
    reject(new Error(err))
  })
}

/** Internal abstraction for all the small public methods above. */
async function request(
  uri: string,
  method: string = 'GET',
  body: BodyInit | null = null,
  json: boolean = true,
  timeout: number = 5000,
): Promise<BackendResponse> {
  try {
    const headers: HeadersInit = {
      Accept: json ? 'application/json' : 'text/plain',
      'Accept-Language': locale.current,
    }
    if (body) headers['Content-Type'] = 'application/json'
    if (uri.startsWith('api/')) headers['X-Api-Key'] = get(profile).config.apiKey

    uri = rtrim(get(urlbase), '/') + '/' + ltrim(uri, '/')
    const response = await fetchWithTimeout(uri, { method, headers, body }, timeout)
    if (response.status === 403) throw LoggedOut
    if (!response.ok)
      throw new Error(
        `${method} ${uri} failed: ${response.status} ${response.statusText}: ${await response.text()}`,
      )

    if (json) return { ok: true, body: await response.json() }
    return { ok: true, body: await response.text() }
  } catch (error) {
    return { ok: false, body: (error as Error).message }
  }
}

/** Generic fetch function with a timeout. */
export async function fetchWithTimeout(
  url: string,
  options: RequestInit = {},
  timeout = 5000,
): Promise<Response> {
  const controller = new AbortController()
  const id = setTimeout(() => controller.abort(), timeout)
  options.mode = 'same-origin'
  options.signal = controller.signal
  options.redirect = 'manual'

  try {
    const response = await fetch(url, options)
    clearTimeout(id)
    return response
  } catch (error) {
    clearTimeout(id)
    if (error instanceof DOMException && error.name === 'AbortError') throw TimedOut
    throw error
  }
}
