/* Auto-generated. DO NOT EDIT. Generator: https://golift.io/goty
 * Edit the source code and run goty again to make updates.
 */

/**
 * The day of the week.
 * @see golang: <time.Weekday>
 */
export enum Weekday {
  Sunday    = 0,
  Monday    = 1,
  Tuesday   = 2,
  Wednesday = 3,
  Thursday  = 4,
  Friday    = 5,
  Saturday  = 6,
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/configfile.Config>
 */
export interface Config {
  hostId: string;
  uiPassword: string;
  bindAddr: string;
  sslCertFile: string;
  sslKeyFile: string;
  upstreams?: string[];
  autoUpdate: string;
  unstableCh: boolean;
  timeout: string;
  retries: number;
  snapshot?: SnapshotConfig;
  services?: ServicesConfig;
  service?: (null | Service)[];
  apt: boolean;
  watchFiles?: (null | WatchFile)[];
  endpoints?: (null | Endpoint)[];
  commands?: (null | Command)[];
  LogConfig?: LogConfig;
  Apps?: Apps;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.Config>
 */
export interface SnapshotConfig extends Plugins {
  timeout: string;
  interval: string;
  zfsPools?: string[];
  useSudo: boolean;
  monitorRaid: boolean;
  monitorDrives: boolean;
  monitorSpace: boolean;
  allDrives: boolean;
  quotas: boolean;
  ioTop: number;
  psTop: number;
  myTop: number;
  ipmi: boolean;
  ipmiSudo: boolean;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.Plugins>
 */
export interface Plugins {
  nvidia?: NvidiaConfig;
  mysql?: (null | MySQLConfig)[];
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.NvidiaConfig>
 */
export interface NvidiaConfig {
  smiPath: string;
  busIDs?: string[];
  disabled: boolean;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.MySQLConfig>
 */
export interface MySQLConfig {
  name: string;
  host: string;
  timeout: string;
  interval: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/services.Config>
 */
export interface ServicesConfig {
  interval: string;
  parallel: number;
  disabled: boolean;
  logFile: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/services.Service>
 */
export interface Service {
  name: string;
  type: string;
  value: string;
  expect: string;
  timeout: string;
  interval: string;
  tags?: Record<string, null | any>;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/filewatch.WatchFile>
 */
export interface WatchFile {
  path: string;
  regex: string;
  skip: string;
  poll: boolean;
  pipe: boolean;
  mustExist: boolean;
  logMatch: boolean;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/endpoints/epconfig.Endpoint>
 */
export interface Endpoint extends CronJob {
  query?: Record<string, null | string[]>;
  header?: Record<string, null | string[]>;
  template: string;
  name: string;
  url: string;
  method: string;
  body: string;
  follow: boolean;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/common/scheduler.CronJob>
 */
export interface CronJob {
  frequency: number;
  interval: number;
  atTimes?: (null | number[])[];
  daysOfWeek?: Weekday[];
  daysOfMonth?: number[];
  months?: number[];
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/commands.Command>
 */
export interface Command extends CmdconfigConfig {};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/commands/cmdconfig.Config>
 */
export interface CmdconfigConfig {
  name: string;
  hash: string;
  shell: boolean;
  log: boolean;
  notify: boolean;
  args: number;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/logs.LogConfig>
 */
export interface LogConfig {
  logFile: string;
  debugLog: string;
  httpLog: string;
  logFiles: number;
  logFileMb: number;
  fileMode: number;
  debug: boolean;
  quiet: boolean;
  noUploads: boolean;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.Apps>
 */
export interface Apps {
  apiKey: string;
  extraKeys?: string[];
  urlbase: string;
  maxBody: number;
  serial: boolean;
  sonarr?: (null | SonarrConfig)[];
  radarr?: (null | RadarrConfig)[];
  lidarr?: (null | LidarrConfig)[];
  readarr?: (null | ReadarrConfig)[];
  prowlarr?: (null | ProwlarrConfig)[];
  deluge?: (null | DelugeConfig)[];
  qbit?: (null | QbitConfig)[];
  rtorrent?: (null | RtorrentConfig)[];
  sabnzbd?: (null | SabNZBConfig)[];
  nzbget?: (null | NZBGetConfig)[];
  transmission?: (null | XmissionConfig)[];
  tautulli?: TautulliConfig;
  plex?: PlexConfig;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.SonarrConfig>
 */
export interface SonarrConfig extends ExtraConfig {
  Config?: StarrConfig;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.ExtraConfig>
 */
export interface ExtraConfig {
  name: string;
  timeout: string;
  interval: string;
  validSsl: boolean;
  deletes: number;
};

/**
 * Config is the data needed to poll Radarr or Sonarr or Lidarr or Readarr.
 * At a minimum, provide a URL and API Key.
 * HTTPUser and HTTPPass are used for Basic HTTP auth, if enabled (not common).
 * Username and Password are for non-API paths with native authentication enabled.
 * @see golang: <golift.io/starr.Config>
 */
export interface StarrConfig {
  apiKey: string;
  url: string;
  httpPass: string;
  httpUser: string;
  username: string;
  password: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.RadarrConfig>
 */
export interface RadarrConfig extends ExtraConfig {
  Config?: StarrConfig;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.LidarrConfig>
 */
export interface LidarrConfig extends ExtraConfig {
  Config?: StarrConfig;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.ReadarrConfig>
 */
export interface ReadarrConfig extends ExtraConfig {
  Config?: StarrConfig;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.ProwlarrConfig>
 */
export interface ProwlarrConfig extends ExtraConfig {
  Config?: StarrConfig;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.DelugeConfig>
 */
export interface DelugeConfig extends ExtraConfig {
  Config?: DelugeConfig0;
};

/**
 * Config is the data needed to poll Deluge.
 * @see golang: <golift.io/deluge.Config>
 */
export interface DelugeConfig0 {
  url: string;
  password: string;
  httppass: string;
  httpuser: string;
  version: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.QbitConfig>
 */
export interface QbitConfig extends ExtraConfig {
  Config?: QbitConfig0;
};

/**
 * Config is the input data needed to return a Qbit struct.
 * This is setup to allow you to easily pass this data in from a config file.
 * @see golang: <golift.io/qbit.Config>
 */
export interface QbitConfig0 {
  url: string;
  user: string;
  pass: string;
  httppass: string;
  httpuser: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.RtorrentConfig>
 */
export interface RtorrentConfig extends ExtraConfig {
  url: string;
  user: string;
  pass: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.SabNZBConfig>
 */
export interface SabNZBConfig extends ExtraConfig {
  Config?: SabnzbdConfig;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/sabnzbd.Config>
 */
export interface SabnzbdConfig {
  url: string;
  apiKey: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.NZBGetConfig>
 */
export interface NZBGetConfig extends ExtraConfig {
  Config?: NzbgetConfig;
};

/**
 * Config is the input data needed to return a NZBGet struct.
 * This is setup to allow you to easily pass this data in from a config file.
 * @see golang: <golift.io/nzbget.Config>
 */
export interface NzbgetConfig {
  url: string;
  user: string;
  pass: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.XmissionConfig>
 */
export interface XmissionConfig extends ExtraConfig {
  url: string;
  user: string;
  pass: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.TautulliConfig>
 */
export interface TautulliConfig extends ExtraConfig, TautulliConfig0 {};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli.Config>
 */
export interface TautulliConfig0 {
  url: string;
  apiKey: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.PlexConfig>
 */
export interface PlexConfig extends ExtraConfig {
  Config?: PlexConfig0;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex.Config>
 */
export interface PlexConfig0 {
  url: string;
  token: string;
};

// Packages parsed:
//   1. github.com/Notifiarr/notifiarr/pkg/apps
//   2. github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex
//   3. github.com/Notifiarr/notifiarr/pkg/apps/apppkg/sabnzbd
//   4. github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli
//   5. github.com/Notifiarr/notifiarr/pkg/configfile
//   6. github.com/Notifiarr/notifiarr/pkg/logs
//   7. github.com/Notifiarr/notifiarr/pkg/services
//   8. github.com/Notifiarr/notifiarr/pkg/snapshot
//   9. github.com/Notifiarr/notifiarr/pkg/triggers/commands
//  10. github.com/Notifiarr/notifiarr/pkg/triggers/commands/cmdconfig
//  11. github.com/Notifiarr/notifiarr/pkg/triggers/common/scheduler
//  12. github.com/Notifiarr/notifiarr/pkg/triggers/endpoints/epconfig
//  13. github.com/Notifiarr/notifiarr/pkg/triggers/filewatch
//  14. golift.io/deluge
//  15. golift.io/nzbget
//  16. golift.io/qbit
//  17. golift.io/starr
