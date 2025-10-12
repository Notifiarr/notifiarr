import { get } from 'svelte/store'
import { _ } from '../../includes/Translate.svelte'
import { deepCopy } from '../../includes/util'
import type { StarrConfig } from '../../api/notifiarrConfig'
import { profile } from '../../api/profile.svelte'
import sonarrLogo from '../../assets/logos/sonarr.png'
import radarrLogo from '../../assets/logos/radarr.png'
import readarrLogo from '../../assets/logos/readarr.png'
import lidarrLogo from '../../assets/logos/lidarr.png'
import prowlarrLogo from '../../assets/logos/prowlarr.png'
import { faStars } from '@fortawesome/sharp-duotone-regular-svg-icons'
import { validate as validator } from '../../includes/instanceValidator'
import type { App } from '../../includes/formsTracker.svelte'

export const page = {
  id: 'StarrApps',
  i: faStars,
  c1: 'darkgoldenrod',
  c2: 'gold',
  d1: 'gold',
  d2: 'darkgoldenrod',
}

export const starrConfig: StarrConfig = {
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
  static readonly tabs = ['sonarr', 'radarr', 'readarr', 'lidarr', 'prowlarr']

  static get title(): Record<string, string> {
    return {
      Sonarr: get(_)('StarrApps.Sonarr.title'),
      Radarr: get(_)('StarrApps.Radarr.title'),
      Readarr: get(_)('StarrApps.Readarr.title'),
      Lidarr: get(_)('StarrApps.Lidarr.title'),
      Prowlarr: get(_)('StarrApps.Prowlarr.title'),
    }
  }

  static readonly Sonarr: App<StarrConfig> = {
    name: 'Sonarr',
    id: page.id + '.Sonarr',
    envPrefix: 'SONARR',
    logo: sonarrLogo,
    empty: starrConfig,
    merge: (index: number, form: StarrConfig) => {
      const c = deepCopy(get(profile).config)
      if (!c.sonarr) c.sonarr = []
      c.sonarr[index] = form
      return c
    },
    validator,
  }

  static readonly Radarr: App<StarrConfig> = {
    name: 'Radarr',
    id: page.id + '.Radarr',
    envPrefix: 'RADARR',
    logo: radarrLogo,
    empty: starrConfig,
    merge: (index: number, form: StarrConfig) => {
      const c = deepCopy(get(profile).config)
      if (!c.radarr) c.radarr = []
      c.radarr[index] = form
      return c
    },
    validator,
  }

  static readonly Readarr: App<StarrConfig> = {
    name: 'Readarr',
    id: page.id + '.Readarr',
    logo: readarrLogo,
    envPrefix: 'READARR',
    empty: starrConfig,
    merge: (index: number, form: StarrConfig) => {
      const c = deepCopy(get(profile).config)
      if (!c.readarr) c.readarr = []
      c.readarr[index] = form
      return c
    },
    validator,
  }

  static readonly Lidarr: App<StarrConfig> = {
    name: 'Lidarr',
    id: page.id + '.Lidarr',
    logo: lidarrLogo,
    envPrefix: 'LIDARR',
    empty: starrConfig,
    merge: (index: number, form: StarrConfig) => {
      const c = deepCopy(get(profile).config)
      if (!c.lidarr) c.lidarr = []
      c.lidarr[index] = form
      return c
    },
    validator,
  }

  static readonly Prowlarr: App<StarrConfig> = {
    name: 'Prowlarr',
    id: page.id + '.Prowlarr',
    logo: prowlarrLogo,
    envPrefix: 'PROWLARR',
    empty: starrConfig,
    merge: (index: number, form: StarrConfig) => {
      const c = deepCopy(get(profile).config)
      if (!c.prowlarr) c.prowlarr = []
      c.prowlarr[index] = form
      return c
    },
    validator,
  }
}
