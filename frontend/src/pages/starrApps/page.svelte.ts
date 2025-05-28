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
import type { App } from '../../includes/formsTracker.svelte'
import { profile } from '../../api/profile.svelte'
import sonarrLogo from '../../assets/logos/sonarr.png'
import radarrLogo from '../../assets/logos/radarr.png'
import readarrLogo from '../../assets/logos/readarr.png'
import lidarrLogo from '../../assets/logos/lidarr.png'
import prowlarrLogo from '../../assets/logos/prowlarr.png'
import { faStars } from '@fortawesome/sharp-duotone-regular-svg-icons'
import { validate } from '../../includes/instanceValidator'

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
      Sonarr: get(_)('StarrApps.Sonarr.title'),
      Radarr: get(_)('StarrApps.Radarr.title'),
      Readarr: get(_)('StarrApps.Readarr.title'),
      Lidarr: get(_)('StarrApps.Lidarr.title'),
      Prowlarr: get(_)('StarrApps.Prowlarr.title'),
    }
  }

  static readonly getValidator = (
    app: 'sonarr' | 'radarr' | 'readarr' | 'lidarr' | 'prowlarr',
  ): ((id: string, value: any, index: number) => string) => {
    return (id: string, value: any, index: number) => {
      return validate(id, value, index, get(profile).config[app] ?? [])
    }
  }

  static readonly Sonarr: App<SonarrConfig> = {
    name: 'Sonarr',
    id: page.id + '.Sonarr',
    logo: sonarrLogo,
    empty: starrConfig,
    merge: (index: number, form: SonarrConfig) => {
      const c = deepCopy(get(profile).config)
      if (!c.sonarr) c.sonarr = []
      c.sonarr[index] = form
      return c
    },
    validator: Starr.getValidator('sonarr'),
  }

  static readonly Radarr: App<RadarrConfig> = {
    name: 'Radarr',
    id: page.id + '.Radarr',
    logo: radarrLogo,
    empty: starrConfig,
    merge: (index: number, form: RadarrConfig) => {
      const c = deepCopy(get(profile).config)
      if (!c.radarr) c.radarr = []
      c.radarr[index] = form
      return c
    },
    validator: Starr.getValidator('radarr'),
  }

  static readonly Readarr: App<ReadarrConfig> = {
    name: 'Readarr',
    id: page.id + '.Readarr',
    logo: readarrLogo,
    empty: starrConfig,
    merge: (index: number, form: ReadarrConfig) => {
      const c = deepCopy(get(profile).config)
      if (!c.readarr) c.readarr = []
      c.readarr[index] = form
      return c
    },
    validator: Starr.getValidator('readarr'),
  }

  static readonly Lidarr: App<LidarrConfig> = {
    name: 'Lidarr',
    id: page.id + '.Lidarr',
    logo: lidarrLogo,
    empty: starrConfig,
    merge: (index: number, form: LidarrConfig) => {
      const c = deepCopy(get(profile).config)
      if (!c.lidarr) c.lidarr = []
      c.lidarr[index] = form
      return c
    },
    validator: Starr.getValidator('lidarr'),
  }

  static readonly Prowlarr: App<ProwlarrConfig> = {
    name: 'Prowlarr',
    id: page.id + '.Prowlarr',
    logo: prowlarrLogo,
    empty: starrConfig,
    merge: (index: number, form: ProwlarrConfig) => {
      const c = deepCopy(get(profile).config)
      if (!c.prowlarr) c.prowlarr = []
      c.prowlarr[index] = form
      return c
    },
    validator: Starr.getValidator('prowlarr'),
  }

  static readonly tabs = ['Sonarr', 'Radarr', 'Readarr', 'Lidarr', 'Prowlarr']
}
