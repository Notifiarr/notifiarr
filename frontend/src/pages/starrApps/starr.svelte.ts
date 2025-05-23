import { get } from 'svelte/store'
import { _ } from '../../includes/Translate.svelte'
import { deepCopy } from '../../includes/util'
import type {
  SonarrConfig,
  StarrConfig,
  ExtraConfig,
  RadarrConfig,
  ReadarrConfig,
  LidarrConfig,
  ProwlarrConfig,
} from '../../api/notifiarrConfig'
import type { Form } from '../../includes/Instance.svelte'
import { profile } from '../../api/profile.svelte'
import sonarrLogo from '../../assets/logos/sonarr.png'
import radarrLogo from '../../assets/logos/radarr.png'
import readarrLogo from '../../assets/logos/readarr.png'
import lidarrLogo from '../../assets/logos/lidarr.png'
import prowlarrLogo from '../../assets/logos/prowlarr.png'
import { faStars } from '@fortawesome/sharp-duotone-regular-svg-icons'

export const page = {
  id: 'StarrApps',
  i: faStars,
  c1: 'darkgoldenrod',
  c2: 'gold',
  d1: 'gold',
  d2: 'darkgoldenrod',
}

export const starrConfig: StarrConfig & ExtraConfig = {
  name: '',
  timeout: '10s',
  interval: '5m0s',
  validSsl: false,
  deletes: 0,
  apiKey: '',
  url: '',
  httpPass: '',
  httpUser: '',
  username: '',
  password: '',
}

export class Starr {
  static Validate(
    id: string,
    index: number,
    value: any,
    instances: (Form | null)[],
  ): string {
    const key = id.split('.')[2]

    if (key == 'name') {
      console.log(id, index, value, instances)
      let found = ''
      instances.forEach((m, i) => {
        if (i !== index && m?.name === value) {
          found = get(_)('phrases.NameInUseByInstance', { values: { number: i + 1 } })
          return
        }
      })
      return found ? found : value ? '' : get(_)('phrases.NameMustNotBeEmpty')
    }

    if (key == 'url') {
      return value.startsWith('http://') || value.startsWith('https://')
        ? ''
        : get(_)('phrases.URLMustBeginWithHttp')
    }

    if (key == 'apiKey' && value.length < 32) {
      return get(_)('phrases.APIKeyMustBe32Characters')
    }

    return ''
  }

  static get title() {
    return {
      Sonarr: get(_)('StarrApps.Sonarr.title'),
      Radarr: get(_)('StarrApps.Radarr.title'),
      Readarr: get(_)('StarrApps.Readarr.title'),
      Lidarr: get(_)('StarrApps.Lidarr.title'),
      Prowlarr: get(_)('StarrApps.Prowlarr.title'),
    }
  }

  static readonly Sonarr = {
    name: 'Sonarr',
    id: page.id + '.Sonarr',
    logo: sonarrLogo,
    empty: starrConfig,
    merge: (index: number, form: Form) => {
      const c = deepCopy(get(profile).config)
      if (!c.sonarr) c.sonarr = []
      c.sonarr[index] = form as SonarrConfig
      return c
    },
  }

  static readonly Radarr = {
    name: 'Radarr',
    id: page.id + '.Radarr',
    logo: radarrLogo,
    empty: starrConfig,
    merge: (index: number, form: Form) => {
      const c = deepCopy(get(profile).config)
      if (!c.radarr) c.radarr = []
      c.radarr[index] = form as RadarrConfig
      return c
    },
  }

  static readonly Readarr = {
    name: 'Readarr',
    id: page.id + '.Readarr',
    logo: readarrLogo,
    empty: starrConfig,
    merge: (index: number, form: Form) => {
      const c = deepCopy(get(profile).config)
      if (!c.readarr) c.readarr = []
      c.readarr[index] = form as ReadarrConfig
      return c
    },
  }

  static readonly Lidarr = {
    name: 'Lidarr',
    id: page.id + '.Lidarr',
    logo: lidarrLogo,
    empty: starrConfig,
    merge: (index: number, form: Form) => {
      const c = deepCopy(get(profile).config)
      if (!c.lidarr) c.lidarr = []
      c.lidarr[index] = form as LidarrConfig
      return c
    },
  }

  static readonly Prowlarr = {
    name: 'Prowlarr',
    id: page.id + '.Prowlarr',
    logo: prowlarrLogo,
    empty: starrConfig,
    merge: (index: number, form: Form) => {
      const c = deepCopy(get(profile).config)
      if (!c.prowlarr) c.prowlarr = []
      c.prowlarr[index] = form as ProwlarrConfig
      return c
    },
  }
}
