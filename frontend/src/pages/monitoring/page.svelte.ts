import { faDisplayChartUpCircleCurrency } from '@fortawesome/sharp-duotone-regular-svg-icons'
import { getUi } from '../../api/fetch'
import { delay, success, warning } from '../../includes/util'
import { get } from 'svelte/store'
import { _ } from 'svelte-i18n'
import type { ClientServicesConfig as ServicesConfig } from '../../api/notifiarrConfig'

export const page = {
  id: 'Monitoring',
  i: faDisplayChartUpCircleCurrency,
  c1: 'darkslateblue',
  c2: 'gainsboro',
  d1: 'papayawhip',
  d2: 'paleturquoise',
}

class Mon {
  public refresh = $state(false)
  public checking = $state<Record<string, boolean>>({})
  public config = $state<ServicesConfig>({ results: [], running: true, disabled: false })

  public states: Record<number, string> = {
    0: 'OK',
    1: 'Warning',
    2: 'Critical',
    3: 'Unknown',
  }
  public colors: Record<number, string> = {
    0: 'success',
    1: 'warning',
    2: 'danger',
    3: 'info',
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
