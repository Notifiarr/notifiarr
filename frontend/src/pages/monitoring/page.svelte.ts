import { getUi } from '../../api/fetch'
import { delay, success, warning } from '../../includes/util'
import { get } from 'svelte/store'
import { _ } from 'svelte-i18n'
import { CheckState, type ClientServicesConfig as ServicesConfig } from '../../api/notifiarrConfig'

export const page = { id: 'Monitoring' }

class Mon {
  public refresh = $state(false)
  public checking = $state<Record<string, boolean>>({})
  public config = $state<ServicesConfig>({ results: [], running: true, disabled: false })

  public states: Record<CheckState, string> = {
    [CheckState.OK]: 'OK',
    [CheckState.Warning]: 'Warning',
    [CheckState.Critical]: 'Critical',
    [CheckState.Unknown]: 'Unknown',
  }
  public colors: Record<CheckState, string> = {
    [CheckState.OK]: 'success',
    [CheckState.Warning]: 'warning',
    [CheckState.Critical]: 'danger',
    [CheckState.Unknown]: 'info',
  }

  public updateBackend = async (e: Event) => {
    this.refresh = true
    e?.preventDefault?.()
    try {
      const resp = await getUi('services/config')
      if (!resp.ok) throw new Error('Failed to get services config')
      this.config = resp.body as ServicesConfig
    } catch (error) {
      warning(`${error}`)
    } finally {
      this.refresh = false
    }
  }

  public check = async (e: Event, name: string) => {
    e?.preventDefault?.()
    if (this.checking[name]) return
    this.checking[name] = true
    const resp = await getUi('services/check/' + name, false)
    if (!resp.ok) warning(get(_)('monitoring.recheckFailed'))
    else {
      success(get(_)('monitoring.checkInitiated'))
      await delay(2000)
      await this.updateBackend(e)
    }
    this.checking[name] = false
  }
}

export const Monitor = new Mon()
