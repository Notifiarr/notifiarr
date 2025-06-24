import { get } from 'svelte/store'
import { urlbase } from '../../api/fetch'
import type { LogFileInfo } from '../../api/notifiarrConfig'
import { failure, success } from '../../includes/util'

export class FileTail {
  private list: string[] = $state([])
  private socket: WebSocket
  private limit: number
  private url: string
  public body: string = $derived(this.list.join('\n'))
  public ok: boolean = $state(false)

  constructor(file: LogFileInfo, limit: number) {
    this.limit = limit
    this.url = `${location.origin.replace(/^http/, 'ws')}${get(urlbase)}ui/ws?source=logs&fileId=${file.id}`
    this.socket = new WebSocket(this.url)
    this.socket.onopen = this.onopen
    this.socket.onmessage = this.onmessage
    this.socket.onerror = this.onerror
    this.socket.onclose = () => (this.ok = false)
  }

  public async destroy() {
    try {
      await this.socket.close()
    } catch (error) {
      failure(`Error destroying websocket: ${error}`)
    } finally {
      this.list = []
      this.body = ''
      this.ok = false
    }
  }

  onmessage = (event: MessageEvent) => {
    this.list.push(event.data)
    // Limit the number of lines to the provided limit.
    if (this.limit != 0 && this.list.length > this.limit) this.list.shift()
  }

  onerror = (event: Event) => {
    failure(`Websocket closed`)
    this.ok = false
    this.socket.close()
  }

  onopen = () => {
    this.ok = true
    success(`Socket Opened: ${this.url}`)
  }
}
