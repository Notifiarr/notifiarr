import Configuration, { page as ConfigP } from '../pages/configuration/Index.svelte'
import SiteTunnel, { page as SiteTunnelP } from '../pages/siteTunnel/Index.svelte'
import StarrApps, { page as StarrAppsP } from '../pages/starrApps/Index.svelte'
import DownloadApps, { page as DownloadAppsP } from '../pages/downloadApps/Index.svelte'
import MediaApps, { page as MediaAppsP } from '../pages/mediaApps/Index.svelte'
import SnapshotApps, { page as SnapshotAppsP } from '../pages/snapshotApps/Index.svelte'
import FileWatcher, { page as FileWatcherPage } from '../pages/fileWatcher/Index.svelte'
import Endpoints, { page as EndpointsPage } from '../pages/endpoints/Index.svelte'
import Commands, { page as CommandsPage } from '../pages/commands/Index.svelte'
import ServiceChecks, { page as ServicesP } from '../pages/serviceChecks/Index.svelte'
import Actions, { page as ActionsP } from '../pages/actions/Index.svelte'
import Integrations, { page as IntegrationsP } from '../pages/integrations/Index.svelte'
import Monitoring, { page as MonitoringP } from '../pages/monitoring/Index.svelte'
import Metrics, { page as MetricsP } from '../pages/stats/Index.svelte'
import LogFiles, { page as LogFilesP } from '../pages/logFiles/Index.svelte'
import System, { page as SystemP } from '../pages/system/Index.svelte'
import Profile, { page as ProfilePage } from '../pages/profile/Index.svelte'
import Landing, { page as LandingPage } from '../Landing.svelte'
import type { Page } from './nav.svelte'
import ProcessList, { page as ProcessListPage } from '../pages/stubs/ProcessList.svelte'
import ClientInfo, { page as ClientInfoPage } from '../pages/stubs/ClientInfo.svelte'
import ApiDocs, { page as ApiDocsPage } from '../pages/stubs/ApiDocs.svelte'
import Languages, { page as LanguagesPage } from '../pages/stubs/Languages.svelte'
import TestAll, { page as TestAllPage } from '../pages/stubs/TestAllInstances.svelte'
import Plex, { page as PlexPage } from '../pages/mediaApps/Plex.svelte'

// Page structure for navigation with icons
// 'id' (from page) is used for navigation AND translations.

// Settings header in navigation menu.
export const settings: Page[] = [
  { component: Configuration, ...ConfigP },
  { component: SiteTunnel, ...SiteTunnelP },
  { component: StarrApps, ...StarrAppsP },
  { component: DownloadApps, ...DownloadAppsP },
  { component: MediaApps, ...MediaAppsP },
  { component: SnapshotApps, ...SnapshotAppsP },
  { component: FileWatcher, ...FileWatcherPage },
  { component: Endpoints, ...EndpointsPage },
  { component: Commands, ...CommandsPage },
  { component: ServiceChecks, ...ServicesP },
]
// Insights header in navigation menu.
export const insights: Page[] = [
  { component: Actions, ...ActionsP },
  { component: Integrations, ...IntegrationsP },
  { component: Monitoring, ...MonitoringP },
  { component: Metrics, ...MetricsP },
  { component: LogFiles, ...LogFilesP },
  { component: System, ...SystemP },
]
// Others do not show up in the navigation menu.
export const others: Page[] = [
  { component: Profile, ...ProfilePage },
  { component: Landing, ...LandingPage },
  { component: ProcessList, ...ProcessListPage },
  { component: ClientInfo, ...ClientInfoPage },
  { component: ApiDocs, ...ApiDocsPage },
  { component: Languages, ...LanguagesPage },
  { component: TestAll, ...TestAllPage },
  { component: Plex, ...PlexPage },
]

export const allPages: Page[] = [...settings, ...insights, ...others]
