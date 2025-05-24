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
  static get title(): Record<string, string> {
    return {
      [Starr.Sonarr.name]: get(_)('StarrApps.Sonarr.title'),
      [Starr.Radarr.name]: get(_)('StarrApps.Radarr.title'),
      [Starr.Readarr.name]: get(_)('StarrApps.Readarr.title'),
      [Starr.Lidarr.name]: get(_)('StarrApps.Lidarr.title'),
      [Starr.Prowlarr.name]: get(_)('StarrApps.Prowlarr.title'),
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

  static readonly tabs = [
    Starr.Sonarr.name.toLowerCase(),
    Starr.Radarr.name.toLowerCase(),
    Starr.Readarr.name.toLowerCase(),
    Starr.Lidarr.name.toLowerCase(),
    Starr.Prowlarr.name.toLowerCase(),
  ]
}
