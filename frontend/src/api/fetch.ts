import { get } from 'svelte/store'
import { urlbase } from './urlbase'
import { delay, ltrim, rtrim, success } from '../lib/util'
import { _ } from 'svelte-i18n'
export const LoggedOut = new Error('logged out')
export const TimedOut = new Error('request timed out')

export type BackendResponse = { ok: boolean; body: any }

/**
 * Get a UI resource.
 * @param uri The URI of the resource to get.
 * @param json Whether to parse the response as JSON.
 * @returns A promise that resolves to the response body as either text or JSON.
 */
export async function getUi(uri: string, json: boolean = true): Promise<BackendResponse> {
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
): Promise<BackendResponse> {
  return await request('ui/' + uri, 'POST', body, json)
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
    await delay(600) // initial delay
    for (let i = 0; i < 20; i++) {
      await delay(300) // delay between checks
      if (await checkReload()) return resolve()
    }

    reject(new Error(get(_)('phrases.ReloadCheckTimedOut')))
  })
}

async function request(
  uri: string,
  method: string = 'GET',
  body: BodyInit | null = null,
  json: boolean = true,
): Promise<BackendResponse> {
  try {
    const headers: Record<string, string> = {}
    if (json) headers['Accept'] = 'application/json'
    else headers['Accept'] = 'text/plain'
    if (body) headers['Content-Type'] = 'application/json'

    uri = rtrim(get(urlbase), '/') + '/' + ltrim(uri, '/')
    const response = await fetchWithTimeout(uri, { method, headers, body })
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

export async function fetchWithTimeout(
  url: string,
  options: RequestInit = {},
  timeout = 5000,
): Promise<Response> {
  const controller = new AbortController()
  const id = setTimeout(() => controller.abort(), timeout)

  try {
    const response = await fetch(url, {
      ...options,
      mode: 'same-origin',
      signal: controller.signal,
      redirect: 'manual',
    })
    clearTimeout(id)
    return response
  } catch (error) {
    clearTimeout(id)
    if (error instanceof DOMException && error.name === 'AbortError') throw TimedOut

    throw error
  }
}
