import { get } from 'svelte/store'
import { _ } from '../../includes/Translate.svelte'
import { deepCopy } from '../../includes/util'
import type {
  DelugeConfig,
  QbitConfig,
  XmissionConfig,
  RtorrentConfig,
  SabNZBConfig,
  NZBGetConfig,
} from '../../api/notifiarrConfig'
import type { Form } from '../../includes/Instance.svelte'
import { profile } from '../../api/profile.svelte'
import qbitLogo from '../../assets/logos/qbittorrent.png'
import rtorrentLogo from '../../assets/logos/rtorrent.png'
import xmissionLogo from '../../assets/logos/transmission.png'
import delugeLogo from '../../assets/logos/deluge.png'
import sabnzbLogo from '../../assets/logos/sabnzb.png'
import nzbgetLogo from '../../assets/logos/nzbget.png'
import { faDownload } from '@fortawesome/sharp-duotone-regular-svg-icons'
import type { InstanceFormValidator } from '../../includes/instanceFormValidator.svelte'

export const page = {
  id: 'Downloaders',
  i: faDownload,
  c1: 'brown',
  c2: 'lightsalmon',
  d1: 'coral',
  d2: 'lightpink',
}

const downloadConfig: Form = {
  name: '',
  timeout: '10s',
  interval: '5m0s',
  validSsl: false,
  url: '',
  username: '', // not for deluge or sabnzb.
  password: '', // not for sabnzb.
  /* SabNZB only */
  apiKey: '',
}

export class Downloaders {
  static get title(): Record<string, string> {
    return {
      [Downloaders.Qbittorrent.name]: get(_)('Downloaders.Qbittorrent.title'),
      [Downloaders.Rtorrent.name]: get(_)('Downloaders.Rtorrent.title'),
      [Downloaders.Xmission.name]: get(_)('Downloaders.Transmission.title'),
      [Downloaders.Deluge.name]: get(_)('Downloaders.Deluge.title'),
      [Downloaders.SabNZB.name]: get(_)('Downloaders.SabNZB.title'),
      [Downloaders.NZBGet.name]: get(_)('Downloaders.NZBGet.title'),
    }
  }

  static readonly Qbittorrent = {
    name: 'Qbittorrent',
    id: page.id + '.Qbittorrent',
    logo: qbitLogo,
    hidden: ['apiKey', 'deletes'],
    empty: downloadConfig,
    merge: (index: number, form: Form) => {
      const c = deepCopy(get(profile).config)
      if (!c.qbit) c.qbit = []
      c.qbit[index] = form as QbitConfig
      return c
    },
  }

  static readonly Rtorrent = {
    name: 'Rtorrent',
    id: page.id + '.Rtorrent',
    logo: rtorrentLogo,
    hidden: ['apiKey', 'deletes'],
    empty: downloadConfig,
    merge: (index: number, form: Form) => {
      const c = deepCopy(get(profile).config)
      if (!c.rtorrent) c.rtorrent = []
      c.rtorrent[index] = form as RtorrentConfig
      return c
    },
  }

  static readonly Xmission = {
    name: 'Transmission',
    id: page.id + '.Transmission',
    logo: xmissionLogo,
    hidden: ['apiKey', 'deletes'],
    empty: downloadConfig,
    merge: (index: number, form: Form) => {
      const c = deepCopy(get(profile).config)
      if (!c.transmission) c.transmission = []
      c.transmission[index] = form as XmissionConfig
      return c
    },
  }

  static readonly Deluge = {
    name: 'Deluge',
    id: page.id + '.Deluge',
    logo: delugeLogo,
    hidden: ['username', 'apiKey', 'deletes'],
    empty: downloadConfig,
    merge: (index: number, form: Form) => {
      const c = deepCopy(get(profile).config)
      if (!c.deluge) c.deluge = []
      c.deluge[index] = form as DelugeConfig
      return c
    },
  }

  static readonly SabNZB = {
    name: 'SabNZB',
    id: page.id + '.SabNZB',
    logo: sabnzbLogo,
    hidden: ['username', 'password', 'deletes'],
    empty: downloadConfig,
    merge: (index: number, form: Form) => {
      const c = deepCopy(get(profile).config)
      if (!c.sabnzbd) c.sabnzbd = []
      c.sabnzbd[index] = form as SabNZBConfig
      return c
    },
  }

  static readonly NZBGet = {
    name: 'NZBGet',
    id: page.id + '.NZBGet',
    logo: nzbgetLogo,
    hidden: ['apiKey', 'deletes'],
    empty: downloadConfig,
    merge: (index: number, form: Form) => {
      const c = deepCopy(get(profile).config)
      if (!c.nzbget) c.nzbget = []
      c.nzbget[index] = form as NZBGetConfig
      return c
    },
  }

  // Keep track of the navigation.
  static readonly tabs = [
    Downloaders.Qbittorrent.name.toLowerCase(),
    Downloaders.Rtorrent.name.toLowerCase(),
    Downloaders.Xmission.name.toLowerCase(),
    Downloaders.Deluge.name.toLowerCase(),
    Downloaders.SabNZB.name.toLowerCase(),
    Downloaders.NZBGet.name.toLowerCase(),
  ]
}
