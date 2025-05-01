export async function fetchWithTimeout(url: string, options: RequestInit = {}, timeout = 5000): Promise<Response> {
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

export function trimPrefix(str: string, prefix: string) {
  if (str.startsWith(prefix)) return str.slice(prefix.length)
  return str
}
