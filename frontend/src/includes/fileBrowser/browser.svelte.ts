import { get } from 'svelte/store'
import { getUi } from '../../api/fetch'
import type { BrowseDir } from '../../api/notifiarrConfig'
import { rtrim, success } from '../util'
import { _ } from '../Translate.svelte'

export class FileBrowser {
  public wd: BrowseDir
  public respErr: string
  public loading: boolean
  public input: string
  private value: string
  private selected: string
  private readonly close: (value: string) => void

  constructor(value: string, close: (value: string) => void) {
    this.value = value
    this.wd = $state({ path: this.value, files: [], dirs: [], sep: '/', mom: '' })
    this.selected = $state(value || '/')
    this.respErr = $state('')
    this.loading = $state(false)
    this.input = $derived(this.wd.path)
    this.close = close
    this.getFiles()
  }

  public readonly cd = (e: Event, to: string, direct = false) => {
    e.preventDefault()
    this.selected = direct ? to : rtrim(this.wd.path, this.wd.sep) + this.wd.sep + to
    this.respErr = ''
    this.getFiles()
  }

  public readonly select = (e: Event, file: string, dir = false) => {
    e.preventDefault()
    this.value = (dir ? '' : rtrim(this.wd.path, this.wd.sep) + this.wd.sep) + file
    this.close(this.value)
  }

  public readonly create = async (path: string, dir = false) => {
    const resp = await getUi('create?dir=' + dir + '&path=' + path, true)
    if (!resp.ok) {
      this.respErr = resp.body
      return
    }

    success(get(_)('FileBrowser.Created', { values: { path } }))
    this.wd = resp.body as BrowseDir
    this.respErr = ''
    // Select the file they just created, and close the picker.
    if (!dir) {
      this.value = path
      this.close(this.value)
    }
  }

  private readonly getFiles = async () => {
    this.loading = true
    const resp = await getUi('browse?dir=' + this.selected, true)
    this.loading = false

    if (resp.ok) {
      this.wd = resp.body as BrowseDir
      this.respErr = ''
      return
    }
    // Set the path to the selected path, and clear the files and directories.
    // Makes navigating out of an error state easier.
    this.wd.mom = this.wd.path
    this.wd.path = this.selected
    this.wd.dirs = this.wd.files = undefined
    this.respErr = resp.body
  }
}
