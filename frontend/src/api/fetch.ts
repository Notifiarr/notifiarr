import { get } from 'svelte/store'
import { profile } from './profile'

export const LoggedOut = new Error('logged out')

/**
 * Get a UI resource.
 * @param uri The URI of the resource to get.
 * @param json Whether to parse the response as JSON.
 * @returns A promise that resolves to the response body as either text or JSON.
 */
export async function getUi(uri: string, json: boolean = true): Promise<string | null> {
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
): Promise<string | null> {
  return await request('ui/' + uri, 'POST', body, json)
}

/**
 * Get an API resource.
 * @param uri The URI of the resource to get.
 * @param json Whether to parse the response as JSON.
 * @returns A promise that resolves to the response body as either text or JSON.
 */
export async function getApi(uri: string, json: boolean = true): Promise<string | null> {
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
): Promise<string | null> {
  return await request('api/' + uri, 'POST', body, json)
}

/**
 * Check if the server has finished reloading its configuration. This is an UP check.
 * This is used to wait for the server to reload the page after a configuration change.
 * @returns A promise that resolves when the server has reloaded the page.
 */
export async function checkReloaded() {
  return new Promise<void>(async resolve => {
    await setTimeout(async () => {
      const checkReload = async (attempts = 0) => {
        if (attempts > 19)
          throw new Error('Server reload check timed out after 20 attempts')
        try {
          await getUi('ping', false)
          resolve()
        } catch {
          await setTimeout(() => checkReload(attempts + 1), 300)
        }
      }
      await checkReload()
    }, 600)
  })
}

async function request(
  uri: string,
  method: string = 'GET',
  body: BodyInit | null = null,
  json: boolean = true,
): Promise<string | null> {
  const urlBase = get(profile)?.config.urlbase || ''

  const headers: Record<string, string> = {}
  if (json) headers['Accept'] = 'application/json'
  else headers['Accept'] = 'text/plain'
  if (body) headers['Content-Type'] = 'application/json'

  const response = await fetchWithTimeout(urlBase + uri, { method, headers, body })
  if (response.status === 403) throw LoggedOut

  if (!response.ok)
    throw new Error(
      `Failed to fetch ${uri}: ${response.status} ${response.statusText}: ${response.text()}`,
    )

  if (json) return response.json()
  return response.text()
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
    throw error
  }
}
