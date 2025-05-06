export function trimPrefix(str: string, prefix: string) {
  return str.slice(str.startsWith(prefix) ? prefix.length : 0)
}
