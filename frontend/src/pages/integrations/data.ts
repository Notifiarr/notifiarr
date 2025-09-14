import delugeLogo from '../../assets/logos/deluge.png'
import nzbgetLogo from '../../assets/logos/nzbget.png'
import qbittorrentLogo from '../../assets/logos/qbittorrent.png'
import rtorrentLogo from '../../assets/logos/rtorrent.png'
import sabnzbLogo from '../../assets/logos/sabnzb.png'
import transmissionLogo from '../../assets/logos/transmission.png'
import plexLogo from '../../assets/logos/plex.png'
import tautulliLogo from '../../assets/logos/tautulli.png'
import sonarrLogo from '../../assets/logos/sonarr.png'
import radarrLogo from '../../assets/logos/radarr.png'
import readarrLogo from '../../assets/logos/readarr.png'
import lidarrLogo from '../../assets/logos/lidarr.png'
import prowlarrLogo from '../../assets/logos/prowlarr.png'

const logos = {
  deluge: delugeLogo,
  nzbget: nzbgetLogo,
  qbit: qbittorrentLogo,
  rtorrent: rtorrentLogo,
  sabnzbd: sabnzbLogo,
  transmission: transmissionLogo,
  plex: plexLogo,
  tautulli: tautulliLogo,
  sonarr: sonarrLogo,
  radarr: radarrLogo,
  readarr: readarrLogo,
  lidarr: lidarrLogo,
  prowlarr: prowlarrLogo,
}

export const getLogo = (app: string) => {
  return logos[app as keyof typeof logos]
}

const titles = {
  deluge: 'Deluge',
  nzbget: 'NZBGet',
  qbit: 'qBittorrent',
  rtorrent: 'Rtorrent',
  sabnzbd: 'SABnzbd',
  transmission: 'Transmission',
  plex: 'Plex',
  tautulli: 'Tautulli',
  sonarr: 'Sonarr',
  radarr: 'Radarr',
  readarr: 'Readarr',
  lidarr: 'Lidarr',
  prowlarr: 'Prowlarr',
}

const colors = {
  deluge: 'info-subtle',
  nzbget: 'success',
  qbit: 'primary-subtle',
  rtorrent: 'primary',
  sabnzbd: 'warning',
  transmission: 'danger',
  plex: 'warning',
  tautulli: 'warning-subtle',
  sonarr: 'info',
  radarr: 'warning',
  readarr: 'danger',
  lidarr: 'success-subtle',
  prowlarr: 'danger-subtle',
}

export const color = (app: string) => {
  return colors[app as keyof typeof colors]
}

export const title = (app: string) => {
  return titles[app as keyof typeof titles]
}
